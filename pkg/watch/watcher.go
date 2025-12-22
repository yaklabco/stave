package watch

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/yaklabco/stave/pkg/st"
	"github.com/yaklabco/stave/pkg/watch/wtarget"
)

var (
	globalMu sync.Mutex                         //nolint:gochecknoglobals // These are intentionally global, and part of a sync.Mutex pattern.
	targets  = make(map[string]*wtarget.Target) //nolint:gochecknoglobals // These are intentionally global, and part of a sync.Mutex pattern.
	watcher  *fsnotify.Watcher                  //nolint:gochecknoglobals // These are intentionally global, and part of a sync.Mutex pattern.
)

func startWatcher() {
	globalMu.Lock()
	defer globalMu.Unlock()
	if watcher != nil {
		return
	}
	var err error
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) != 0 {
					hfcErr := handleFileChange(event.Name)
					if hfcErr != nil {
						panic(fmt.Errorf("failed to handle file change %q: %w", event.Name, hfcErr))
					}
				}
			case _, ok := <-watcher.Errors:
				if !ok {
					return
				}
			}
		}
	}()

	// Watch current directory and its subdirectories
	walkErr := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return watcher.Add(path)
		}
		return nil
	})

	if walkErr != nil {
		fatalErr := st.Fatalf(1, "failed to start watcher: %v", walkErr)
		if fatalErr != nil {
			slog.Error("starting watcher failed, and so did call to st.Fatalf",
				slog.Any("watcher_error", walkErr),
				slog.Any("st_fatalf_error", fatalErr),
			)
		}
	}
}
