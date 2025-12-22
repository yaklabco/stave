package watch

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/gobwas/glob"
	"github.com/yaklabco/stave/pkg/st"
	"github.com/yaklabco/stave/pkg/stack"
	"github.com/yaklabco/stave/pkg/watch/mode"
	"github.com/yaklabco/stave/pkg/watch/wctx"
	"github.com/yaklabco/stave/pkg/watch/wtarget"
)

var (
	stateMu sync.Mutex                         //nolint:gochecknoglobals // These are intentionally global, and part of a sync.Mutex pattern.
	states  = make(map[string]*wtarget.Target) //nolint:gochecknoglobals // These are intentionally global, and part of a sync.Mutex pattern.
	watcher *fsnotify.Watcher                  //nolint:gochecknoglobals // These are intentionally global, and part of a sync.Mutex pattern.
)

func GetTargetState(name string) *wtarget.Target {
	name = strings.ToLower(name)
	stateMu.Lock()
	defer stateMu.Unlock()
	if s, ok := states[name]; ok {
		return s
	}
	theState := &wtarget.Target{
		Name:      name,
		RerunChan: make(chan struct{}, 1),
		DepIDs:    make(map[string]bool),
	}
	states[name] = theState
	return theState
}

func callerTargetName() string {
	pcs := make([]uintptr, stack.MaxStackDepthToCheck)
	n := runtime.Callers(3, pcs)
	if n == 0 {
		return ""
	}
	frames := runtime.CallersFrames(pcs[:n])
	for {
		frame, more := frames.Next()
		name := frame.Function
		// Skip internal watch package functions, but allow Test functions for testing.
		if strings.HasPrefix(name, "github.com/yaklabco/stave/pkg/watch.") && !strings.Contains(name, ".Test") {
			if !more {
				break
			}
			continue
		}
		return wctx.DisplayName(name)
	}
	return ""
}

// Watch registers glob patterns to watch for the current target.
func Watch(patterns ...string) {
	ctx := wctx.GetActive()
	target := wctx.GetCurrent(ctx)
	if target == "" {
		target = callerTargetName()
	}
	if target == "" {
		return
	}

	outermost := mode.GetOutermostTarget()
	if !mode.IsOverallWatchMode() {
		if !strings.EqualFold(target, outermost) {
			return
		}

		mode.SetOverallWatchMode(true)
	}

	// In overall watch mode, we always use the outermost target's state.
	theState := GetTargetState(outermost)
	theState.Mu.Lock()
	defer theState.Mu.Unlock()

	foundWatcher := false
	for _, w := range theState.Watchers {
		if w == target {
			foundWatcher = true
			break
		}
	}
	if !foundWatcher {
		theState.Watchers = append(theState.Watchers, target)
	}

	startWatcher()

	for _, p := range patterns {
		absP := p
		if !filepath.IsAbs(p) {
			if a, err := filepath.Abs(p); err == nil {
				absP = a
			}
		}

		// Prevent duplicate patterns
		found := false
		for _, existing := range theState.Patterns {
			if existing == absP {
				found = true
				break
			}
		}
		if found {
			continue
		}

		theState.Patterns = append(theState.Patterns, absP)
		g, err := glob.Compile(absP)
		if err == nil {
			theState.Globs = append(theState.Globs, g)
		}

		// Add non-wildcard prefix to watcher
		dir := absP
		if idx := strings.IndexAny(absP, "*?[]{}"); idx != -1 {
			dir = absP[:idx]
			if lastSlash := strings.LastIndexAny(dir, "/\\"); lastSlash != -1 {
				dir = dir[:lastSlash]
			} else {
				dir = "."
			}
		}
		stateMu.Lock()
		if watcher != nil {
			err := watcher.Add(dir)
			if err != nil {
				panic(fmt.Errorf("failed to add %q to watcher: %w", dir, err))
			}
		}
		stateMu.Unlock()
	}

	ctx, cancel := context.WithCancel(ctx)
	theState.CancelFuncs = append(theState.CancelFuncs, cancel)

	// Register the new cancellable context as the active context for the target AND outermost target.
	wctx.Register(target, ctx)
	if !strings.EqualFold(target, outermost) {
		wctx.Register(outermost, ctx)
	}
}

