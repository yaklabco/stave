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
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/yaklabco/stave/pkg/watch"
)

type watchHandle struct {
	t         *testing.T
	cmd       *exec.Cmd
	lines     chan string
	stderr    *strings.Builder
	allStdout *strings.Builder
	mu        sync.Mutex
	cancel    context.CancelFunc
}

func (h *watchHandle) wait(expected string) {
	h.t.Helper()
	timeout := time.After(watch.WatchTestFullDuration)
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
	_ = h.cmd.Process.Signal(os.Interrupt) //nolint:errcheck // Intentionally ignored.
	_ = h.cmd.Wait()                       //nolint:errcheck // Intentionally ignored.
	cancel()
}

func (h *watchHandle) stdout() string {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.allStdout.String()
}

func startWatch(t *testing.T, ctx context.Context, bin, dir string, args ...string) *watchHandle {
	t.Helper()
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Dir = dir
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		if cmd.Process != nil {
			return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		}
		return nil
	}

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

func buildStave(t *testing.T, absRoot, binName string) string {
	t.Helper()
	staveBin := filepath.Join(absRoot, "dist", binName)
	buildCmd := exec.Command("go", "build", "-o", staveBin, "github.com/yaklabco/stave")
	buildCmd.Dir = absRoot
	out, err := buildCmd.CombinedOutput()
	require.NoError(t, err, "failed to build stave: %s", string(out))
	return staveBin
}

func runStave(t *testing.T, bin, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "stave failed: %s", string(out))
	return string(out)
}
