package watch_test

import (
	"context"
	"testing"
	"time"

	"github.com/yaklabco/stave/pkg/watch"
	"github.com/yaklabco/stave/pkg/watch/mode"
	"github.com/yaklabco/stave/pkg/watch/wctx"
)

func targetA() {
	watch.Deps(targetB)
}

func targetB() {
	watch.Deps(targetA)
}

func TestWatchCycle(t *testing.T) {
	name := "targetA"
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
		targetA()
	}()

	select {
	case <-done:
		if panicVal == nil {
			t.Errorf("Expected panic due to circular dependency, but it didn't happen")
		} else {
			t.Logf("Recovered from expected panic: %v", panicVal)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Test timed out: circular dependency not detected (infinite loop)")
	}
}
