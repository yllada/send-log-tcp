package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Connection timeouts according to best practices
const (
	connectionTimeout = 10 * time.Second
	writeTimeout      = 30 * time.Second
	readTimeout       = 30 * time.Second
)

// ================================================================================
// CONNECTION SERVICE - Handles network connections (TCP, UDP, TLS)
// Follows Single Responsibility Principle: manages only connection state
// ================================================================================

// ConnectionService manages network connections to syslog servers
type ConnectionService struct {
	ctx         context.Context
	conn        net.Conn
	connMu      sync.Mutex
	isConnected bool
}

// NewConnectionService creates a new ConnectionService instance
func NewConnectionService() *ConnectionService {
	return &ConnectionService{}
}

// SetContext sets the Wails runtime context
func (c *ConnectionService) SetContext(ctx context.Context) {
	c.ctx = ctx
}

// IsConnected returns the current connection state
func (c *ConnectionService) IsConnected() bool {
	c.connMu.Lock()
	defer c.connMu.Unlock()
	return c.isConnected
}

// GetConnection returns the current connection (for use by other services)
func (c *ConnectionService) GetConnection() net.Conn {
	c.connMu.Lock()
	defer c.connMu.Unlock()
	return c.conn
}

// CheckConnection verifies if a connection can be established
// Supports TCP, UDP, and TCP+TLS (RFC 5425 for secure syslog)
func (c *ConnectionService) CheckConnection(address string, port string, protocol string, useTLS bool, tlsVerify bool, caCertPath string, clientCertPath string, clientKeyPath string) (bool, error) {
	fullAddress := net.JoinHostPort(address, port)

	conn, err := dialConnection(address, port, protocol, useTLS, tlsVerify, caCertPath, clientCertPath, clientKeyPath)
	if err != nil {
		runtime.LogError(c.ctx, fmt.Sprintf("Error connecting to %s: %v", fullAddress, err))
		return false, err
	}

	// Store the active connection
	c.connMu.Lock()
	c.conn = conn
	c.isConnected = true
	c.connMu.Unlock()

	protocolInfo := protocol
	if useTLS {
		protocolInfo = fmt.Sprintf("%s+TLS", protocol)
	}
	runtime.LogInfo(c.ctx, fmt.Sprintf("Connected to %s via %s", fullAddress, protocolInfo))
	return true, nil
}

// Disconnect closes the active connection if it exists
func (c *ConnectionService) Disconnect() {
	c.connMu.Lock()
	defer c.connMu.Unlock()

	if c.isConnected {
		err := c.conn.Close()
		if err != nil {
			log.Printf("Error disconnecting: %v", err)
		} else {
			log.Println("Disconnected successfully.")
		}
		c.isConnected = false
	}
}

