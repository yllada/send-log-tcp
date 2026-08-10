package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// ================================================================================
// SYSLOG SERVICE - Handles syslog message formatting and sending
// Follows Single Responsibility Principle: manages message formatting and transmission
// ================================================================================

// FramingMethod specifies the framing method for TCP per RFC 6587
type FramingMethod string

const (
	// OctetCounting implements octet counting method (RFC 6587 Section 3.4.1)
	OctetCounting FramingMethod = "octet-counting"

	// NonTransparent implements non-transparent framing (RFC 6587 Section 3.4.2)
	NonTransparent FramingMethod = "non-transparent"
)

// MessageFormat specifies how the syslog message is assembled before framing
type MessageFormat string

const (
	// MessageFormatRFC5424 builds a full RFC 5424 header (version, timestamp,
	// hostname, app-name, nil procid/msgid/structured-data) around the message
	MessageFormatRFC5424 MessageFormat = "rfc5424"

	// MessageFormatRFC3164 builds a legacy BSD syslog header (RFC 3164) around
	// the message
	MessageFormatRFC3164 MessageFormat = "rfc3164"

	// MessageFormatRawPRI emits only "<PRI>" followed by the message bytes
	// verbatim: no version digit, no space, no timestamp, no hostname, no
	// app-name, no BOM and no trailing newline.
	//
	// This exists to replay real vendor logs. A captured line (for example an
	// ESET ESMC event) already carries its own complete RFC 5424 header, so
	// injecting another one would corrupt the event. The operator supplies
	// everything after the PRI.
	MessageFormatRawPRI MessageFormat = "raw-pri"
)

// IsValidMessageFormat reports whether the message format is recognized
func IsValidMessageFormat(format MessageFormat) bool {
	switch format {
	case MessageFormatRFC5424, MessageFormatRFC3164, MessageFormatRawPRI:
		return true
	default:
		return false
	}
}

// resolveMessageFormat determines the effective message format.
//
// A recognized, non-empty format always wins. When the format is absent (or not
// recognized) it falls back to the legacy UseRFC5424 boolean, which keeps
// templates stored before MessageFormat existed working without any storage
// migration.
func resolveMessageFormat(format MessageFormat, useRFC5424 bool) MessageFormat {
	if IsValidMessageFormat(format) {
		return format
	}
	if useRFC5424 {
		return MessageFormatRFC5424
	}
	return MessageFormatRFC3164
}

// SyslogConfig holds the configuration for sending syslog messages
type SyslogConfig struct {
	Address       string        `json:"Address"`
	Port          string        `json:"Port"`
	Protocol      string        `json:"Protocol"`
	Messages      []string      `json:"Messages"`
	FramingMethod FramingMethod `json:"FramingMethod"`
	Facility      uint8         `json:"Facility"`
	Severity      uint8         `json:"Severity"`
	Hostname      string        `json:"Hostname"`
	Appname       string        `json:"Appname"`
	MessageFormat MessageFormat `json:"MessageFormat"`
	// UseRFC5424 is the legacy format switch. It is kept because stored
	// templates and profiles persist it; MessageFormat takes precedence when set.
	UseRFC5424     bool   `json:"UseRFC5424"`
	UseTLS         bool   `json:"UseTLS"`
	TLSVerify      bool   `json:"TLSVerify"`
	CACertPath     string `json:"CACertPath"`
	ClientCertPath string `json:"ClientCertPath"`
	ClientKeyPath  string `json:"ClientKeyPath"`
}

// SyslogResponse contains the result of send operations
type SyslogResponse struct {
	SentMessages []string `json:"sentMessages"`
	Errors       []string `json:"errors"`
}

// SyslogService handles syslog message sending operations
type SyslogService struct {
	ctx context.Context
}

// NewSyslogService creates a new SyslogService instance
func NewSyslogService() *SyslogService {
	return &SyslogService{}
}

// SetContext sets the Wails runtime context
func (s *SyslogService) SetContext(ctx context.Context) {
	s.ctx = ctx
}

