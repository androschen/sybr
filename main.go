package main

import (
	"context"
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/getlantern/systray"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

// Global variables to share between Wails and systray
var (
	globalWatcher *WindowWatcher
	globalApp     *App
	wailsCtx      context.Context
)

func main() {
	// Get executable path for auto-start
	exePath, err := os.Executable()
	if err != nil {
		fmt.Printf("Error getting executable path: %v\n", err)
		exePath = ""
	} else {
		exePath, _ = filepath.Abs(exePath)
	}

	// Create app instance
	app := NewApp()
	globalApp = app

	// Create watcher with nil context initially
	// The context will be set in OnStartup with the live Wails context
	watcher := NewWindowWatcher(nil)
	globalWatcher = watcher
	app.watcher = watcher

	// Start systray in a goroutine
	go func() {
		systray.Run(func() {
			setupSystemTray(exePath, watcher)
		}, func() {
			// On exit, stop monitoring
			if watcher != nil {
				watcher.StopMonitoring()
			}
		})
	}()

	// Create application with options
	// In Wails v2, we must explicitly bind the App struct using the Bind field
	// This generates bindings that make methods available via window.go.main.App

	// Check if we're in dev mode
	// Wails v2 should automatically proxy to dev server when running wails dev
	// But if it's not working, we can try to detect dev mode
	isDev := os.Getenv("WAILS_ENV") == "dev" || os.Getenv("devmode") == "true"

	assetServerOptions := &assetserver.Options{
		Assets: assets,
	}

	// In dev mode, Wails should automatically proxy to Vite dev server
	// Make sure Vite is running on http://localhost:34115 before starting Wails
	if isDev {
		fmt.Println("🔧 Dev mode detected - Wails should proxy to Vite dev server")
		fmt.Println("🔧 Make sure Vite is running: cd frontend && npm run dev")
	} else {
		fmt.Println("📦 Production mode - using embedded assets")
	}

	err = wails.Run(&options.App{
		Title:            "Window Monitor",
		Width:            1200,
		Height:           800,
		AssetServer:      assetServerOptions,
		BackgroundColour: &options.RGBA{R: 18, G: 18, B: 18, A: 1},
		// Bind the app instance - this is critical for frontend access
		Bind: []interface{}{
			app,
		},
		OnStartup: func(ctx context.Context) {
			// Store context for systray menu
			wailsCtx = ctx

			// Call app's OnStartup - this will set the context on the watcher
			// and start monitoring with the correct Wails context
			app.OnStartup(ctx)
		},
		OnShutdown: func(ctx context.Context) {
			// Stop monitoring when app shuts down
			if watcher != nil {
				watcher.StopMonitoring()
			}
			// Quit systray
			systray.Quit()
		},
	})

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

func setupSystemTray(exePath string, watcher *WindowWatcher) {
	// Set up system tray icon and menu
	// Only set icon if we have a valid one
	if icon := getIcon(); icon != nil && len(icon) > 0 {
		systray.SetIcon(icon)
	}
	systray.SetTitle("Window Monitor")
	systray.SetTooltip("Window Monitor - Running")

	// Create menu items
	mStatus := systray.AddMenuItem("Window Monitor", "Window Monitor Status")
	mStatus.Disable()
	systray.AddSeparator()

	mShowWindow := systray.AddMenuItem("Show Window", "Show the main window")
	mHideWindow := systray.AddMenuItem("Hide Window", "Hide the main window")
	systray.AddSeparator()

	mEnableAutoStart := systray.AddMenuItem("Enable Auto-Start", "Enable auto-start on Windows boot")
	mDisableAutoStart := systray.AddMenuItem("Disable Auto-Start", "Disable auto-start on Windows boot")
	systray.AddSeparator()

	mQuit := systray.AddMenuItem("Quit", "Quit Window Monitor")

	// Check auto-start status and update menu
	updateAutoStartMenu(mEnableAutoStart, mDisableAutoStart)

	// Handle menu clicks
	go func() {
		for {
			select {
			case <-mShowWindow.ClickedCh:
				if globalApp != nil {
					globalApp.ShowWindow()
				}
			case <-mHideWindow.ClickedCh:
				if globalApp != nil {
					globalApp.HideWindow()
				}
			case <-mEnableAutoStart.ClickedCh:
				if exePath != "" {
					if err := EnableAutoStart(exePath); err != nil {
						fmt.Printf("Failed to enable auto-start: %v\n", err)
						systray.SetTooltip("Window Monitor - Failed to enable auto-start")
					} else {
						fmt.Println("Auto-start enabled successfully")
						systray.SetTooltip("Window Monitor - Auto-start Enabled")
						updateAutoStartMenu(mEnableAutoStart, mDisableAutoStart)
					}
				}
			case <-mDisableAutoStart.ClickedCh:
				if err := DisableAutoStart(); err != nil {
					fmt.Printf("Failed to disable auto-start: %v\n", err)
					systray.SetTooltip("Window Monitor - Failed to disable auto-start")
				} else {
					fmt.Println("Auto-start disabled successfully")
					systray.SetTooltip("Window Monitor - Auto-start Disabled")
					updateAutoStartMenu(mEnableAutoStart, mDisableAutoStart)
				}
			case <-mQuit.ClickedCh:
				if watcher != nil {
					watcher.StopMonitoring()
				}
				// Quit Wails app
				if wailsCtx != nil {
					runtime.Quit(wailsCtx)
				}
				systray.Quit()
				return
			}
		}
	}()
}

// updateAutoStartMenu updates the menu items based on auto-start status
func updateAutoStartMenu(mEnable, mDisable *systray.MenuItem) {
	enabled, err := IsAutoStartEnabled()
	if err != nil {
		fmt.Printf("Error checking auto-start status: %v\n", err)
		return
	}

	if enabled {
		mEnable.Hide()
		mDisable.Show()
	} else {
		mEnable.Show()
		mDisable.Hide()
	}
}

// getIcon returns a simple icon byte array
// For a real application, you would load an actual .ico file
// Returning nil means no custom icon will be set (systray will use default)
func getIcon() []byte {
	// Return nil - systray will handle the default icon
	// To add a custom icon:
	// 1. Create a .ico file (16x16 or 32x32 recommended)
	// 2. Embed it: //go:embed icon.ico
	// 3. var iconData []byte
	// 4. Return iconData here
	return nil
}

// getExecutablePath returns the current executable path
func getExecutablePath() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Abs(exePath)
}