// SelectCACertificate opens a dialog to select a CA certificate file
func (c *ConnectionService) SelectCACertificate() (string, error) {
	filePath, err := runtime.OpenFileDialog(c.ctx, runtime.OpenDialogOptions{
		Title: "Select CA Certificate",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Certificate Files (*.pem, *.crt, *.cer)",
				Pattern:     "*.pem;*.crt;*.cer;*.ca-bundle",
			},
			{
				DisplayName: "All Files (*.*)",
				Pattern:     "*.*",
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to open file dialog: %w", err)
	}
	return filePath, nil
}

// SelectClientCertificate opens a dialog to select a client certificate (mTLS)
func (c *ConnectionService) SelectClientCertificate() (string, error) {
	filePath, err := runtime.OpenFileDialog(c.ctx, runtime.OpenDialogOptions{
		Title: "Select Client Certificate",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Certificate Files (*.pem, *.crt, *.cer)",
				Pattern:     "*.pem;*.crt;*.cer",
			},
			{
				DisplayName: "All Files (*.*)",
				Pattern:     "*.*",
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to open file dialog: %w", err)
	}
	return filePath, nil
}

// SelectClientKey opens a dialog to select a client private key (mTLS)
func (c *ConnectionService) SelectClientKey() (string, error) {
	filePath, err := runtime.OpenFileDialog(c.ctx, runtime.OpenDialogOptions{
		Title: "Select Client Private Key",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Key Files (*.pem, *.key)",
				Pattern:     "*.pem;*.key",
			},
			{
				DisplayName: "All Files (*.*)",
				Pattern:     "*.*",
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to open file dialog: %w", err)
	}
	return filePath, nil
}

// ================================================================================
// INTERNAL CONNECTION HELPERS (not exported to frontend)
// ================================================================================

// dialTLS establishes a TLS connection to the remote server
func dialTLS(address, port string, tlsVerify bool, caCertPath, clientCertPath, clientKeyPath string) (net.Conn, error) {
	fullAddress := net.JoinHostPort(address, port)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: !tlsVerify,
		MinVersion:         tls.VersionTLS12,
		MaxVersion:         tls.VersionTLS13,
	}

	if tlsVerify {
		tlsConfig.ServerName = address

		if caCertPath != "" {
			rootCAs, err := loadCACertificate(caCertPath)
			if err != nil {
				return nil, fmt.Errorf("failed to load CA certificate: %w", err)
			}
			tlsConfig.RootCAs = rootCAs
			log.Printf("Using custom CA certificate: %s", caCertPath)
		}
	}

	if clientCertPath != "" && clientKeyPath != "" {
		clientCert, err := loadClientCertificate(clientCertPath, clientKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{clientCert}
		log.Printf("Using client certificate for mTLS: %s", clientCertPath)
	} else if clientCertPath != "" || clientKeyPath != "" {
		return nil, fmt.Errorf("both client certificate and key are required for mTLS")
	}

	dialer := &net.Dialer{
		Timeout: connectionTimeout,
	}

	tlsConn, err := tls.DialWithDialer(dialer, "tcp", fullAddress, tlsConfig)
	if err != nil {
		return nil, fmt.Errorf("TLS connection failed: %w", err)
	}

	if err := tlsConn.Handshake(); err != nil {
		_ = tlsConn.Close()
		return nil, fmt.Errorf("TLS handshake failed: %w", err)
	}

	state := tlsConn.ConnectionState()
	log.Printf("TLS connection established: version=%s, cipher=%s, server=%s",
		tlsVersionName(state.Version),
		tls.CipherSuiteName(state.CipherSuite),
		state.ServerName,
	)

	return tlsConn, nil
}

// tlsVersionName converts TLS version code to readable name
func tlsVersionName(version uint16) string {
	switch version {
	case tls.VersionTLS10:
		return "TLS 1.0"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	default:
		return fmt.Sprintf("Unknown (0x%04X)", version)
	}
}

// loadCACertificate loads a CA certificate from PEM file
func loadCACertificate(certPath string) (*x509.CertPool, error) {
	caCert, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate file: %w", err)
	}

	rootCAs := x509.NewCertPool()
	if ok := rootCAs.AppendCertsFromPEM(caCert); !ok {
		return nil, fmt.Errorf("failed to parse CA certificate (invalid PEM format)")
	}

	return rootCAs, nil
}

// loadClientCertificate loads client certificate and private key for mTLS
func loadClientCertificate(certPath, keyPath string) (tls.Certificate, error) {
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to load client certificate/key pair: %w", err)
	}
	return cert, nil
}

// dialConnection establishes a connection based on protocol configuration
func dialConnection(address, port, protocol string, useTLS, tlsVerify bool, caCertPath, clientCertPath, clientKeyPath string) (net.Conn, error) {
	if useTLS && protocol == "udp" {
		return nil, fmt.Errorf("TLS is not supported over UDP protocol")
	}

	if protocol == "tcp" && useTLS {
		return dialTLS(address, port, tlsVerify, caCertPath, clientCertPath, clientKeyPath)
	}

	fullAddress := net.JoinHostPort(address, port)
	dialer := &net.Dialer{
		Timeout: connectionTimeout,
	}

	conn, err := dialer.Dial(protocol, fullAddress)
	if err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}

	return conn, nil
}
