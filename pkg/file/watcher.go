package file

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// FileChangeCallback is a function type for the callback
type FileChangeCallback func(path string, data []byte, err error)

// Watcher represents a file watcher
type Watcher struct {
	fsWatcher *fsnotify.Watcher
	callback  FileChangeCallback
	done      chan struct{}
	wg        sync.WaitGroup
}

// NewWatcher creates a new Watcher
func NewWatcher(callback FileChangeCallback) (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	return &Watcher{
		fsWatcher: fsWatcher,
		callback:  callback,
		done:      make(chan struct{}),
	}, nil
}

// WatchFiles starts watching the given files for changes
func (w *Watcher) WatchFiles(paths []string) error {
	lastMod := make(map[string]time.Time)
	var mu sync.Mutex

	// Add paths to the watcher
	for _, path := range paths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("failed to get absolute path for %s: %w", path, err)
		}
		err = w.fsWatcher.Add(absPath)
		if err != nil {
			return fmt.Errorf("failed to add %s to watcher: %w", absPath, err)
		}
		// Initialize last modification time
		fileInfo, err := os.Stat(absPath)
		if err != nil {
			return fmt.Errorf("failed to get file info for %s: %w", absPath, err)
		}
		lastMod[absPath] = fileInfo.ModTime()
	}

	// Start watching for changes
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		for {
			select {
			case event, ok := <-w.fsWatcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					mu.Lock()
					lastModTime := lastMod[event.Name]
					mu.Unlock()

					// Check if enough time has passed since the last modification
					if time.Since(lastModTime) > 100*time.Millisecond {
						mu.Lock()
						lastMod[event.Name] = time.Now()
						mu.Unlock()

						// Read the file content
						data, err := os.ReadFile(event.Name)
						if err != nil {
							w.callback(event.Name, nil, fmt.Errorf("failed to read file %s: %w", event.Name, err))
						} else {
							w.callback(event.Name, data, nil)
						}
					}
				}
			case err, ok := <-w.fsWatcher.Errors:
				if !ok {
					return
				}
				w.callback("", nil, fmt.Errorf("watcher error: %w", err))
			case <-w.done:
				return
			}
		}
	}()

	return nil
}

// Stop stops the file watcher
func (w *Watcher) Stop() error {
	close(w.done)
	w.wg.Wait()
	return w.fsWatcher.Close()
}
