package main

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx     context.Context
	watcher *WindowWatcher
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// OnStartup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) OnStartup(ctx context.Context) {
	a.ctx = ctx
}

// GetCurrentWindow returns the current active window information
func (a *App) GetCurrentWindow() (*WindowInfo, error) {
	if a.watcher == nil {
		// Return nil instead of empty struct to indicate no watcher
		return nil, nil
	}
	info, err := a.watcher.GetActiveWindow()
	if err != nil {
		return nil, err
	}
	// Ensure we always return valid data
	if info == nil {
		return nil, nil
	}
	return info, nil
}

// OnWindowChanged is called when the active window changes
// This is used by the frontend to listen for window changes
func (a *App) OnWindowChanged(fn func(*WindowInfo)) {
	// This will be handled by the watcher emitting events
	// The frontend will use runtime.EventsOn to listen
}

// StartMonitoring starts the window monitoring
func (a *App) StartMonitoring() error {
	if a.watcher == nil {
		return nil
	}
	return a.watcher.StartMonitoring()
}

// StopMonitoring stops the window monitoring
func (a *App) StopMonitoring() {
	if a.watcher != nil {
		a.watcher.StopMonitoring()
	}
}

// EnableAutoStart enables auto-start on Windows boot
func (a *App) EnableAutoStart() error {
	exePath, err := getExecutablePath()
	if err != nil {
		return err
	}
	return EnableAutoStart(exePath)
}

// DisableAutoStart disables auto-start on Windows boot
func (a *App) DisableAutoStart() error {
	return DisableAutoStart()
}

// IsAutoStartEnabled checks if auto-start is enabled
func (a *App) IsAutoStartEnabled() (bool, error) {
	return IsAutoStartEnabled()
}

// ShowWindow shows the main window
func (a *App) ShowWindow() {
	if a.ctx != nil {
		runtime.WindowShow(a.ctx)
	}
}

// HideWindow hides the main window
func (a *App) HideWindow() {
	if a.ctx != nil {
		runtime.WindowHide(a.ctx)
	}
}

