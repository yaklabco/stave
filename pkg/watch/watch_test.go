package watch

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/stave/pkg/watch/mode"
	"github.com/yaklabco/stave/pkg/watch/wctx"
	"github.com/yaklabco/stave/pkg/watch/wstack"
)

func TestWatchRegistration(t *testing.T) {
	name := wctx.DisplayName("github.com/yaklabco/stave/pkg/watch.TestWatchRegistration")
	mode.SetOutermostTarget(name)
	ctx := wctx.WithCurrent(context.Background(), name)
	wctx.Register(name, ctx)
	defer wctx.Unregister(name)

	Watch("*.txt")

	assert.True(t, mode.IsOverallWatchMode())
	s := GetTargetState(name)
	absTxt, err := filepath.Abs("*.txt")
	require.NoError(t, err)
	assert.Contains(t, s.Patterns, absTxt)
}

func TestWatchDeps(t *testing.T) {
	name := wctx.DisplayName("github.com/yaklabco/stave/pkg/watch.TestWatchDeps")
	mode.SetOverallWatchMode(true)
	mode.SetOutermostTarget(name)
	ctx := wctx.WithCurrent(context.Background(), name)
	wctx.Register(name, ctx)
	defer wctx.Unregister(name)

	var runCount int
	depFn := func() {
		runCount++
	}

	Deps(depFn)
	assert.Equal(t, 1, runCount)

	s := GetTargetState(name)
	assert.Len(t, s.Deps, 1)
}

func TestWatchCancellation(t *testing.T) {
	name := wctx.DisplayName("github.com/yaklabco/stave/pkg/watch.TestWatchCancellation")
	mode.SetOutermostTarget(name)
	mode.SetOverallWatchMode(true)
	ctx := wctx.WithCurrent(context.Background(), name)
	wctx.Register(name, ctx)
	defer wctx.Unregister(name)

	Watch("*.go")

	// Get the updated context from registry
	ctx = wctx.Get(name)
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
		name = wstack.CallerTargetName()
		pc, _, _, _ := runtime.Caller(1)
		fullName = runtime.FuncForPC(pc).Name()
	}

	MainTarget()
	t.Logf("Full caller name: %s", fullName)
	t.Logf("Caller target name: %s", name)
}
