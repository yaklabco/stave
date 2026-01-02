package stave

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/yaklabco/stave/pkg/watch"
)

func TestWatchMultiTarget(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping watch test in short mode")
	}

	absRoot, err := filepath.Abs("../..")
	require.NoError(t, err)
	staveBin := buildStave(t, absRoot, "stave_test_multi")

	runTest := func(targets ...string) (bool, bool, int) {
		var firstRunWatch, firstRunOther bool
		var reRunCount int
		tmpDir := t.TempDir()
		watchDir := filepath.Join(tmpDir, "watchme")
		require.NoError(t, os.Mkdir(watchDir, 0755))

		setupStaveProject(t, tmpDir, absRoot, fmt.Sprintf(`
func WatchTarget() {
	watch.Watch("%s/**")
	fmt.Println("RUNNING_WATCH_TARGET")
}

func OtherTarget() {
	fmt.Println("RUNNING_OTHER_TARGET")
}
`, watchDir))

		ctx := t.Context()
		args := append([]string{"-v"}, targets...)
		handle := startWatch(t, ctx, staveBin, tmpDir, args...)
		defer handle.stop()

		timeout := time.After(watch.WatchTestHalfDuration)
		for {
			select {
			case line, ok := <-handle.lines:
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

				if firstRunWatch && firstRunOther && reRunCount == 0 {
					// Trigger change
					time.Sleep(500 * time.Millisecond) // Give watcher time to start
					testFile := filepath.Join(watchDir, "test.txt")
					require.NoError(t, os.WriteFile(testFile, []byte("hello"), 0644))
				}
				if reRunCount > 0 {
					return firstRunWatch, firstRunOther, reRunCount
				}
			case <-timeout:
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
		if r == 0 {
			t.Errorf("Expected re-run, got 0")
		}
	})
}