// SendSyslogMessages sends syslog messages and returns the result
func (s *SyslogService) SendSyslogMessages(config SyslogConfig) SyslogResponse {
	response := SyslogResponse{
		SentMessages: []string{},
		Errors:       []string{},
	}

	// Validate configuration
	if err := validateConfig(&config); err != nil {
		response.Errors = append(response.Errors, fmt.Sprintf("Invalid configuration: %v", err))
		return response
	}

	// Emit start event
	runtime.EventsEmit(s.ctx, "syslog:sending", map[string]interface{}{
		"total": len(config.Messages),
	})

	fullAddress := net.JoinHostPort(config.Address, config.Port)

	// Establish connection
	conn, err := dialConnection(config.Address, config.Port, config.Protocol, config.UseTLS, config.TLSVerify, config.CACertPath, config.ClientCertPath, config.ClientKeyPath)
	if err != nil {
		response.Errors = append(response.Errors, fmt.Sprintf("Error connecting to %s: %v", fullAddress, err))
		return response
	}
	defer conn.Close()

	// Send messages based on protocol
	if config.Protocol == "tcp" {
		return s.sendTCPMessages(conn, config)
	}
	return s.sendUDPMessages(conn, config)
}

// sendTCPMessages sends TCP messages with proper framing per RFC 6587
func (s *SyslogService) sendTCPMessages(conn net.Conn, config SyslogConfig) SyslogResponse {
	response := SyslogResponse{
		SentMessages: []string{},
		Errors:       []string{},
	}

	// Create framer with appropriate configuration
	framer := NewFramer(FramingConfig{
		Method:           config.FramingMethod,
		ValidateUTF8:     true,
		MaxMessageLength: 0,
	})

	totalMessages := len(config.Messages)
	for i, message := range config.Messages {
		// Build syslog message
		syslogMsg, err := buildSyslogMessage(config, message)
		if err != nil {
			response.Errors = append(response.Errors, fmt.Sprintf("Error building message: %v", err))
			continue
		}

		// Apply TCP framing
		framedMsg, err := framer.Frame(syslogMsg)
		if err != nil {
			response.Errors = append(response.Errors, fmt.Sprintf("Error framing message: %v", err))
			continue
		}

		// Set write deadline
		if err := conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
			runtime.LogWarning(s.ctx, fmt.Sprintf("Failed to set write deadline: %v", err))
		}

		// Send with robust handling
		if err := writeAll(conn, framedMsg); err != nil {
			runtime.LogError(s.ctx, fmt.Sprintf("Error sending message: %v", err))
			response.Errors = append(response.Errors, fmt.Sprintf("Error sending message: %v", err))
		} else {
			runtime.LogDebug(s.ctx, fmt.Sprintf("Sent message: %s", syslogMsg))
			response.SentMessages = append(response.SentMessages, syslogMsg)
		}

		// Emit progress event
		runtime.EventsEmit(s.ctx, "syslog:progress", map[string]interface{}{
			"current": i + 1,
			"total":   totalMessages,
			"percent": float64(i+1) / float64(totalMessages) * 100,
		})
	}

	// Emit completion event
	runtime.EventsEmit(s.ctx, "syslog:complete", map[string]interface{}{
		"sent":   len(response.SentMessages),
		"errors": len(response.Errors),
	})

	return response
}

// sendUDPMessages sends UDP messages (one message per packet)
func (s *SyslogService) sendUDPMessages(conn net.Conn, config SyslogConfig) SyslogResponse {
	response := SyslogResponse{
		SentMessages: []string{},
		Errors:       []string{},
	}

	totalMessages := len(config.Messages)
	for i, message := range config.Messages {
		syslogMsg, err := buildSyslogMessage(config, message)
		if err != nil {
			response.Errors = append(response.Errors, fmt.Sprintf("Error building message: %v", err))
			continue
		}

		if err := conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
			runtime.LogWarning(s.ctx, fmt.Sprintf("Failed to set write deadline: %v", err))
		}

		if err := writeAll(conn, []byte(syslogMsg)); err != nil {
			runtime.LogError(s.ctx, fmt.Sprintf("Error sending message: %v", err))
			response.Errors = append(response.Errors, fmt.Sprintf("Error sending message: %v", err))
		} else {
			runtime.LogDebug(s.ctx, fmt.Sprintf("Sent message: %s", syslogMsg))
			response.SentMessages = append(response.SentMessages, syslogMsg)
		}

		runtime.EventsEmit(s.ctx, "syslog:progress", map[string]interface{}{
			"current": i + 1,
			"total":   totalMessages,
			"percent": float64(i+1) / float64(totalMessages) * 100,
		})
	}

	runtime.EventsEmit(s.ctx, "syslog:complete", map[string]interface{}{
		"sent":   len(response.SentMessages),
		"errors": len(response.Errors),
	})

	return response
}

