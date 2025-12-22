package watch

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/yaklabco/stave/pkg/watch/mode"
	"github.com/yaklabco/stave/pkg/watch/wctx"
)

func TestDeadlock(t *testing.T) {
	name := "test-target"
	mode.SetOutermostTarget(name)
	mode.SetOverallWatchMode(true)
	ctx := wctx.WithCurrent(context.Background(), name)
	wctx.Register(name, ctx)
	defer wctx.Unregister(name)

	// Ensure the state exists
	_ = GetTargetState(name)

	// Create a dummy file for watcher
	err := os.WriteFile("file.txt", []byte("hello"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("file.txt")

	done1 := make(chan bool)
	done2 := make(chan bool)

	// Thread 1: Call Watch repeatedly
	go func() {
		for range 1000 {
			Watch("file.txt")
		}
		done1 <- true
	}()

	// Thread 2: Call handleFileChange repeatedly
	go func() {
		for range 1000 {
			hfcErr := handleFileChange("file.txt")
			if hfcErr != nil {
				t.Error(hfcErr)
			}
		}
		done2 <- true
	}()

	timeout := time.After(10 * time.Second)
	t1Finished := false
	t2Finished := false

	for !t1Finished || !t2Finished {
		select {
		case <-done1:
			t1Finished = true
		case <-done2:
			t2Finished = true
		case <-timeout:
			t.Fatal("Deadlock detected: threads did not finish within timeout")
		}
	}
}
