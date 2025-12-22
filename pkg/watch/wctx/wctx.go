package wctx

import (
	"context"
	"runtime"
	"strings"
	"sync"

	"github.com/yaklabco/stave/pkg/stack"
	"github.com/yaklabco/stave/pkg/watch/wtarget"
)

type contextKey int

const (
	targetKey contextKey = iota
	currentTargetKey
)

var (
	activeContexts sync.Map //nolint:gochecknoglobals // This is intentionally global, and part of a sync.Map pattern.
)

// Register registers the current context for a target.
func Register(name string, ctx context.Context) { //nolint:revive // This is intentional, we are registering a context by name.
	activeContexts.Store(name, ctx)
}

// Unregister unregisters the context for a target.
func Unregister(name string) {
	activeContexts.Delete(name)
}

// Get returns the registered context for a target name.
func Get(name string) context.Context {
	if v, ok := activeContexts.Load(name); ok {
		resultCtx, ok := v.(context.Context)
		if ok {
			return resultCtx
		}
	}

	return nil
}

// GetActive returns the context of the nearest active target in the call stack.
func GetActive() context.Context {
	pcs := make([]uintptr, stack.MaxStackDepthToCheck)
	n := runtime.Callers(2, pcs) // skip GetActive
	if n == 0 {
		return context.Background()
	}

	frames := runtime.CallersFrames(pcs[:n])
	for {
		frame, more := frames.Next()
		name := DisplayName(frame.Function)
		if ctx := Get(name); ctx != nil {
			return ctx
		}
		if !more {
			break
		}
	}
	return context.Background()
}

// WithCurrent returns a new context with the target name attached.
func WithCurrent(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, currentTargetKey, name)
}

// GetCurrent returns the target name from the context, or empty string if not found.
func GetCurrent(ctx context.Context) string {
	if name, ok := ctx.Value(currentTargetKey).(string); ok {
		return name
	}
	return ""
}

// WithConfig returns a new context with the target state attached.
func WithConfig(ctx context.Context, t *wtarget.Target) context.Context {
	return context.WithValue(ctx, targetKey, t)
}

// GetConfig returns the target state from the context, or nil if not found.
func GetConfig(ctx context.Context) *wtarget.Target {
	v := ctx.Value(targetKey)
	if t, ok := v.(*wtarget.Target); ok {
		return t
	}

	return nil
}

// DisplayName returns a human-readable name for the target.
func DisplayName(name string) string {
	name = strings.TrimPrefix(name, "main.")
	return strings.ReplaceAll(name, ".", ":")
}
