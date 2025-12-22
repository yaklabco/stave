package watch_test

import (
	"context"
	"testing"
	"time"

	"github.com/yaklabco/stave/pkg/st"
	"github.com/yaklabco/stave/pkg/watch"
	"github.com/yaklabco/stave/pkg/watch/mode"
	"github.com/yaklabco/stave/pkg/watch/wctx"
)

// Helper to run a test that is expected to panic due to a cycle
func assertCyclePanic(t *testing.T, name string, fn func()) {
	t.Helper()
	st.ResetCycles()
	st.ResetOnces()
	mode.SetOutermostTarget(name)
	mode.SetOverallWatchMode(true)
	ctx := wctx.WithCurrent(context.Background(), name)
	wctx.Register(name, ctx)
	defer wctx.Unregister(name)

	done := make(chan bool)
	var panicVal any

	go func() {
		defer func() {
			panicVal = recover()
			done <- true
		}()
		fn()
	}()

	select {
	case <-done:
		if panicVal == nil {
			t.Errorf("Expected panic due to circular dependency in %s, but it didn't happen", t.Name())
		} else {
			t.Logf("Recovered from expected panic in %s: %v", t.Name(), panicVal)
		}
	case <-time.After(2 * time.Second):
		t.Errorf("Test %s timed out: circular dependency not detected (infinite loop)", t.Name())
	}
}

// Case 1: watch.Deps -> watch.Deps
func targetWatchA() { watch.Deps(targetWatchB) }
func targetWatchB() { watch.Deps(targetWatchA) }

func TestWatchWatchCycle(t *testing.T) {
	assertCyclePanic(t, "targetWatchA", targetWatchA)
}

// Case 2: st.Deps -> st.Deps
func targetStA() { st.Deps(targetStB) }
func targetStB() { st.Deps(targetStA) }

func TestStStCycle(t *testing.T) {
	assertCyclePanic(t, "targetStA", targetStA)
}

// Case 3: watch.Deps -> st.Deps
func targetWatchStA() { watch.Deps(targetWatchStB) }
func targetWatchStB() { st.Deps(targetWatchStA) }

func TestWatchStCycle(t *testing.T) {
	assertCyclePanic(t, "targetWatchStA", targetWatchStA)
}

// Case 4: st.Deps -> watch.Deps
func targetStWatchA() { st.Deps(targetStWatchB) }
func targetStWatchB() { watch.Deps(targetStWatchA) }

func TestStWatchCycle(t *testing.T) {
	assertCyclePanic(t, "targetStWatchA", targetStWatchA)
}

// Case 5: Longer cycle with mix
func targetMixA() { watch.Deps(targetMixB) }
func targetMixB() { st.Deps(targetMixC) }
func targetMixC() { watch.Deps(targetMixA) }

func TestMixCycle(t *testing.T) {
	assertCyclePanic(t, "targetMixA", targetMixA)
}
