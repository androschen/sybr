package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/sys/windows"
)

// WindowWatcher monitors the active window on Windows
type WindowWatcher struct {
	ctx           context.Context
	currentTitle  string
	currentExe    string
	mu            sync.RWMutex
	stopChan      chan struct{}
	running       bool
	lastWarnedExe string // Track last warned app to avoid spam
}

// WindowInfo represents information about the active window
type WindowInfo struct {
	Title string `json:"title"`
	Exe   string `json:"exe"`
}

// NewWindowWatcher creates a new WindowWatcher instance
func NewWindowWatcher(ctx context.Context) *WindowWatcher {
	return &WindowWatcher{
		ctx:      ctx,
		stopChan: make(chan struct{}),
	}
}

// SetContext sets the Wails context for event emission
// This must be called with the context from App.OnStartup for events to work
func (ww *WindowWatcher) SetContext(ctx context.Context) {
	ww.mu.Lock()
	defer ww.mu.Unlock()
	ww.ctx = ctx
}

// GetActiveWindow returns the current active window's title and process name
func (ww *WindowWatcher) GetActiveWindow() (*WindowInfo, error) {
	// Get the foreground window handle
	hwnd := windows.GetForegroundWindow()
	if hwnd == 0 {
		return nil, fmt.Errorf("failed to get foreground window")
	}

	// Get window title
	title, err := ww.getWindowTitle(uintptr(hwnd))
	if err != nil {
		return nil, fmt.Errorf("failed to get window title: %w", err)
	}

	// Get process name
	exe, err := ww.getProcessName(uintptr(hwnd))
	if err != nil {
		return nil, fmt.Errorf("failed to get process name: %w", err)
	}

	return &WindowInfo{
		Title: title,
		Exe:   exe,
	}, nil
}

// getWindowTitle retrieves the title of a window using GetWindowTextW
func (ww *WindowWatcher) getWindowTitle(hwnd uintptr) (string, error) {
	user32 := windows.NewLazyDLL("user32.dll")
	getWindowTextLengthW := user32.NewProc("GetWindowTextLengthW")
	getWindowTextW := user32.NewProc("GetWindowTextW")

	// Get the length of the window text
	ret, _, _ := getWindowTextLengthW.Call(uintptr(hwnd))
	length := int32(ret)
	if length == 0 {
		return "", nil // Empty title is valid
	}

	// Allocate buffer with the correct size + 1 for null terminator
	buf := make([]uint16, length+1)
	ret, _, err := getWindowTextW.Call(
		uintptr(hwnd),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)),
	)
	if ret == 0 && err != nil {
		return "", err
	}

	title := windows.UTF16ToString(buf)
	return strings.TrimSpace(title), nil
}

// getProcessName retrieves the executable name of the process owning the window
func (ww *WindowWatcher) getProcessName(hwnd uintptr) (string, error) {
	var processID uint32
	user32 := windows.NewLazyDLL("user32.dll")
	getWindowThreadProcessId := user32.NewProc("GetWindowThreadProcessId")

	// Get the process ID
	ret, _, err := getWindowThreadProcessId.Call(
		uintptr(hwnd),
		uintptr(unsafe.Pointer(&processID)),
	)
	if ret == 0 && err != nil {
		return "", fmt.Errorf("failed to get process ID: %w", err)
	}
	if processID == 0 {
		return "", fmt.Errorf("invalid process ID")
	}

	// Open the process
	processHandle, err := windows.OpenProcess(
		windows.PROCESS_QUERY_INFORMATION|windows.PROCESS_VM_READ,
		false,
		processID,
	)
	if err != nil {
		return "", fmt.Errorf("failed to open process: %w", err)
	}
	defer windows.CloseHandle(processHandle)

	// Get the module base name (GetModuleBaseNameW is in psapi.dll, not kernel32.dll)
	psapi := windows.NewLazyDLL("psapi.dll")
	getModuleBaseNameW := psapi.NewProc("GetModuleBaseNameW")

	buf := make([]uint16, windows.MAX_PATH)
	ret, _, err = getModuleBaseNameW.Call(
		uintptr(processHandle),
		0,
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)),
	)
	if ret == 0 && err != nil {
		return "", fmt.Errorf("failed to get module base name: %w", err)
	}

	exe := windows.UTF16ToString(buf)
	return strings.ToLower(exe), nil
}

