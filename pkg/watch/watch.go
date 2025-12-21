package watch

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/gobwas/glob"
	"github.com/yaklabco/stave/internal/ish"
	"github.com/yaklabco/stave/pkg/st"
	"github.com/yaklabco/stave/pkg/stack"
	"github.com/yaklabco/stave/pkg/stctx"
)

var (
	stateMu sync.Mutex                      //nolint:gochecknoglobals // These are intentionally global, and part of a sync.Mutex pattern.
	states  = make(map[string]*targetState) //nolint:gochecknoglobals // These are intentionally global, and part of a sync.Mutex pattern.
	watcher *fsnotify.Watcher               //nolint:gochecknoglobals // These are intentionally global, and part of a sync.Mutex pattern.
)

type targetState struct {
	name     string
	patterns []string
	globs    []glob.Glob
	deps     []interface{}
	watchers []string
	depIDs   map[string]bool
	cancels  []context.CancelFunc
	mu       sync.Mutex
	reRun    chan struct{}
}

func getTargetState(name string) *targetState {
	name = strings.ToLower(name)
	stateMu.Lock()
	defer stateMu.Unlock()
	if s, ok := states[name]; ok {
		return s
	}
	theState := &targetState{
		name:   name,
		reRun:  make(chan struct{}, 1),
		depIDs: make(map[string]bool),
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
		return stctx.DisplayName(name)
	}
	return ""
}

