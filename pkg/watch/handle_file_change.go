package watch

import (
	"os"
	"path/filepath"

	"github.com/yaklabco/stave/pkg/watch/wtarget"
)

func handleFileChange(path string) error {
	absPath := path
	if !filepath.IsAbs(path) {
		if a, err := filepath.Abs(path); err == nil {
			absPath = a
		}
	}

	globalMu.Lock()
	if info, err := os.Stat(absPath); err == nil && info.IsDir() {
		if watcher != nil {
			err := watcher.Add(absPath)
			if err != nil {
				globalMu.Unlock()
				return err
			}
		}
	}

	allStates := make([]*wtarget.Target, 0, len(targets))
	for _, s := range targets {
		allStates = append(allStates, s)
	}
	globalMu.Unlock()

	for _, theState := range allStates {
		theState.Mu.Lock()
		matched := false
		for _, g := range theState.Globs {
			if g.Match(absPath) {
				matched = true
				break
			}
		}
		if matched {
			for _, cancel := range theState.CancelFuncs {
				cancel()
			}
			theState.CancelFuncs = nil
			select {
			case theState.RerunChan <- struct{}{}:
			default:
			}
		}
		theState.Mu.Unlock()
	}

	return nil
}
