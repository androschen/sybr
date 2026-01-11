package main

import (
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Windows MessageBox constants
const (
	MB_OK                = 0x00000000
	MB_OKCANCEL          = 0x00000001
	MB_ABORTRETRYIGNORE  = 0x00000002
	MB_YESNOCANCEL       = 0x00000003
	MB_YESNO             = 0x00000004
	MB_RETRYCANCEL       = 0x00000005
	MB_ICONERROR         = 0x00000010
	MB_ICONQUESTION      = 0x00000020
	MB_ICONWARNING       = 0x00000030
	MB_ICONINFORMATION   = 0x00000040
	MB_TOPMOST           = 0x00040000
	MB_SETFOREGROUND     = 0x00010000
	MB_SYSTEMMODAL       = 0x00001000
)

// ShowSystemWarning displays a native Windows MessageBox that stays on top of all windows
// It uses MB_TOPMOST to keep it floating above everything, and MB_ICONERROR for visibility
func ShowSystemWarning(title, message string) (int, error) {
	fmt.Printf("🔔 ShowSystemWarning: Preparing MessageBox\n")
	fmt.Printf("   Title: %s\n", title)
	fmt.Printf("   Message: %s\n", message)
	
	// Load user32.dll
	user32 := windows.NewLazyDLL("user32.dll")
	fmt.Printf("✅ user32.dll loaded\n")
	
	// Get MessageBoxW function (Unicode version)
	messageBoxW := user32.NewProc("MessageBoxW")
	fmt.Printf("✅ MessageBoxW proc obtained\n")
	
	// Convert Go strings to UTF-16 pointers
	titlePtr, err := syscall.UTF16PtrFromString(title)
	if err != nil {
		fmt.Printf("❌ Failed to convert title: %v\n", err)
		return 0, fmt.Errorf("failed to convert title to UTF-16: %w", err)
	}
	fmt.Printf("✅ Title converted to UTF-16\n")
	
	messagePtr, err := syscall.UTF16PtrFromString(message)
	if err != nil {
		fmt.Printf("❌ Failed to convert message: %v\n", err)
		return 0, fmt.Errorf("failed to convert message to UTF-16: %w", err)
	}
	fmt.Printf("✅ Message converted to UTF-16\n")
	
	// Combine flags: MB_TOPMOST | MB_ICONERROR | MB_OK | MB_SETFOREGROUND
	// MB_TOPMOST: Keep window on top
	// MB_ICONERROR: Red X icon + error sound
	// MB_OK: Show OK button
	// MB_SETFOREGROUND: Bring to foreground
	flags := MB_TOPMOST | MB_ICONERROR | MB_OK | MB_SETFOREGROUND
	fmt.Printf("📋 Flags: MB_TOPMOST | MB_ICONERROR | MB_OK | MB_SETFOREGROUND (0x%X)\n", flags)
	
	// Call MessageBoxW
	// Parameters:
	//   hWnd: 0 (NULL) - standalone message, not attached to any window
	//   lpText: message pointer
	//   lpCaption: title pointer
	//   uType: flags
	fmt.Printf("📞 Calling MessageBoxW...\n")
	ret, _, err := messageBoxW.Call(
		0,                              // hWnd = NULL (standalone)
		uintptr(unsafe.Pointer(messagePtr)), // lpText
		uintptr(unsafe.Pointer(titlePtr)),   // lpCaption
		uintptr(flags),                      // uType
	)
	
	fmt.Printf("📞 MessageBoxW returned: ret=%d, err=%v\n", ret, err)
	
	if ret == 0 {
		fmt.Printf("❌ MessageBoxW failed - return value is 0\n")
		return 0, fmt.Errorf("MessageBoxW failed: %w", err)
	}
	
	// Return value indicates which button was clicked (IDOK = 1 for OK button)
	fmt.Printf("✅ MessageBox displayed successfully, user clicked button: %d\n", int(ret))
	return int(ret), nil
}
