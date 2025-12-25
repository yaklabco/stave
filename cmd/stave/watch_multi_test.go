package stave

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestWatchMultiTarget(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping watch test in short mode")
	}

	tmpDir := t.TempDir()
	stavefile := filepath.Join(tmpDir, "stavefile.go")
	watchDir := filepath.Join(tmpDir, "watchme")
	require.NoError(t, os.Mkdir(watchDir, 0755))

	content := fmt.Sprintf(`//go:build stave
package main

import (
	"fmt"
	"github.com/yaklabco/stave/pkg/watch"
)

func WatchTarget() {
	watch.Watch("%s/**")
	fmt.Println("RUNNING_WATCH_TARGET")
}

func OtherTarget() {
	fmt.Println("RUNNING_OTHER_TARGET")
}
`, watchDir)
	require.NoError(t, os.WriteFile(stavefile, []byte(content), 0644))

	absRoot, err := filepath.Abs("../..")
	require.NoError(t, err)

	gomod := fmt.Sprintf(`module testwatch
go 1.23

require github.com/yaklabco/stave v0.0.0
replace github.com/yaklabco/stave => %s
`, absRoot)
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(gomod), 0644))

	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = tmpDir
	tidyOut, err := tidyCmd.CombinedOutput()
	require.NoError(t, err, "go mod tidy failed: %s", string(tidyOut))

	staveBin := filepath.Join(absRoot, "dist", "stave_test_multi")
	buildCmd := exec.Command("go", "build", "-o", staveBin, "github.com/yaklabco/stave")
	buildCmd.Dir = absRoot
	out, err := buildCmd.CombinedOutput()
	require.NoError(t, err, "failed to build stave: %s", string(out))

	// Case 1: WatchTarget OtherTarget
	runTest := func(targets ...string) (bool, bool, int) {
		ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
		defer cancel()

		args := append([]string{"-v"}, targets...)
		cmd := exec.CommandContext(ctx, staveBin, args...)
		cmd.Dir = tmpDir

		stdout, err := cmd.StdoutPipe()
		require.NoError(t, err)
		cmd.Stderr = cmd.Stdout
		require.NoError(t, cmd.Start())

		scanner := bufio.NewScanner(stdout)

		outputChan := make(chan string)
		go func() {
			for scanner.Scan() {
				outputChan <- scanner.Text()
			}
			close(outputChan)
		}()

		firstRunWatch := false
		firstRunOther := false
		watching := false
		reRunCount := 0

		timer := time.NewTimer(5 * time.Second)
		defer timer.Stop()

		for {
			select {
			case line, ok := <-outputChan:
				if !ok {
					return firstRunWatch, firstRunOther, reRunCount
				}
				if strings.Contains(line, "RUNNING_WATCH_TARGET") {
					if !firstRunWatch {
						firstRunWatch = true
					} else {
						reRunCount++
					}
				}
				if strings.Contains(line, "RUNNING_OTHER_TARGET") {
					firstRunOther = true
				}
				if strings.Contains(line, "watching for changes...") {
					watching = true
				}

				if firstRunWatch && firstRunOther && watching && reRunCount == 0 {
					// Trigger change
					time.Sleep(200 * time.Millisecond) // Give watcher time to start
					testFile := filepath.Join(watchDir, "test.txt")
					require.NoError(t, os.WriteFile(testFile, []byte("hello"), 0644))
				}
				if reRunCount > 0 {
					cancel() // Stop the process
				}
			case <-timer.C:
				cancel()
				return firstRunWatch, firstRunOther, reRunCount
			case <-ctx.Done():
				return firstRunWatch, firstRunOther, reRunCount
			}
		}
	}

	t.Run("WatchTarget_OtherTarget", func(t *testing.T) {
		w, o, r := runTest("watchtarget", "othertarget")
		if !w || !o {
			t.Errorf("Expected both to run, got watch=%v, other=%v", w, o)
		}
		if r == 0 {
			t.Errorf("Expected re-run, got 0")
		}
	})

	t.Run("OtherTarget_WatchTarget", func(t *testing.T) {
		w, o, r := runTest("othertarget", "watchtarget")
		if !w || !o {
			t.Errorf("Expected both to run, got watch=%v, other=%v", w, o)
		}
		// If bug exists, it might not re-run because WatchTarget is not outermost.
		if r == 0 {
			t.Errorf("Expected re-run, got 0")
		}
	})
}
