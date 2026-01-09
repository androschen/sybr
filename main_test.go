package main

import (
	"context"
	"testing"
	"time"
)

// TestGetActiveWindow tests the GetActiveWindow function
func TestGetActiveWindow(t *testing.T) {
	ctx := context.Background()
	watcher := NewWindowWatcher(ctx)

	info, err := watcher.GetActiveWindow()
	if err != nil {
		t.Fatalf("GetActiveWindow() failed: %v", err)
	}

	if info == nil {
		t.Fatal("GetActiveWindow() returned nil info")
	}

	t.Logf("Window Title: %s", info.Title)
	t.Logf("Process Name: %s", info.Exe)

	// At minimum, we should have some process name
	if info.Exe == "" {
		t.Error("Process name is empty")
	}
}

// TestStartMonitoring tests that monitoring can start and stop
func TestStartMonitoring(t *testing.T) {
	ctx := context.Background()
	watcher := NewWindowWatcher(ctx)

	// Start monitoring
	err := watcher.StartMonitoring()
	if err != nil {
		t.Fatalf("StartMonitoring() failed: %v", err)
	}

	// Wait a bit
	time.Sleep(2 * time.Second)

	// Stop monitoring
	watcher.StopMonitoring()

	// Verify we can stop it again without error
	watcher.StopMonitoring()
}


