package watch

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/stave/pkg/stctx"
)

func TestWatchRegistration(t *testing.T) {
	name := stctx.DisplayName("github.com/yaklabco/stave/pkg/watch.TestWatchRegistration")
	stctx.SetOutermostTarget(name)
	ctx := stctx.ContextWithTarget(context.Background(), name)
	stctx.RegisterTargetContext(ctx, name)
	defer stctx.UnregisterTargetContext(name)

	Watch("*.txt")

	assert.True(t, stctx.IsOverallWatchMode())
	s := getTargetState(name)
	absTxt, err := filepath.Abs("*.txt")
	require.NoError(t, err)
	assert.Contains(t, s.patterns, absTxt)
}

func TestWatchDeps(t *testing.T) {
	name := stctx.DisplayName("github.com/yaklabco/stave/pkg/watch.TestWatchDeps")
	stctx.SetOverallWatchMode(true)
	stctx.SetOutermostTarget(name)
	ctx := stctx.ContextWithTarget(context.Background(), name)
	stctx.RegisterTargetContext(ctx, name)
	defer stctx.UnregisterTargetContext(name)

	var runCount int
	depFn := func() {
		runCount++
	}

	Deps(depFn)
	assert.Equal(t, 1, runCount)

	s := getTargetState(name)
	assert.Len(t, s.deps, 1)
}

func TestWatchCancellation(t *testing.T) {
	name := stctx.DisplayName("github.com/yaklabco/stave/pkg/watch.TestWatchCancellation")
	stctx.SetOutermostTarget(name)
	stctx.SetOverallWatchMode(true)
	ctx := stctx.ContextWithTarget(context.Background(), name)
	stctx.RegisterTargetContext(ctx, name)
	defer stctx.UnregisterTargetContext(name)

	Watch("*.go")

	// Get the updated context from registry
	ctx = stctx.GetTargetContext(name)
	require.NoError(t, ctx.Err())

	// Simulate file change
	require.NoError(t, handleFileChange("test.go"))

	// Context should be cancelled
	assert.ErrorIs(t, ctx.Err(), context.Canceled)
}

func TestCallerTargetName(t *testing.T) {
	var name string
	var fullName string
	MainTarget := func() { //nolint:gocritic,revive // Keeping realistic capitalization for test fidelity.
		name = callerTargetName()
		pc, _, _, _ := runtime.Caller(1)
		fullName = runtime.FuncForPC(pc).Name()
	}

	MainTarget()
	t.Logf("Full caller name: %s", fullName)
	t.Logf("Caller target name: %s", name)
}
