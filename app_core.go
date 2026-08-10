// Package main implements the Send-Log-TCP desktop application: a Wails-based
// GUI for composing and sending syslog messages over TCP, UDP and TLS.
package main

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Version information - injected at build time via ldflags
// Use: go build -ldflags "-X main.Version=1.4.2"
var Version = "dev"

// ================================================================================
// APP CORE - Main application struct with lifecycle management
// Follows Single Responsibility Principle: handles only app lifecycle and metadata
// ================================================================================

// App struct is the main application controller
// It manages lifecycle hooks and coordinates between services
type App struct {
	ctx context.Context

	// Service references for coordination
	connectionService *ConnectionService
	syslogService     *SyslogService
	stressTestService *StressTestService
}

// NewApp creates a new App instance
func NewApp() *App {
	return &App{}
}

// SetServices injects service dependencies (called from main.go)
func (a *App) SetServices(conn *ConnectionService, syslog *SyslogService, stress *StressTestService) {
	a.connectionService = conn
	a.syslogService = syslog
	a.stressTestService = stress
}

// startup is called at application startup
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// domReady is called after front-end resources have been loaded
func (a *App) domReady(_ context.Context) {
	// Center the window on the screen
	runtime.WindowCenter(a.ctx)
}

// beforeClose is called when the application is about to quit
func (a *App) beforeClose(_ context.Context) (prevent bool) {
	// Check if connection service has active connection
	if a.connectionService != nil && a.connectionService.IsConnected() {
		result, err := runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
			Type:          runtime.QuestionDialog,
			Title:         "Confirm Exit",
			Message:       "There is an active connection. Are you sure you want to exit?",
			Buttons:       []string{"Yes", "No"},
			DefaultButton: "No",
			CancelButton:  "No",
		})
		if err != nil || result != "Yes" {
			return true // Prevent close
		}
	}

	// Clean up connections
	if a.connectionService != nil {
		a.connectionService.Disconnect()
	}

	// Stop stress test if running
	if a.stressTestService != nil {
		a.stressTestService.StopContinuousSend()
	}

	return false
}

// shutdown is called at application termination
func (a *App) shutdown(_ context.Context) {
	// Ensure all connections are closed
	if a.connectionService != nil {
		a.connectionService.Disconnect()
	}
}

// ================================================================================
// APP METADATA API
// ================================================================================

// GetVersion returns the application version
// The version is injected at build time via ldflags
func (a *App) GetVersion() string {
	return Version
}

// OpenURL opens a URL in the default system browser
func (a *App) OpenURL(url string) {
	runtime.BrowserOpenURL(a.ctx, url)
	runtime.LogDebug(a.ctx, "Opened URL in browser: "+url)
}

// OpenGitHub opens the project's GitHub repository in the default browser
func (a *App) OpenGitHub() {
	runtime.BrowserOpenURL(a.ctx, "https://github.com/yllada/Send-Log-TCP")
}