// Watch registers glob patterns to watch for the current target.
func Watch(patterns ...string) {
	ctx := stctx.GetActiveContext()
	target := stctx.GetCurrentTarget(ctx)
	if target == "" {
		target = callerTargetName()
	}
	if target == "" {
		return
	}

	outermost := stctx.GetOutermostTarget()
	if !stctx.IsOverallWatchMode() {
		if !strings.EqualFold(target, outermost) {
			return
		}

		stctx.SetOverallWatchMode(true)
	}

	// In overall watch mode, we always use the outermost target's state.
	theState := getTargetState(outermost)
	theState.mu.Lock()
	defer theState.mu.Unlock()

	foundWatcher := false
	for _, w := range theState.watchers {
		if w == target {
			foundWatcher = true
			break
		}
	}
	if !foundWatcher {
		theState.watchers = append(theState.watchers, target)
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
		for _, existing := range theState.patterns {
			if existing == absP {
				found = true
				break
			}
		}
		if found {
			continue
		}

		theState.patterns = append(theState.patterns, absP)
		g, err := glob.Compile(absP)
		if err == nil {
			theState.globs = append(theState.globs, g)
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
	theState.cancels = append(theState.cancels, cancel)

	// Register the new cancellable context as the active context for the target AND outermost target.
	stctx.RegisterTargetContext(ctx, target)
	if !strings.EqualFold(target, outermost) {
		stctx.RegisterTargetContext(ctx, outermost)
	}
}

// Deps registers watch-specific dependencies for the current target.
func Deps(fns ...interface{}) {
	ctx := stctx.GetActiveContext()
	target := stctx.GetCurrentTarget(ctx)
	if target == "" {
		target = callerTargetName()
	}

	outermost := stctx.GetOutermostTarget()
	if !stctx.IsOverallWatchMode() {
		if target != "" && strings.EqualFold(target, outermost) {
			stctx.SetOverallWatchMode(true)
		}
	}

	if !stctx.IsOverallWatchMode() {
		st.CtxDeps(ctx, fns...)
		return
	}

	// In overall watch mode, we use the outermost target's state for dependency tracking.
	theState := getTargetState(outermost)
	theState.mu.Lock()

	foundWatcher := false
	for _, w := range theState.watchers {
		if w == target {
			foundWatcher = true
			break
		}
	}
	if !foundWatcher {
		theState.watchers = append(theState.watchers, target)
	}

	toRun := make([]interface{}, 0, len(fns))
	for _, theFunc := range fns {
		id := st.F(theFunc).ID()
		if !theState.depIDs[id] {
			theState.depIDs[id] = true
			theState.deps = append(theState.deps, theFunc)
		}
		toRun = append(toRun, theFunc)
	}
	theState.mu.Unlock()

	// Run them now
	runDeps(ctx, toRun)
}

func runDeps(ctx context.Context, fns []interface{}) {
	var wg sync.WaitGroup
	for _, theFunc := range fns {
		wg.Add(1)
		go func(fn interface{}) {
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

	allStates := make([]*targetState, 0, len(states))
	for _, s := range states {
		allStates = append(allStates, s)
	}
	stateMu.Unlock()

	for _, theState := range allStates {
		theState.mu.Lock()
		matched := false
		for _, g := range theState.globs {
			if g.Match(absPath) {
				matched = true
				break
			}
		}
		if matched {
			for _, cancel := range theState.cancels {
				cancel()
			}
			theState.cancels = nil
			select {
			case theState.reRun <- struct{}{}:
			default:
			}
		}
		theState.mu.Unlock()
	}

	return nil
}

// sh counterparts

func RunCmd(cmd string, args ...string) func(args ...string) error {
	return func(args2 ...string) error {
		return Run(cmd, append(args, args2...)...)
	}
}

func OutCmd(cmd string, args ...string) func(args ...string) (string, error) {
	return func(args2 ...string) (string, error) {
		return Output(cmd, append(args, args2...)...)
	}
}

func Run(cmd string, args ...string) error {
	return RunWith(nil, cmd, args...)
}

func RunV(cmd string, args ...string) error {
	_, err := Exec(nil, os.Stdin, os.Stdout, os.Stderr, cmd, args...)
	return err
}

func RunWith(env map[string]string, cmd string, args ...string) error {
	return ish.Run(stctx.GetActiveContext(), env, cmd, args...)
}

func RunWithV(env map[string]string, cmd string, args ...string) error {
	return ish.RunV(stctx.GetActiveContext(), env, cmd, args...)
}

func Output(cmd string, args ...string) (string, error) {
	return ish.Output(stctx.GetActiveContext(), nil, cmd, args...)
}

func OutputWith(env map[string]string, cmd string, args ...string) (string, error) {
	return ish.Output(stctx.GetActiveContext(), env, cmd, args...)
}

func Piper(stdin io.Reader, stdout, stderr io.Writer, cmd string, args ...string) error {
	return ish.Piper(stctx.GetActiveContext(), nil, stdin, stdout, stderr, cmd, args...)
}

func PiperWith(env map[string]string, stdin io.Reader, stdout, stderr io.Writer, cmd string, args ...string) error {
	return ish.Piper(stctx.GetActiveContext(), env, stdin, stdout, stderr, cmd, args...)
}

func Exec(env map[string]string, stdin io.Reader, stdout, stderr io.Writer, cmd string, args ...string) (bool, error) {
	return ish.Exec(stctx.GetActiveContext(), env, stdin, stdout, stderr, cmd, args...)
}

func Rm(path string) error {
	ctx := stctx.GetActiveContext()
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return ish.Rm(path)
}

func Copy(dst string, src string) error {
	ctx := stctx.GetActiveContext()
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return ish.Copy(dst, src)
}

func CmdRan(err error) bool {
	return ish.CmdRan(err)
}

func ExitStatus(err error) int {
	return ish.ExitStatus(err)
}

// IsOverallWatchMode returns whether we are in overall watch mode.
func IsOverallWatchMode() bool {
	return stctx.IsOverallWatchMode()
}

// SetOutermostTarget sets the name of the outermost target.
func SetOutermostTarget(name string) {
	stctx.SetOutermostTarget(name)
}

// RegisterTargetContext registers the current context for a target.
func RegisterTargetContext(ctx context.Context, name string) {
	stctx.RegisterTargetContext(ctx, name)
}

// UnregisterTargetContext unregisters the context for a target.
func UnregisterTargetContext(name string) {
	stctx.UnregisterTargetContext(name)
}

// ResetWatchDeps resets the once-cache for all dependencies registered via watch.Deps for the given target.
func ResetWatchDeps(target string) {
	theState := getTargetState(target)
	theState.mu.Lock()
	defer theState.mu.Unlock()
	st.ResetSpecificOnces(theState.deps...)
	st.ResetOncesByName(theState.watchers...)
}

// ReRunLoop should be called by the main function for the outermost target if in watch mode.
func ReRunLoop(ctx context.Context, targetName string, fn func() error) {
	theState := getTargetState(targetName)
	for {
		select {
		case <-ctx.Done():
			return
		case <-theState.reRun:
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
