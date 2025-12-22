package watch

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/gobwas/glob"
	"github.com/yaklabco/stave/pkg/st"
	"github.com/yaklabco/stave/pkg/watch/mode"
	"github.com/yaklabco/stave/pkg/watch/wctx"
	"github.com/yaklabco/stave/pkg/watch/wstack"
	"github.com/yaklabco/stave/pkg/watch/wtarget"
)

func GetTargetState(name string) *wtarget.Target {
	name = strings.ToLower(name)
	globalMu.Lock()
	defer globalMu.Unlock()
	if s, ok := targets[name]; ok {
		return s
	}
	theState := &wtarget.Target{
		Name:      name,
		RerunChan: make(chan struct{}, 1),
		DepIDs:    make(map[string]bool),
	}
	targets[name] = theState
	return theState
}

// Watch registers glob patterns to watch for the current target.
func Watch(patterns ...string) {
	ctx := wctx.GetActive()
	target := wctx.GetCurrent(ctx)
	if target == "" {
		target = wstack.CallerTargetName()
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
		globalMu.Lock()
		if watcher != nil {
			err := watcher.Add(dir)
			if err != nil {
				panic(fmt.Errorf("failed to add %q to watcher: %w", dir, err))
			}
		}
		globalMu.Unlock()
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
	st.Deps(fns...)

	ctx := wctx.GetActive()
	target := wctx.GetCurrent(ctx)
	if target == "" {
		target = wstack.CallerTargetName()
	}

	outermost := mode.GetOutermostTarget()
	if !mode.IsOverallWatchMode() {
		if target != "" && strings.EqualFold(target, outermost) {
			mode.SetOverallWatchMode(true)
		}
	}

	if !mode.IsOverallWatchMode() {
		return
	}

	// In overall watch mode, we use the outermost target's state for dependency tracking.
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

	for _, theFunc := range fns {
		id := st.F(theFunc).ID()
		if !theState.DepIDs[id] {
			theState.DepIDs[id] = true
			theState.Deps = append(theState.Deps, theFunc)
		}
	}
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
