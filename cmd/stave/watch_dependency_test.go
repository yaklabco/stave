package stave

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestWatchDependencyBehavior(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping watch test in short mode")
	}

	absRoot, err := filepath.Abs("../..")
	require.NoError(t, err)
	staveBin := filepath.Join(absRoot, "dist", "stave_test")

	// Ensure stave is built
	buildCmd := exec.Command("go", "build", "-o", staveBin, "github.com/yaklabco/stave")
	buildCmd.Dir = absRoot
	out, err := buildCmd.CombinedOutput()
	require.NoError(t, err, "failed to build stave: %s", string(out))

	t.Run("SimplePrint", func(t *testing.T) {
		tmpDir := t.TempDir()
		setupStaveProject(t, tmpDir, absRoot, `
func Hello() {
	_ = watch.Watch
	fmt.Println("HELLO_WORLD")
}
`)
		out := runStave(t, staveBin, tmpDir, "-v", "hello")
		require.Contains(t, out, "HELLO_WORLD")
	})

	t.Run("TransitiveWatch_Ignored_When_Outermost_Uses_st.Deps", func(t *testing.T) {
		tmpDir := t.TempDir()
		setupStaveProject(t, tmpDir, absRoot, `
import "github.com/yaklabco/stave/pkg/st"
func Sub() {
	watch.Watch("watchme/**")
	fmt.Println("RUNNING_SUB")
}
func Outermost() {
	st.Deps(Sub)
	fmt.Println("RUNNING_OUTERMOST")
}
`)
		require.NoError(t, os.Mkdir(filepath.Join(tmpDir, "watchme"), 0755))

		out := runStave(t, staveBin, tmpDir, "-v", "outermost")
		require.Contains(t, out, "RUNNING_SUB")
		require.Contains(t, out, "RUNNING_OUTERMOST")
		require.NotContains(t, out, "WATCH MODE: re-running target")
	})

	t.Run("TransitiveWatch_Respected_When_Outermost_Uses_watch.Deps", func(t *testing.T) {
		tmpDir := t.TempDir()
		watchDir := filepath.Join(tmpDir, "watchme")
		require.NoError(t, os.Mkdir(watchDir, 0755))

		setupStaveProject(t, tmpDir, absRoot, fmt.Sprintf(`
func Sub() {
	watch.Watch("%s/**")
	fmt.Println("RUNNING_SUB")
}
func Outermost() {
	fmt.Println("STARTING_OUTERMOST")
	watch.Deps(Sub)
	fmt.Println("RUNNING_OUTERMOST")
}
`, watchDir))

		ctx := t.Context()

		theWatchHandle := startWatch(t, ctx, staveBin, tmpDir, "-v", "outermost")
		defer theWatchHandle.stop()

		theWatchHandle.wait("RUNNING_SUB")
		theWatchHandle.wait("RUNNING_OUTERMOST")

		// Modify file to trigger re-run
		require.NoError(t, os.WriteFile(filepath.Join(watchDir, "test.txt"), []byte("hi"), 0644))

		theWatchHandle.wait("RUNNING_SUB")
		theWatchHandle.wait("RUNNING_OUTERMOST")
	})

	t.Run("TransitiveWatch_Respected_When_Outermost_Uses_watch.Watch", func(t *testing.T) {
		tmpDir := t.TempDir()
		watchDir1 := filepath.Join(tmpDir, "watch1")
		watchDir2 := filepath.Join(tmpDir, "watch2")
		require.NoError(t, os.Mkdir(watchDir1, 0755))
		require.NoError(t, os.Mkdir(watchDir2, 0755))

		setupStaveProject(t, tmpDir, absRoot, fmt.Sprintf(`
import "github.com/yaklabco/stave/pkg/st"
func Sub() {
	watch.Watch("%s/**")
	fmt.Println("RUNNING_SUB")
}
func Outermost() {
	watch.Watch("%s/**")
	st.Deps(Sub)
	fmt.Println("RUNNING_OUTERMOST")
}
`, watchDir2, watchDir1))

		ctx := t.Context()

		theWatchHandle := startWatch(t, ctx, staveBin, tmpDir, "-v", "outermost")
		defer theWatchHandle.stop()

		theWatchHandle.wait("RUNNING_SUB")
		theWatchHandle.wait("RUNNING_OUTERMOST")

		// Modify file in Sub's watch dir
		require.NoError(t, os.WriteFile(filepath.Join(watchDir2, "test.txt"), []byte("hi"), 0644))

		theWatchHandle.wait("RUNNING_SUB")
		theWatchHandle.wait("RUNNING_OUTERMOST")
	})

	t.Run("WatchRerunBehavior_OnlyResetsWatchDeps", func(t *testing.T) {
		tmpDir := t.TempDir()
		watchDir := filepath.Join(tmpDir, "watchme")
		require.NoError(t, os.Mkdir(watchDir, 0755))

		setupStaveProject(t, tmpDir, absRoot, fmt.Sprintf(`
import "github.com/yaklabco/stave/pkg/st"

func Build() {
	fmt.Println("RUNNING_BUILD")
}

func LintGo() {
	fmt.Println("RUNNING_LINT_GO")
}

func WatchDir() {
	st.Deps(Build)
	watch.Deps(LintGo)
	watch.Watch("%s/**")
	fmt.Println("RUNNING_WATCHDIR")
}
`, watchDir))

		ctx := t.Context()

		theWatchHandle := startWatch(t, ctx, staveBin, tmpDir, "-v", "watchdir")
		defer theWatchHandle.stop()

		// Initial run
		theWatchHandle.wait("RUNNING_BUILD")
		theWatchHandle.wait("RUNNING_LINT_GO")
		theWatchHandle.wait("RUNNING_WATCHDIR")

		// Trigger re-run
		time.Sleep(1 * time.Second)
		require.NoError(t, os.WriteFile(filepath.Join(watchDir, "test.txt"), []byte("change1"), 0644))

		theWatchHandle.wait("RUNNING_LINT_GO")
		theWatchHandle.wait("RUNNING_WATCHDIR")

		// Trigger another re-run
		time.Sleep(500 * time.Millisecond)
		require.NoError(t, os.WriteFile(filepath.Join(watchDir, "test.txt"), []byte("change2"), 0644))
		theWatchHandle.wait("RUNNING_LINT_GO")
		theWatchHandle.wait("RUNNING_WATCHDIR")

		theWatchHandle.stop()

		allOutput := theWatchHandle.stdout()
		buildCount := strings.Count(allOutput, "RUNNING_BUILD")
		lintCount := strings.Count(allOutput, "RUNNING_LINT_GO")

		require.Equal(t, 1, buildCount, "Build should only run once. Full output:\n%s", allOutput)
		require.GreaterOrEqual(t, lintCount, 3, "LintGo should run at least 3 times. Full output:\n%s", allOutput)
	})
}

