package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// ================================================================================
// STRESS TEST SERVICE - Handles continuous/stress test message sending
// Follows Single Responsibility Principle: manages only stress test operations
// ================================================================================

// ContinuousSendConfig configuration for continuous/stress test mode
type ContinuousSendConfig struct {
	// Connection settings
	Address       string        `json:"Address"`
	Port          string        `json:"Port"`
	Protocol      string        `json:"Protocol"`
	Message       string        `json:"Message"`
	FramingMethod FramingMethod `json:"FramingMethod"`
	Facility      uint8         `json:"Facility"`
	Severity      uint8         `json:"Severity"`
	Hostname      string        `json:"Hostname"`
	Appname       string        `json:"Appname"`
	MessageFormat MessageFormat `json:"MessageFormat"`
	// UseRFC5424 is the legacy format switch, kept for stored configs.
	// MessageFormat takes precedence when set.
	UseRFC5424     bool   `json:"UseRFC5424"`
	UseTLS         bool   `json:"UseTLS"`
	TLSVerify      bool   `json:"TLSVerify"`
	CACertPath     string `json:"CACertPath"`
	ClientCertPath string `json:"ClientCertPath"`
	ClientKeyPath  string `json:"ClientKeyPath"`

	// Continuous send settings
	Duration       int  `json:"Duration"`
	MessagesPerSec int  `json:"MessagesPerSec"`
	MaxMessages    int  `json:"MaxMessages"`
	RandomizeData  bool `json:"RandomizeData"`
}

// ContinuousStats real-time statistics for continuous send
type ContinuousStats struct {
	TotalSent      int64   `json:"totalSent"`
	TotalErrors    int64   `json:"totalErrors"`
	CurrentRate    float64 `json:"currentRate"`
	ElapsedSeconds float64 `json:"elapsedSeconds"`
	IsRunning      bool    `json:"isRunning"`
	StartTime      int64   `json:"startTime"`
	TargetRate     int     `json:"targetRate"`
	Duration       int     `json:"duration"`
}

// StressTestService handles continuous/stress test operations
type StressTestService struct {
	ctx              context.Context
	continuousCancel context.CancelFunc
	continuousMu     sync.RWMutex
	continuousStats  ContinuousStats
	isRunning        bool
}

// NewStressTestService creates a new StressTestService instance
func NewStressTestService() *StressTestService {
	return &StressTestService{}
}

// SetContext sets the Wails runtime context
func (s *StressTestService) SetContext(ctx context.Context) {
	s.ctx = ctx
}

// IsContinuousRunning returns whether continuous send is currently active
func (s *StressTestService) IsContinuousRunning() bool {
	s.continuousMu.RLock()
	defer s.continuousMu.RUnlock()
	return s.isRunning
}

// GetContinuousStats returns current statistics for continuous send
func (s *StressTestService) GetContinuousStats() ContinuousStats {
	s.continuousMu.RLock()
	defer s.continuousMu.RUnlock()
	return s.continuousStats
}

// StartContinuousSend starts sending messages continuously based on config
func (s *StressTestService) StartContinuousSend(config ContinuousSendConfig) error {
	s.continuousMu.Lock()
	if s.isRunning {
		s.continuousMu.Unlock()
		return fmt.Errorf("continuous send already running")
	}

	// Create cancellation context
	ctx, cancel := context.WithCancel(s.ctx)
	s.continuousCancel = cancel
	s.isRunning = true
	s.continuousStats = ContinuousStats{
		TotalSent:   0,
		TotalErrors: 0,
		CurrentRate: 0,
		IsRunning:   true,
		StartTime:   time.Now().Unix(),
		TargetRate:  config.MessagesPerSec,
		Duration:    config.Duration,
	}
	s.continuousMu.Unlock()

	// Emit started event
	runtime.EventsEmit(s.ctx, "continuous:started", s.continuousStats)

	// Start the background goroutine
	go s.runContinuousSend(ctx, config)

	return nil
}

// StopContinuousSend stops the continuous send operation
func (s *StressTestService) StopContinuousSend() {
	s.continuousMu.Lock()
	defer s.continuousMu.Unlock()

	if s.continuousCancel != nil {
		s.continuousCancel()
		s.continuousCancel = nil
	}
	s.isRunning = false
	s.continuousStats.IsRunning = false

	// Emit stopped event
	runtime.EventsEmit(s.ctx, "continuous:stopped", s.continuousStats)
}

