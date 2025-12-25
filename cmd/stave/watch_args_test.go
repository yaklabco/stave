package stave

import (
	"bufio"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestWatchWithArgs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping watch test in short mode")
	}

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	absRoot, err := filepath.Abs("../..")
	require.NoError(t, err)
	tmpDir := t.TempDir()
	staveBin := filepath.Join(tmpDir, "stave_watch_args_test")

	// Ensure stave is built
	buildCmd := exec.Command("go", "build", "-o", staveBin, "github.com/yaklabco/stave")
	buildCmd.Dir = absRoot
	out, err := buildCmd.CombinedOutput()
	require.NoError(t, err, "failed to build stave: %s", string(out))

	watchDir := filepath.Join(tmpDir, "watchme")
	require.NoError(t, os.Mkdir(watchDir, 0755))

	setupStaveProject(t, tmpDir, absRoot, `
func WatchDir(dir string) {
	watch.Watch(fmt.Sprintf("%s/**", dir))
	fmt.Printf("WATCHING_DIR: %s\n", dir)
}
`)

	cmd := exec.CommandContext(ctx, staveBin, "-v", "watchdir", watchDir)
	cmd.Dir = tmpDir

	stdout, err := cmd.StdoutPipe()
	require.NoError(t, err)
	stderr := &strings.Builder{}
	cmd.Stderr = cmd.Stdout
	require.NoError(t, cmd.Start())

	scanner := bufio.NewScanner(stdout)

	waitForOutput := func(expected string) {
		done := make(chan bool, 1)
		go func() {
			for scanner.Scan() {
				line := scanner.Text()
				stderr.WriteString(line + "\n")
				if strings.Contains(line, expected) {
					done <- true
					return
				}
			}
			done <- false
		}()

		select {
		case ok := <-done:
			if !ok {
				t.Errorf("reached EOF waiting for %q. Output: %s", expected, stderr.String())
				return
			}
		case <-ctx.Done():
			t.Errorf("context done waiting for %q. Output: %s", expected, stderr.String())
			return
		case <-time.After(15 * time.Second):
			t.Errorf("timed out waiting for %q. Output: %s", expected, stderr.String())
			return
		}
	}

	waitForOutput("WATCHING_DIR")

	// Trigger a file change
	time.Sleep(1 * time.Second) // Give it a moment to start watching
	require.NoError(t, os.WriteFile(filepath.Join(watchDir, "test.txt"), []byte("hi"), 0644))

	// Wait for re-run
	waitForOutput("WATCHING_DIR")

	if t.Failed() {
		// If it failed, let's see why
		t.Logf("Stderr: %s", stderr.String())
	}

	sigProcessErr := cmd.Process.Signal(os.Interrupt)
	require.NoError(t, sigProcessErr)
	cmdWaitErr := cmd.Wait()
	require.NoError(t, cmdWaitErr)
}
