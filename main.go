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

	// Create context and watcher
	ctx := context.Background()
	watcher := NewWindowWatcher(ctx)
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

	// Create Wails application
	err = wails.Run(&options.App{
		Title:  "Window Monitor",
		Width:  1200,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		// Bind the App struct to make it available in the frontend
		// This allows React to call methods like GetCurrentWindow, EnableAutoStart, etc.
		// Wails will automatically generate bindings for all exported methods
		OnStartup: func(ctx context.Context) {
			// Call app's OnStartup
			app.OnStartup(ctx)

			// Store context
			wailsCtx = ctx
			watcher.ctx = ctx

			// Start monitoring automatically
			if err := watcher.StartMonitoring(); err != nil {
				fmt.Printf("Failed to start window monitoring: %v\n", err)
			} else {
				fmt.Println("Window monitoring started successfully")
			}
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
