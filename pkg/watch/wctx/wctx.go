package wctx

import (
	"context"
	"runtime"
	"strings"
	"sync"

	"github.com/yaklabco/stave/pkg/stack"
)

type contextKey int

const (
	targetKey contextKey = iota
	currentTargetKey
)

var (
	activeContexts sync.Map //nolint:gochecknoglobals // This is intentionally global, and part of a sync.Map pattern.
)

// RegisterContext registers the current context for a target.
func RegisterContext(name string, ctx context.Context) { //nolint:revive // This is intentional, we are registering a context by name.
	activeContexts.Store(name, ctx)
}

// UnregisterContext unregisters the context for a target.
func UnregisterContext(name string) {
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
	return context.WithValue(ctx, targetKey, state)
}

// GetTargetState returns the target state from the context, or nil if not found.
func GetTargetState(ctx context.Context) any {
	return ctx.Value(targetKey)
}

// DisplayName returns a human-readable name for the target.
func DisplayName(name string) string {
	name = strings.TrimPrefix(name, "main.")
	return strings.ReplaceAll(name, ".", ":")
}
