package main

import (
	"context"
	"fmt"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx     context.Context
	watcher *WindowWatcher
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		ctx:     nil,
		watcher: nil, // Will be set in main.go
	}
}

// OnStartup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) OnStartup(ctx context.Context) {
	a.ctx = ctx

	// Set the Wails context on the watcher - this is critical for events to work
	if a.watcher != nil {
		a.watcher.SetContext(ctx)
		fmt.Println("✓ Watcher context set in OnStartup")

		// Start monitoring in a goroutine to avoid blocking
		go func() {
			if err := a.watcher.StartMonitoring(); err != nil {
				fmt.Printf("❌ Failed to start window monitoring: %v\n", err)
			} else {
				fmt.Println("✓ Window monitoring started successfully")

				// Emit initial window immediately after starting
				go func() {
					time.Sleep(500 * time.Millisecond) // Give it a moment to initialize
					if info, err := a.watcher.GetActiveWindow(); err == nil && info != nil {
						fmt.Printf("📤 Emitting initial window: [%s] %s\n", info.Exe, info.Title)
						runtime.EventsEmit(ctx, "window-changed", info)
					}
				}()
			}
		}()
	} else {
		fmt.Println("❌ Watcher is nil in OnStartup!")
	}
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
	if a.watcher == nil {
		return
	}
	go func() {
		for {
			info, err := a.watcher.GetActiveWindow()
			if err != nil {
				continue
			}
			fn(info)
		}
	}()
	// This will be handled by the watcher emitting events
	// The frontend will use runtime.EventsOn to listen
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
