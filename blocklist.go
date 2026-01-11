package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// BlockedApp represents a blocked application
// Note: Field names must be capitalized for JSON export in Go
type BlockedApp struct {
	ExecutableName string `json:"executableName"` // e.g., "chrome.exe"
	DisplayName    string `json:"displayName"`    // e.g., "Google Chrome"
}

// BlocklistManager manages the blocklist storage
type BlocklistManager struct {
	filePath string
	mu       sync.RWMutex
	apps     []BlockedApp
}

var (
	globalBlocklist *BlocklistManager
	blocklistOnce   sync.Once
)

// getBlocklistPath returns the path to the blocklist JSON file
func getBlocklistPath() (string, error) {
	// In dev mode, use current working directory
	// In production, use executable directory
	wd, err := os.Getwd()
	if err != nil {
		// Fallback to executable directory
		exePath, exeErr := os.Executable()
		if exeErr != nil {
			return "", fmt.Errorf("failed to get working directory or executable path: %w", err)
		}
		wd = filepath.Dir(exePath)
	}
	
	blocklistPath := filepath.Join(wd, "blocking_list.json")
	fmt.Printf("📁 Blocklist file path: %s\n", blocklistPath)
	return blocklistPath, nil
}

// GetBlocklistManager returns the global blocklist manager instance
func GetBlocklistManager() (*BlocklistManager, error) {
	var err error
	blocklistOnce.Do(func() {
		filePath, pathErr := getBlocklistPath()
		if pathErr != nil {
			err = pathErr
			return
		}
		globalBlocklist = &BlocklistManager{
			filePath: filePath,
			apps:     []BlockedApp{},
		}
		// Load existing blocklist
		if loadErr := globalBlocklist.load(); loadErr != nil {
			// If file doesn't exist, that's okay - start with empty list
			if !os.IsNotExist(loadErr) {
				err = loadErr
			}
		}
	})
	return globalBlocklist, err
}

// load reads the blocklist from the JSON file
// Note: Caller must hold the lock if modifying apps
func (bm *BlocklistManager) load() error {
	// Don't acquire lock here - allow caller to manage it
	// This prevents deadlock issues

	data, err := os.ReadFile(bm.filePath)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		bm.mu.Lock()
		bm.apps = []BlockedApp{}
		bm.mu.Unlock()
		return nil
	}

	var apps []BlockedApp
	if err := json.Unmarshal(data, &apps); err != nil {
		return err
	}

	bm.mu.Lock()
	bm.apps = apps
	bm.mu.Unlock()
	return nil
}

// save writes the blocklist to the JSON file
// Note: Caller must hold the lock
func (bm *BlocklistManager) save() error {
	// Don't acquire lock here - caller should already have it
	// This prevents deadlock when called from AddApp/RemoveApp which already hold the lock

	data, err := json.MarshalIndent(bm.apps, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal blocklist: %w", err)
	}

	fmt.Printf("📝 Writing blocklist to file: %s\n", bm.filePath)
	err = os.WriteFile(bm.filePath, data, 0644)
	if err != nil {
		fmt.Printf("❌ WriteFile error: %v\n", err)
		return err
	}
	fmt.Printf("✅ WriteFile completed\n")
	return nil
}

// AddApp adds an app to the blocklist
func (bm *BlocklistManager) AddApp(executableName, displayName string) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	// Normalize executable name (lowercase, ensure .exe)
	executableName = strings.ToLower(strings.TrimSpace(executableName))
	if !strings.HasSuffix(executableName, ".exe") {
		executableName += ".exe"
	}

	fmt.Printf("📝 BlocklistManager.AddApp: executableName=%s, displayName=%s\n", executableName, displayName)

	// Check if already exists
	for _, app := range bm.apps {
		if app.ExecutableName == executableName {
			return fmt.Errorf("app '%s' is already in the blocklist", executableName)
		}
	}

	// Add to list
	if displayName == "" {
		displayName = executableName
	}
	bm.apps = append(bm.apps, BlockedApp{
		ExecutableName: executableName,
		DisplayName:    displayName,
	})

	fmt.Printf("📝 Added app to in-memory list. Total apps: %d\n", len(bm.apps))
	
	err := bm.save()
	if err != nil {
		fmt.Printf("❌ Failed to save blocklist: %v\n", err)
		return err
	}
	
	fmt.Printf("✅ Blocklist saved successfully to: %s\n", bm.filePath)
	return nil
}

// RemoveApp removes an app from the blocklist
func (bm *BlocklistManager) RemoveApp(executableName string) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	// Normalize executable name
	executableName = strings.ToLower(strings.TrimSpace(executableName))
	if !strings.HasSuffix(executableName, ".exe") {
		executableName += ".exe"
	}

	// Find and remove
	newApps := []BlockedApp{}
	found := false
	for _, app := range bm.apps {
		if app.ExecutableName == executableName {
			found = true
			continue
		}
		newApps = append(newApps, app)
	}

	if !found {
		return fmt.Errorf("app '%s' not found in blocklist", executableName)
	}

	bm.apps = newApps
	return bm.save()
}

// GetApps returns a copy of all blocked apps
// It also reloads from file to ensure we have the latest data
func (bm *BlocklistManager) GetApps() []BlockedApp {
	// Reload from file to ensure we have latest data
	if err := bm.load(); err != nil {
		// If file doesn't exist, that's okay - return empty list
		if !os.IsNotExist(err) {
			fmt.Printf("⚠️  Warning: Failed to reload blocklist: %v\n", err)
		}
	}

	bm.mu.RLock()
	apps := make([]BlockedApp, len(bm.apps))
	copy(apps, bm.apps)
	bm.mu.RUnlock()
	
	fmt.Printf("📋 GetApps returning %d apps: %v\n", len(apps), apps)
	return apps
}

// GetExecutableNames returns a list of executable names only
func (bm *BlocklistManager) GetExecutableNames() []string {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	names := make([]string, len(bm.apps))
	for i, app := range bm.apps {
		names[i] = app.ExecutableName
	}
	return names
}

// IsBlocked checks if an executable name is in the blocklist
func (bm *BlocklistManager) IsBlocked(executableName string) bool {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	// Normalize for comparison
	executableName = strings.ToLower(strings.TrimSpace(executableName))

	for _, app := range bm.apps {
		if app.ExecutableName == executableName {
			return true
		}
	}
	return false
}

// GetBlockedApp returns the BlockedApp if found, nil otherwise
func (bm *BlocklistManager) GetBlockedApp(executableName string) *BlockedApp {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	// Normalize for comparison
	executableName = strings.ToLower(strings.TrimSpace(executableName))

	for _, app := range bm.apps {
		if app.ExecutableName == executableName {
			return &app
		}
	}
	return nil
}
