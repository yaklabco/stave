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

func TestWatchFunctionality(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping watch test in short mode")
	}

	// 1. Setup a temporary stave project
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

func WatchTarget() error {
	watch.Watch("%s/**")
	fmt.Println("RUNNING_TARGET")
	return nil
}
`, watchDir)
	require.NoError(t, os.WriteFile(stavefile, []byte(content), 0644))

	// Ensure go.mod is present so it can find stave pkg
	// We'll point to the current project
	absRoot, err := filepath.Abs("../..")
	require.NoError(t, err)

	gomod := fmt.Sprintf(`module testwatch
go 1.23

require github.com/yaklabco/stave v0.0.0
replace github.com/yaklabco/stave => %s
`, absRoot)
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(gomod), 0644))

	// Run go mod tidy to fix dependencies
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = tmpDir
	tidyOut, err := tidyCmd.CombinedOutput()
	require.NoError(t, err, "go mod tidy failed: %s", string(tidyOut))

	// 2. Build stave binary
	staveBin := filepath.Join(absRoot, "dist", "stave_test")
	buildCmd := exec.Command("go", "build", "-o", staveBin, "github.com/yaklabco/stave")
	buildCmd.Dir = absRoot
	out, err := buildCmd.CombinedOutput()
	require.NoError(t, err, "failed to build stave: %s", string(out))

	// 3. Run stave watch
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, staveBin, "watchtarget")
	cmd.Dir = tmpDir

	stderr := &strings.Builder{}
	cmd.Stderr = stderr
	stdout, err := cmd.StdoutPipe()
	require.NoError(t, err)
	require.NoError(t, cmd.Start())

	// Read stdout to wait for output
	scanner := bufio.NewScanner(stdout)
	waitForOutput := func(expected string) {
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, expected) {
				return
			}
		}
		if err := scanner.Err(); err != nil {
			t.Fatalf("error reading stdout: %v", err)
		}
		t.Fatalf("reached EOF while waiting for %q. Stderr: %q", expected, stderr.String())
	}

	waitForOutput("RUNNING_TARGET")

	// 4. Modify a file in watchDir
	testFile := filepath.Join(watchDir, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("hello"), 0644))

	// 5. Check if it re-runs
	waitForOutput("RUNNING_TARGET")

	// Drain stdout in background to prevent blocking
	go func() {
		for scanner.Scan() {
		}
	}()

	// 6. Stop stave gracefully
	require.NoError(t, cmd.Process.Signal(os.Interrupt))

	// 7. Wait for exit with timeout
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		t.Logf("Stave exited with: %v", err)
	case <-time.After(10 * time.Second):
		t.Errorf("Stave did not exit within 10s after SIGINT. Stderr: %s", stderr.String())
		killErr := cmd.Process.Kill()
		if killErr != nil {
			t.Errorf("failed to kill stave process: %v", killErr)
		}
	}
}

func TestWatchDepsFunctionality(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping watch test in short mode")
	}

	// 1. Setup a temporary stave project
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

func DepTarget() {
	fmt.Println("RUNNING_DEP")
}

func WatchTarget() error {
	watch.Watch("%s/**")
	watch.Deps(DepTarget)
	fmt.Println("RUNNING_TARGET")
	return nil
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

	staveBin := filepath.Join(absRoot, "dist", "stave_test")

	// 3. Run stave watch
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, staveBin, "watchtarget")
	cmd.Dir = tmpDir

	stderr := &strings.Builder{}
	cmd.Stderr = stderr
	stdout, err := cmd.StdoutPipe()
	require.NoError(t, err)
	require.NoError(t, cmd.Start())

	// Read stdout to wait for output
	scanner := bufio.NewScanner(stdout)
	waitForOutput := func(expected string) {
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, expected) {
				return
			}
		}
		if err := scanner.Err(); err != nil {
			t.Fatalf("error reading stdout: %v", err)
		}
		t.Fatalf("reached EOF while waiting for %q. Stderr: %q", expected, stderr.String())
	}

	waitForOutput("RUNNING_DEP")
	waitForOutput("RUNNING_TARGET")

	// 4. Modify a file in watchDir
	testFile := filepath.Join(watchDir, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("hello"), 0644))

	// 5. Check if it re-runs
	waitForOutput("RUNNING_DEP")
	waitForOutput("RUNNING_TARGET")

	// Drain stdout in background to prevent blocking
	go func() {
		for scanner.Scan() {
		}
	}()

	// 6. Stop stave gracefully
	require.NoError(t, cmd.Process.Signal(os.Interrupt))

	// 7. Wait for exit with timeout
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		t.Logf("Stave exited with: %v", err)
	case <-time.After(10 * time.Second):
		t.Errorf("Stave did not exit within 10s after SIGINT. Stderr: %s", stderr.String())
		killErr := cmd.Process.Kill()
		if killErr != nil {
			t.Errorf("failed to kill stave process: %v", killErr)
		}
	}
}
