package file

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestWatcher(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "watcher_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create two test files
	file1 := filepath.Join(tempDir, "file1.txt")
	file2 := filepath.Join(tempDir, "file2.txt")

	for _, file := range []string{file1, file2} {
		err := os.WriteFile(file, []byte("initial content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	var mu sync.Mutex
	callbackCalled := make(map[string]int)

	callback := func(path string, data []byte, err error) {
		mu.Lock()
		defer mu.Unlock()
		callbackCalled[path]++
		fmt.Printf("Callback called for %s\n", path)
	}

	watcher, err := NewWatcher(callback)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}

	err = watcher.WatchFiles([]string{file1, file2})
	if err != nil {
		t.Fatalf("Failed to start watching files: %v", err)
	}

	// Wait a bit to ensure watcher is ready
	time.Sleep(100 * time.Millisecond)

	// Modify file1
	err = os.WriteFile(file1, []byte("modified content"), 0644)
	if err != nil {
		t.Fatalf("Failed to modify file1: %v", err)
	}

	// Wait for callback to be called
	time.Sleep(200 * time.Millisecond)

	// Stop the watcher
	err = watcher.Stop()
	if err != nil {
		t.Fatalf("Failed to stop watcher: %v", err)
	}

	// Check callback calls
	mu.Lock()
	defer mu.Unlock()

	if callbackCalled[file1] != 1 {
		t.Errorf("Expected callback to be called once for file1, but was called %d times", callbackCalled[file1])
	}

	if callbackCalled[file2] != 0 {
		t.Errorf("Expected callback not to be called for file2, but was called %d times", callbackCalled[file2])
	}
}