// StartMonitoring starts polling the active window every second
// and emits Wails events when the active window changes
func (ww *WindowWatcher) StartMonitoring() error {
	ww.mu.Lock()
	if ww.running {
		ww.mu.Unlock()
		return fmt.Errorf("monitoring already running")
	}
	ww.running = true
	ww.mu.Unlock()

	go ww.monitorLoop()
	return nil
}

// StopMonitoring stops the monitoring loop
func (ww *WindowWatcher) StopMonitoring() {
	ww.mu.Lock()
	defer ww.mu.Unlock()
	if !ww.running {
		return
	}
	close(ww.stopChan)
	ww.running = false
	ww.stopChan = make(chan struct{})
}

// monitorLoop runs the monitoring ticker
func (ww *WindowWatcher) monitorLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	firstWindow := true

	for {
		select {
		case <-ww.stopChan:
			return
		case <-ticker.C:
			info, err := ww.GetActiveWindow()
			if err != nil {
				// Log error but continue monitoring
				fmt.Printf("Error getting active window: %v\n", err)
				continue
			}

			// Skip if info is invalid
			if info == nil || (info.Title == "" && info.Exe == "") {
				continue
			}

			ww.mu.Lock()
			titleChanged := ww.currentTitle != info.Title
			exeChanged := ww.currentExe != info.Exe
			isFirstWindow := firstWindow && (ww.currentTitle == "" && ww.currentExe == "")
			ww.mu.Unlock()

			if titleChanged || exeChanged || isFirstWindow {
				ww.mu.Lock()
				ww.currentTitle = info.Title
				ww.currentExe = info.Exe
				firstWindow = false
				ww.mu.Unlock()

				// Print to console for debugging (terminal output)
				fmt.Printf("Active Window Changed: [%s] %s\n", info.Exe, info.Title)

				// Check if app is blocked
				bm, err := GetBlocklistManager()
				if err == nil && bm != nil {
					// Normalize executable name for comparison
					exeLower := strings.ToLower(strings.TrimSpace(info.Exe))
					if bm.IsBlocked(exeLower) {
						// Only warn if this is a different app (avoid spam)
						ww.mu.Lock()
						shouldWarn := ww.lastWarnedExe != exeLower
						if shouldWarn {
							ww.lastWarnedExe = exeLower
						}
						ww.mu.Unlock()

						if shouldWarn {
							blockedApp := bm.GetBlockedApp(exeLower)
							displayName := exeLower
							if blockedApp != nil && blockedApp.DisplayName != "" {
								displayName = blockedApp.DisplayName
							}

							// Emit warning event
							ww.mu.RLock()
							ctx := ww.ctx
							ww.mu.RUnlock()
							if ctx != nil {
								warningData := map[string]interface{}{
									"executableName": exeLower,
									"displayName":    displayName,
									"title":          info.Title,
								}
								fmt.Printf("⚠️  Blocked app detected: [%s] %s\n", exeLower, info.Title)
								runtime.EventsEmit(ctx, "warning-detected", warningData)
							}
						}
					} else {
						// Reset last warned if app is not blocked
						ww.mu.Lock()
						if ww.lastWarnedExe == exeLower {
							ww.lastWarnedExe = ""
						}
						ww.mu.Unlock()
					}
				}

				// Emit Wails event if context is available
				// This sends the data to the frontend history log
				ww.mu.RLock()
				ctx := ww.ctx
				ww.mu.RUnlock()
				if ctx != nil {
					fmt.Printf("📤 Emitting event 'window-changed' with data: [%s] %s\n", info.Exe, info.Title)
					runtime.EventsEmit(ctx, "window-changed", info)
				} else {
					fmt.Println("⚠️  Cannot emit event: context is nil")
				}
			}
		}
	}
}