// Deps registers watch-specific dependencies for the current target.
func Deps(fns ...any) {
	ctx := wctx.GetActive()
	target := wctx.GetCurrent(ctx)
	if target == "" {
		target = callerTargetName()
	}

	outermost := mode.GetOutermostTarget()
	if !mode.IsOverallWatchMode() {
		if target != "" && strings.EqualFold(target, outermost) {
			mode.SetOverallWatchMode(true)
		}
	}

	if !mode.IsOverallWatchMode() {
		st.CtxDeps(ctx, fns...)
		return
	}

	// In overall watch mode, we use the outermost target's state for dependency tracking.
	theState := GetTargetState(outermost)
	theState.Mu.Lock()

	foundWatcher := false
	for _, w := range theState.Watchers {
		if w == target {
			foundWatcher = true
			break
		}
	}
	if !foundWatcher {
		theState.Watchers = append(theState.Watchers, target)
	}

	toRun := make([]any, 0, len(fns))
	for _, theFunc := range fns {
		id := st.F(theFunc).ID()
		if !theState.DepIDs[id] {
			theState.DepIDs[id] = true
			theState.Deps = append(theState.Deps, theFunc)
		}
		toRun = append(toRun, theFunc)
	}
	theState.Mu.Unlock()

	// Run them now
	runDeps(ctx, toRun)
}

func runDeps(ctx context.Context, fns []any) {
	var wg sync.WaitGroup
	for _, theFunc := range fns {
		wg.Add(1)
		go func(fn any) {
			defer wg.Done()

			depRunErr := st.RunFn(ctx, fn)
			if depRunErr != nil {
				fatalErr := st.Fatalf(1, "dependency failed: %v", depRunErr)
				if fatalErr != nil {
					slog.Error("dependency failed, and so did call to st.Fatalf",
						slog.Any("dependency_error", depRunErr),
						slog.Any("st_fatalf_error", fatalErr),
					)
				}
			}
		}(theFunc)
	}
	wg.Wait()
}

func startWatcher() {
	stateMu.Lock()
	defer stateMu.Unlock()
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

func handleFileChange(path string) error {
	absPath := path
	if !filepath.IsAbs(path) {
		if a, err := filepath.Abs(path); err == nil {
			absPath = a
		}
	}

	stateMu.Lock()
	if info, err := os.Stat(absPath); err == nil && info.IsDir() {
		if watcher != nil {
			err := watcher.Add(absPath)
			if err != nil {
				stateMu.Unlock()
				return err
			}
		}
	}

	allStates := make([]*wtarget.Target, 0, len(states))
	for _, s := range states {
		allStates = append(allStates, s)
	}
	stateMu.Unlock()

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

// ResetWatchDeps resets the once-cache for all dependencies registered via watch.Deps for the given target.
func ResetWatchDeps(target string) {
	theState := GetTargetState(target)
	theState.Mu.Lock()
	defer theState.Mu.Unlock()
	st.ResetSpecificOnces(theState.Deps...)
	st.ResetOncesByName(theState.Watchers...)
}

// RerunLoop should be called by the main function for the outermost target if in watch mode.
func RerunLoop(ctx context.Context, targetName string, fn func() error) {
	theState := GetTargetState(targetName)
	for {
		select {
		case <-ctx.Done():
			return
		case <-theState.RerunChan:
			slog.Info("WATCH MODE: re-running target", "target", targetName)
			reRunErr := fn()
			if reRunErr != nil {
				fatalErr := st.Fatalf(1, "re-run failed: %v", reRunErr)
				if fatalErr != nil {
					slog.Error("rerun failed, and so did call to st.Fatalf",
						slog.Any("rerun_error", reRunErr),
						slog.Any("st_fatalf_error", fatalErr),
					)
				}
			}
		}
	}
}
