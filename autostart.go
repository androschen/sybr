package main

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

const (
	// Registry key for current user auto-start
	registryKeyPath = `SOFTWARE\Microsoft\Windows\CurrentVersion\Run`
	appName         = "WindowMonitor"
)

// EnableAutoStart adds the application to Windows startup registry
func EnableAutoStart(exePath string) error {
	// Open the registry key
	key, err := registry.OpenKey(registry.CURRENT_USER, registryKeyPath, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("failed to open registry key: %w", err)
	}
	defer key.Close()

	// Convert to absolute path
	absPath, err := filepath.Abs(exePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Verify the executable exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("executable not found: %s", absPath)
	}

	// Set the registry value
	err = key.SetStringValue(appName, absPath)
	if err != nil {
		return fmt.Errorf("failed to set registry value: %w", err)
	}

	return nil
}

// DisableAutoStart removes the application from Windows startup registry
func DisableAutoStart() error {
	// Open the registry key
	key, err := registry.OpenKey(registry.CURRENT_USER, registryKeyPath, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("failed to open registry key: %w", err)
	}
	defer key.Close()

	// Delete the registry value
	err = key.DeleteValue(appName)
	if err != nil {
		if err == registry.ErrNotExist {
			// Value doesn't exist, which is fine
			return nil
		}
		return fmt.Errorf("failed to delete registry value: %w", err)
	}

	return nil
}

// IsAutoStartEnabled checks if auto-start is currently enabled
func IsAutoStartEnabled() (bool, error) {
	// Open the registry key
	key, err := registry.OpenKey(registry.CURRENT_USER, registryKeyPath, registry.QUERY_VALUE)
	if err != nil {
		return false, fmt.Errorf("failed to open registry key: %w", err)
	}
	defer key.Close()

	// Try to read the value
	_, _, err = key.GetStringValue(appName)
	if err != nil {
		if err == registry.ErrNotExist {
			return false, nil
		}
		return false, fmt.Errorf("failed to read registry value: %w", err)
	}

	return true, nil
}