// ================================================================================
// SYSLOG MESSAGE FORMATTING HELPERS
// ================================================================================

// buildSyslogMessage constructs a syslog message per RFC 5424, RFC 3164 or the
// raw-pri passthrough mode
func buildSyslogMessage(config SyslogConfig, message string) (string, error) {
	priority := config.Facility*8 + config.Severity

	if priority > 191 {
		return "", fmt.Errorf("invalid priority %d (facility=%d, severity=%d)", priority, config.Facility, config.Severity)
	}

	switch resolveMessageFormat(config.MessageFormat, config.UseRFC5424) {
	case MessageFormatRawPRI:
		return buildRawPRIMessage(priority, message), nil
	case MessageFormatRFC3164:
		return buildRFC3164Message(priority, config, message), nil
	default:
		return buildRFC5424Message(priority, config, message), nil
	}
}

// buildRawPRIMessage prepends only the priority and emits the message bytes
// verbatim.
//
// Exactly one rule applies: "<PRI>" followed by the message. No version digit,
// no separating space, no timestamp, no hostname, no app-name, no BOM and no
// trailing newline. Hostname and Appname from the config are deliberately
// ignored because the replayed payload already supplies them.
func buildRawPRIMessage(priority uint8, message string) string {
	return fmt.Sprintf("<%d>%s", priority, message)
}

// buildRFC5424Message constructs message per RFC 5424
func buildRFC5424Message(priority uint8, config SyslogConfig, message string) string {
	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05.000Z07:00")

	hostname := config.Hostname
	if hostname == "" {
		hostname = "-"
	}

	appname := config.Appname
	if appname == "" {
		appname = "-"
	}

	return fmt.Sprintf("<%d>1 %s %s %s - - - %s",
		priority, timestamp, hostname, appname, message)
}

// buildRFC3164Message constructs message per RFC 3164 (BSD syslog)
func buildRFC3164Message(priority uint8, config SyslogConfig, message string) string {
	timestamp := time.Now().Format("Jan  2 15:04:05")

	hostname := config.Hostname
	if hostname == "" {
		hostname, _ = os.Hostname()
		if hostname == "" {
			hostname = "localhost"
		}
	}

	appname := config.Appname
	if appname == "" {
		appname = "app"
	}

	return fmt.Sprintf("<%d>%s %s %s: %s",
		priority, timestamp, hostname, appname, message)
}

// writeAll ensures all bytes are written (handles partial writes)
func writeAll(w io.Writer, data []byte) error {
	totalWritten := 0
	dataLen := len(data)

	for totalWritten < dataLen {
		n, err := w.Write(data[totalWritten:])
		if err != nil {
			return fmt.Errorf("write failed after %d/%d bytes: %w", totalWritten, dataLen, err)
		}
		totalWritten += n
	}

	return nil
}

// validateConfig validates and normalizes configuration
func validateConfig(config *SyslogConfig) error {
	if config.Address == "" {
		return fmt.Errorf("address is required")
	}
	if config.Port == "" {
		return fmt.Errorf("port is required")
	}

	if config.Facility > 23 {
		return fmt.Errorf("facility must be 0-23 (got %d)", config.Facility)
	}
	if config.Severity > 7 {
		return fmt.Errorf("severity must be 0-7 (got %d)", config.Severity)
	}

	if config.FramingMethod == "" {
		if config.Protocol == "tcp" {
			config.FramingMethod = RecommendedFramingMethod()
		}
	}

	if config.FramingMethod != "" && !IsValidFramingMethod(config.FramingMethod) {
		return fmt.Errorf("invalid framing method '%s'", config.FramingMethod)
	}

	// An empty MessageFormat is valid: legacy configs fall back to UseRFC5424.
	if config.MessageFormat != "" && !IsValidMessageFormat(config.MessageFormat) {
		return fmt.Errorf("invalid message format '%s' (expected '%s', '%s' or '%s')",
			config.MessageFormat, MessageFormatRFC5424, MessageFormatRFC3164, MessageFormatRawPRI)
	}

	if config.Hostname == "" {
		hostname, err := os.Hostname()
		if err == nil && hostname != "" {
			config.Hostname = hostname
		} else {
			config.Hostname = "localhost"
		}
	}

	if config.Appname == "" {
		config.Appname = "sendlog"
	}

	return nil
}
