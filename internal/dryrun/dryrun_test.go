package dryrun

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const trueStr = "true"

// These tests verify dry-run behavior by spawning a fresh process of the
// current test binary with purpose-built helper flags defined in this package's
// TestMain (see testmain_test.go). Spawning a new process ensures the
// sync.Once guards inside dryrun.go evaluate environment variables afresh.

func TestIsDryRunRequestedEnv(t *testing.T) {
	cmd := exec.Command(os.Args[0], "-printIsDryRunRequested")
	cmd.Env = append(os.Environ(), RequestedEnv+"=1", PossibleEnv+"=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("subprocess failed: %v", err)
	}
	if strings.TrimSpace(string(out)) != trueStr {
		t.Fatalf("expected true, got %q", strings.TrimSpace(string(out)))
	}
}

func TestIsDryRunPossibleEnv(t *testing.T) {
	cmd := exec.Command(os.Args[0], "-printIsDryRunPossible")
	cmd.Env = append(os.Environ(), PossibleEnv+"=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("subprocess failed: %v", err)
	}
	if strings.TrimSpace(string(out)) != trueStr {
		t.Fatalf("expected true, got %q", strings.TrimSpace(string(out)))
	}
}

func TestIsDryRunRequiresBoth(t *testing.T) {
	// Only requested set => not possible, so overall false
	cmd := exec.Command(os.Args[0], "-printIsDryRun")
	cmd.Env = append(os.Environ(), RequestedEnv+"=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("subprocess failed: %v", err)
	}
	if strings.TrimSpace(string(out)) != "false" {
		t.Fatalf("expected false, got %q", strings.TrimSpace(string(out)))
	}

	// Only possible set => not requested, so overall false
	cmd = exec.Command(os.Args[0], "-printIsDryRun")
	cmd.Env = append(os.Environ(), PossibleEnv+"=1")
	out, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("subprocess failed: %v", err)
	}
	if strings.TrimSpace(string(out)) != "false" {
		t.Fatalf("expected false, got %q", strings.TrimSpace(string(out)))
	}

	// Both set => true
	cmd = exec.Command(os.Args[0], "-printIsDryRun")
	cmd.Env = append(os.Environ(), RequestedEnv+"=1", PossibleEnv+"=1")
	out, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("subprocess failed: %v", err)
	}
	if strings.TrimSpace(string(out)) != trueStr {
		t.Fatalf("expected true, got %q", strings.TrimSpace(string(out)))
	}
}

func TestWrap(t *testing.T) {
	tempDir := t.TempDir()
	binDir := filepath.Join(tempDir, "bin")
	if err := os.Mkdir(binDir, 0755); err != nil {
		t.Fatal(err)
	}

	cmdName := "my-wrap-test-cmd"
	cmdPath := filepath.Join(binDir, cmdName)
	if err := os.WriteFile(cmdPath, []byte("#!/bin/sh\necho hello\n"), 0755); err != nil {
		t.Fatal(err)
	}

	// Ensure we are not in dry-run mode.
	SetRequested(false)
	SetPossible(false)

	ctx := context.Background()
	theEnv := map[string]string{
		"PATH": binDir,
	}

	cmd := Wrap(ctx, theEnv, cmdName)

	if cmd.Path != cmdPath {
		t.Errorf("expected cmd.Path to be %q, got %q", cmdPath, cmd.Path)
	}

	t.Run("CommandNotFound", func(t *testing.T) {
		cmdNotFound := Wrap(ctx, theEnv, "non-existent-cmd")
		// It should not have been resolved to a full path in binDir
		if strings.Contains(cmdNotFound.Path, binDir) {
			t.Errorf("command should not have been found in %q, but got %q", binDir, cmdNotFound.Path)
		}
	})
}
