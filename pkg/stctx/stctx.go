package stctx

import (
	"context"
	"runtime"
	"strings"
	"sync"

	"github.com/yaklabco/stave/pkg/stack"
)

type contextKey string

const (
	currentTargetKey contextKey = "currentTarget"
	targetStateKey   contextKey = "targetState"
)

var (
	activeContexts sync.Map //nolint:gochecknoglobals // This is intentionally global, and part of a sync.Map pattern.
)

// RegisterTargetContext registers the current context for a target.
func RegisterTargetContext(ctx context.Context, name string) {
	activeContexts.Store(name, ctx)
}

// UnregisterTargetContext unregisters the context for a target.
func UnregisterTargetContext(name string) {
	activeContexts.Delete(name)
}

// GetTargetContext returns the registered context for a target name.
func GetTargetContext(name string) context.Context {
	if v, ok := activeContexts.Load(name); ok {
		resultCtx, ok := v.(context.Context)
		if ok {
			return resultCtx
		}
	}

	return nil
}

// GetActiveContext returns the context of the nearest active target in the call stack.
func GetActiveContext() context.Context {
	pcs := make([]uintptr, stack.MaxStackDepthToCheck)
	n := runtime.Callers(2, pcs) // skip GetActiveContext
	if n == 0 {
		return context.Background()
	}

	frames := runtime.CallersFrames(pcs[:n])
	for {
		frame, more := frames.Next()
		name := DisplayName(frame.Function)
		if ctx := GetTargetContext(name); ctx != nil {
			return ctx
		}
		if !more {
			break
		}
	}
	return context.Background()
}

// ContextWithTarget returns a new context with the target name attached.
func ContextWithTarget(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, currentTargetKey, name)
}

// GetCurrentTarget returns the target name from the context, or empty string if not found.
func GetCurrentTarget(ctx context.Context) string {
	if name, ok := ctx.Value(currentTargetKey).(string); ok {
		return name
	}
	return ""
}

// ContextWithTargetState returns a new context with the target state attached.
func ContextWithTargetState(ctx context.Context, state any) context.Context {
	return context.WithValue(ctx, targetStateKey, state)
}

// GetTargetState returns the target state from the context, or nil if not found.
func GetTargetState(ctx context.Context) any {
	return ctx.Value(targetStateKey)
}

var (
	overallWatchMode bool       //nolint:gochecknoglobals // These are intentionally global, and part of a sync.Mutex pattern.
	outermostTarget  string     //nolint:gochecknoglobals // These are intentionally global, and part of a sync.Mutex pattern.
	watchModeMu      sync.Mutex //nolint:gochecknoglobals // These are intentionally global, and part of a sync.Mutex pattern.
)

// SetOverallWatchMode sets whether we are in overall watch mode.
func SetOverallWatchMode(b bool) {
	watchModeMu.Lock()
	overallWatchMode = b
	watchModeMu.Unlock()
}

// IsOverallWatchMode returns whether we are in overall watch mode.
func IsOverallWatchMode() bool {
	watchModeMu.Lock()
	defer watchModeMu.Unlock()
	return overallWatchMode
}

// SetOutermostTarget sets the name of the outermost target.
func SetOutermostTarget(name string) {
	watchModeMu.Lock()
	outermostTarget = name
	watchModeMu.Unlock()
}

// GetOutermostTarget returns the name of the outermost target.
func GetOutermostTarget() string {
	watchModeMu.Lock()
	defer watchModeMu.Unlock()
	return outermostTarget
}

// DisplayName returns a human-readable name for the target.
func DisplayName(name string) string {
	name = strings.TrimPrefix(name, "main.")
	return strings.ReplaceAll(name, ".", ":")
}