func (h *watchHandle) wait(expected string) {
	h.t.Helper()
	timeout := time.After(15 * time.Second)
	for {
		select {
		case line, ok := <-h.lines:
			if !ok {
				h.t.Fatalf("reached EOF waiting for %q. Stderr: %s", expected, h.stderr.String())
			}
			if strings.Contains(line, expected) {
				return
			}
		case <-timeout:
			h.t.Fatalf("timed out waiting for %q. Stderr: %s", expected, h.stderr.String())
		}
	}
}

func (h *watchHandle) stop() {
	h.t.Helper()
	h.mu.Lock()
	defer h.mu.Unlock()
	cancel := h.cancel
	if cancel == nil {
		return
	}
	h.cancel = nil

	// Ignore error if process already finished
	require.NoError(h.t, h.cmd.Process.Signal(os.Interrupt))
	require.NoError(h.t, h.cmd.Wait())
	cancel()
}

func (h *watchHandle) stdout() string {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.allStdout.String()
}

func runStave(t *testing.T, bin, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "stave failed: %s", string(out))
	return string(out)
}

type watchHandle struct {
	t         *testing.T
	cmd       *exec.Cmd
	lines     chan string
	stderr    *strings.Builder
	allStdout *strings.Builder
	mu        sync.Mutex
	cancel    context.CancelFunc
}

func startWatch(t *testing.T, ctx context.Context, bin, dir string, args ...string) *watchHandle {
	t.Helper()
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Dir = dir

	stdout, err := cmd.StdoutPipe()
	require.NoError(t, err)
	stderr := &strings.Builder{}
	cmd.Stderr = cmd.Stdout

	require.NoError(t, cmd.Start())

	handle := &watchHandle{
		t:         t,
		cmd:       cmd,
		lines:     make(chan string, 100),
		stderr:    stderr,
		allStdout: &strings.Builder{},
		cancel:    cancel,
	}

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			handle.mu.Lock()
			handle.allStdout.WriteString(line + "\n")
			handle.mu.Unlock()
			handle.lines <- line
		}
		close(handle.lines)
	}()

	return handle
}

func setupStaveProject(t *testing.T, tmpDir, absRoot, targets string) {
	t.Helper()

	stavefile := filepath.Join(tmpDir, "stavefile.go")

	content := fmt.Sprintf(`//go:build stave
package main

import (
	"fmt"
	"github.com/yaklabco/stave/pkg/watch"
)

%s
`, targets)
	require.NoError(t, os.WriteFile(stavefile, []byte(content), 0644))

	gomod := fmt.Sprintf(`module testwatch
go 1.23

require github.com/yaklabco/stave v0.0.0
replace github.com/yaklabco/stave => %s
`, absRoot)
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(gomod), 0644))

	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = tmpDir
	out, err := tidyCmd.CombinedOutput()
	require.NoError(t, err, "go mod tidy failed: %s", string(out))
}