// runContinuousSend is the goroutine that handles continuous message sending
func (s *StressTestService) runContinuousSend(ctx context.Context, config ContinuousSendConfig) {
	defer func() {
		s.continuousMu.Lock()
		s.isRunning = false
		s.continuousStats.IsRunning = false
		s.continuousMu.Unlock()
		runtime.EventsEmit(s.ctx, "continuous:stopped", s.GetContinuousStats())
	}()

	// Build syslog config
	syslogConfig := SyslogConfig{
		Address:        config.Address,
		Port:           config.Port,
		Protocol:       config.Protocol,
		FramingMethod:  config.FramingMethod,
		Facility:       config.Facility,
		Severity:       config.Severity,
		Hostname:       config.Hostname,
		Appname:        config.Appname,
		MessageFormat:  config.MessageFormat,
		UseRFC5424:     config.UseRFC5424,
		UseTLS:         config.UseTLS,
		TLSVerify:      config.TLSVerify,
		CACertPath:     config.CACertPath,
		ClientCertPath: config.ClientCertPath,
		ClientKeyPath:  config.ClientKeyPath,
	}

	// Validate config
	if err := validateConfig(&syslogConfig); err != nil {
		runtime.LogError(s.ctx, fmt.Sprintf("Invalid config: %v", err))
		runtime.EventsEmit(s.ctx, "continuous:error", err.Error())
		return
	}

	// Establish connection
	conn, err := dialConnection(
		config.Address, config.Port, config.Protocol,
		config.UseTLS, config.TLSVerify,
		config.CACertPath, config.ClientCertPath, config.ClientKeyPath,
	)
	if err != nil {
		runtime.LogError(s.ctx, fmt.Sprintf("Connection failed: %v", err))
		runtime.EventsEmit(s.ctx, "continuous:error", err.Error())
		return
	}
	defer func() { _ = conn.Close() }()

	// Create framer for TCP
	var framer *Framer
	if config.Protocol == "tcp" {
		framer = NewFramer(FramingConfig{
			Method:           config.FramingMethod,
			ValidateUTF8:     true,
			MaxMessageLength: 0,
		})
	}

	// Rate limiting with ticker
	interval := time.Second / time.Duration(config.MessagesPerSec)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Duration timer if configured
	var durationTimer <-chan time.Time
	if config.Duration > 0 {
		durationTimer = time.After(time.Duration(config.Duration) * time.Second)
	}

	// Stats tracking
	startTime := time.Now()
	var sentCount, errorCount int64
	var sequenceNum int64
	lastStatsUpdate := time.Now()

	// Main loop
	for {
		select {
		case <-ctx.Done():
			return

		case <-durationTimer:
			runtime.LogInfo(s.ctx, "Continuous send duration completed")
			return

		case <-ticker.C:
			// Check max messages limit
			if config.MaxMessages > 0 && sentCount >= int64(config.MaxMessages) {
				runtime.LogInfo(s.ctx, fmt.Sprintf("Max messages limit reached: %d", config.MaxMessages))
				return
			}

			// Build message with optional randomization
			message := config.Message
			if config.RandomizeData {
				sequenceNum++
				message = fmt.Sprintf("[seq=%d ts=%s] %s",
					sequenceNum,
					time.Now().Format("15:04:05.000"),
					config.Message,
				)
			}

			// Build syslog message
			syslogMsg, err := buildSyslogMessage(syslogConfig, message)
			if err != nil {
				errorCount++
				continue
			}

			// Send message
			var sendErr error
			if config.Protocol == "tcp" && framer != nil {
				framedMsg, err := framer.Frame(syslogMsg)
				if err != nil {
					errorCount++
					continue
				}
				if err := conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
					runtime.LogWarning(s.ctx, fmt.Sprintf("Failed to set write deadline: %v", err))
				}
				sendErr = writeAll(conn, framedMsg)
			} else {
				if err := conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
					runtime.LogWarning(s.ctx, fmt.Sprintf("Failed to set write deadline: %v", err))
				}
				sendErr = writeAll(conn, []byte(syslogMsg))
			}

			if sendErr != nil {
				errorCount++
				runtime.LogWarning(s.ctx, fmt.Sprintf("Send error: %v", sendErr))
			} else {
				sentCount++
			}

			// Update stats every 250ms
			if time.Since(lastStatsUpdate) >= 250*time.Millisecond {
				elapsed := time.Since(startTime).Seconds()
				currentRate := float64(sentCount) / elapsed

				s.continuousMu.Lock()
				s.continuousStats.TotalSent = sentCount
				s.continuousStats.TotalErrors = errorCount
				s.continuousStats.ElapsedSeconds = elapsed
				s.continuousStats.CurrentRate = currentRate
				stats := s.continuousStats
				s.continuousMu.Unlock()

				runtime.EventsEmit(s.ctx, "continuous:stats", stats)
				lastStatsUpdate = time.Now()
			}
		}
	}
}
