package stave

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yaklabco/stave/config"
	"github.com/yaklabco/stave/internal/hooks"
)

// testConfigPerm is the permission mode for test config files.
const testConfigPerm = 0o644

// testHookPerm is the permission mode for test hook scripts.
const testHookPerm = 0o755

// testChdir changes to the given directory and returns a cleanup function.
func testChdir(t *testing.T, dir string) {
	t.Helper()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd failed: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir failed: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Errorf("Cleanup Chdir failed: %v", err)
		}
	})
}

// copyModFiles copies go.mod and go.sum from the repo root to the destination directory.
// This is needed for tests that compile stavefiles in temp directories.
func copyModFiles(t *testing.T, dstDir string) {
	t.Helper()

	// Find the repo root by looking for go.mod
	repoRoot, err := findRepoRoot()
	if err != nil {
		t.Fatalf("Failed to find repo root: %v", err)
	}

	// Copy go.mod
	modContent, err := os.ReadFile(filepath.Join(repoRoot, "go.mod"))
	if err != nil {
		t.Fatalf("Failed to read go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dstDir, "go.mod"), modContent, testConfigPerm); err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Copy go.sum
	sumContent, err := os.ReadFile(filepath.Join(repoRoot, "go.sum"))
	if err != nil {
		t.Fatalf("Failed to read go.sum: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dstDir, "go.sum"), sumContent, testConfigPerm); err != nil {
		t.Fatalf("Failed to write go.sum: %v", err)
	}
}

// findRepoRoot finds the repository root by looking for go.mod.
func findRepoRoot() (string, error) {
	// Start from the current working directory
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			return "", os.ErrNotExist
		}
		dir = parent
	}
}

func TestRunHooksCommand_Help(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	code := RunHooksCommand(&stdout, &stderr, []string{"-h"})

	if code != 0 {
		t.Errorf("RunHooksCommand -h returned %d, want 0", code)
	}
	if !strings.Contains(stdout.String(), "stave hooks") {
		t.Error("Help output should contain 'stave hooks'")
	}
}

func TestRunHooksCommand_UnknownSubcommand(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	code := RunHooksCommand(&stdout, &stderr, []string{"unknown"})

	if code != 2 {
		t.Errorf("RunHooksCommand unknown returned %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "unknown hooks subcommand") {
		t.Error("Error should mention unknown subcommand")
	}
}

func TestRunHooksCommand_Init(t *testing.T) {
	t.Parallel()

	config.ResetGlobal()

	var stdout, stderr bytes.Buffer
	code := RunHooksCommand(&stdout, &stderr, []string{"init"})

	if code != 0 {
		t.Errorf("RunHooksCommand init returned %d, want 0: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "hooks:") {
		t.Error("Init output should show example configuration")
	}
}

func TestRunHooksCommand_List_NoConfig(t *testing.T) {
	t.Parallel()

	config.ResetGlobal()

	var stdout, stderr bytes.Buffer
	code := RunHooksCommand(&stdout, &stderr, []string{"list"})

	if code != 0 {
		t.Errorf("RunHooksCommand list returned %d, want 0: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "No hooks configured") {
		t.Error("List output should indicate no hooks configured")
	}
}

func TestRunHooksCommand_List_WithConfig(t *testing.T) {
	config.ResetGlobal()

	// Create temp directory with config
	tmpDir := t.TempDir()
	configContent := `
hooks:
  pre-commit:
    - target: fmt
    - target: lint
  pre-push:
    - target: test
`
	configPath := filepath.Join(tmpDir, "stave.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), testConfigPerm); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	testChdir(t, tmpDir)

	var stdout, stderr bytes.Buffer
	code := RunHooksCommand(&stdout, &stderr, []string{"list"})

	if code != 0 {
		t.Errorf("RunHooksCommand list returned %d, want 0: %s", code, stderr.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "pre-commit") {
		t.Error("List output should contain 'pre-commit'")
	}
	if !strings.Contains(out, "pre-push") {
		t.Error("List output should contain 'pre-push'")
	}
	if !strings.Contains(out, "fmt") {
		t.Error("List output should contain target 'fmt'")
	}
}

func TestRunHooksCommand_Install_NotGitRepo(t *testing.T) {
	config.ResetGlobal()

	// Create temp directory (not a git repo)
	tmpDir := t.TempDir()
	configContent := `
hooks:
  pre-commit:
    - target: fmt
`
	configPath := filepath.Join(tmpDir, "stave.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), testConfigPerm); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	testChdir(t, tmpDir)

	var stdout, stderr bytes.Buffer
	code := RunHooksCommand(&stdout, &stderr, []string{"install"})

	if code != 1 {
		t.Errorf("RunHooksCommand install in non-git repo returned %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "not a git repository") {
		t.Error("Error should mention not a git repository")
	}
}

func TestRunHooksCommand_Install_NoHooksConfig(t *testing.T) {
	config.ResetGlobal()

	// Create temp directory with git repo but no hooks config
	tmpDir := t.TempDir()

	// Resolve symlinks (macOS /var -> /private/var)
	tmpDir, err := filepath.EvalSymlinks(tmpDir)
	if err != nil {
		t.Fatalf("EvalSymlinks failed: %v", err)
	}

	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("git init failed: %v", err)
	}

	// Create empty config (no hooks)
	configPath := filepath.Join(tmpDir, "stave.yaml")
	if err := os.WriteFile(configPath, []byte("verbose: true\n"), testConfigPerm); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	testChdir(t, tmpDir)

	var stdout, stderr bytes.Buffer
	code := RunHooksCommand(&stdout, &stderr, []string{"install"})

	if code != 1 {
		t.Errorf("RunHooksCommand install with no hooks returned %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "No hooks configured") {
		t.Error("Error should mention no hooks configured")
	}
}

func TestRunHooksCommand_Install_CreatesScripts(t *testing.T) {
	config.ResetGlobal()

	// Create temp directory with git repo and hooks config
	tmpDir := t.TempDir()

	// Resolve symlinks (macOS /var -> /private/var)
	tmpDir, err := filepath.EvalSymlinks(tmpDir)
	if err != nil {
		t.Fatalf("EvalSymlinks failed: %v", err)
	}

	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("git init failed: %v", err)
	}

	configContent := `
hooks:
  pre-commit:
    - target: fmt
  pre-push:
    - target: test
`
	configPath := filepath.Join(tmpDir, "stave.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), testConfigPerm); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	testChdir(t, tmpDir)

	var stdout, stderr bytes.Buffer
	code := RunHooksCommand(&stdout, &stderr, []string{"install"})

	if code != 0 {
		t.Errorf("RunHooksCommand install returned %d, want 0: %s", code, stderr.String())
	}

	// Verify hooks were created
	hooksDir := filepath.Join(tmpDir, ".git", "hooks")
	preCommitPath := filepath.Join(hooksDir, "pre-commit")
	prePushPath := filepath.Join(hooksDir, "pre-push")

	if _, err := os.Stat(preCommitPath); os.IsNotExist(err) {
		t.Error("pre-commit hook should have been created")
	}
	if _, err := os.Stat(prePushPath); os.IsNotExist(err) {
		t.Error("pre-push hook should have been created")
	}

	// Verify they are Stave-managed
	managed, err := hooks.IsStaveManaged(preCommitPath)
	if err != nil {
		t.Fatalf("IsStaveManaged failed: %v", err)
	}
	if !managed {
		t.Error("pre-commit should be Stave-managed")
	}
}

func TestRunHooksCommand_Install_ExistingNonStaveHook_Fails(t *testing.T) {
	config.ResetGlobal()

	// Create temp directory with git repo
	tmpDir := t.TempDir()

	// Resolve symlinks (macOS /var -> /private/var)
	tmpDir, err := filepath.EvalSymlinks(tmpDir)
	if err != nil {
		t.Fatalf("EvalSymlinks failed: %v", err)
	}

	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("git init failed: %v", err)
	}

	// Create existing non-Stave hook
	hooksDir := filepath.Join(tmpDir, ".git", "hooks")
	preCommitPath := filepath.Join(hooksDir, "pre-commit")
	if err := os.WriteFile(preCommitPath, []byte("#!/bin/sh\necho 'custom hook'\n"), testHookPerm); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	configContent := `
hooks:
  pre-commit:
    - target: fmt
`
	configPath := filepath.Join(tmpDir, "stave.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), testConfigPerm); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	testChdir(t, tmpDir)

	var stdout, stderr bytes.Buffer
	code := RunHooksCommand(&stdout, &stderr, []string{"install"})

	if code != 1 {
		t.Errorf("RunHooksCommand install with existing hook returned %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "not installed by Stave") {
		t.Error("Error should mention hook was not installed by Stave")
	}
}

func TestRunHooksCommand_Install_Force(t *testing.T) {
	config.ResetGlobal()

	// Create temp directory with git repo
	tmpDir := t.TempDir()

	// Resolve symlinks (macOS /var -> /private/var)
	tmpDir, err := filepath.EvalSymlinks(tmpDir)
	if err != nil {
		t.Fatalf("EvalSymlinks failed: %v", err)
	}

	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("git init failed: %v", err)
	}

	// Create existing non-Stave hook
	hooksDir := filepath.Join(tmpDir, ".git", "hooks")
	preCommitPath := filepath.Join(hooksDir, "pre-commit")
	if err := os.WriteFile(preCommitPath, []byte("#!/bin/sh\necho 'custom hook'\n"), testHookPerm); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	configContent := `
hooks:
  pre-commit:
    - target: fmt
`
	configPath := filepath.Join(tmpDir, "stave.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), testConfigPerm); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	testChdir(t, tmpDir)

	var stdout, stderr bytes.Buffer
	code := RunHooksCommand(&stdout, &stderr, []string{"install", "--force"})

	if code != 0 {
		t.Errorf("RunHooksCommand install --force returned %d, want 0: %s", code, stderr.String())
	}

	// Verify hook was overwritten
	managed, err := hooks.IsStaveManaged(preCommitPath)
	if err != nil {
		t.Fatalf("IsStaveManaged failed: %v", err)
	}
	if !managed {
		t.Error("pre-commit should now be Stave-managed after --force")
	}
}

func TestRunHooksCommand_Install_UpdatesStaveHook(t *testing.T) {
	config.ResetGlobal()

	// Create temp directory with git repo
	tmpDir := t.TempDir()

	// Resolve symlinks (macOS /var -> /private/var)
	tmpDir, err := filepath.EvalSymlinks(tmpDir)
	if err != nil {
		t.Fatalf("EvalSymlinks failed: %v", err)
	}

	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("git init failed: %v", err)
	}

	// Create existing Stave hook
	hooksDir := filepath.Join(tmpDir, ".git", "hooks")
	preCommitPath := filepath.Join(hooksDir, "pre-commit")
	if err := hooks.WriteHookScript(preCommitPath, hooks.ScriptParams{HookName: "pre-commit"}); err != nil {
		t.Fatalf("WriteHookScript failed: %v", err)
	}

	configContent := `
hooks:
  pre-commit:
    - target: fmt
`
	configPath := filepath.Join(tmpDir, "stave.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), testConfigPerm); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	testChdir(t, tmpDir)

	var stdout, stderr bytes.Buffer
	code := RunHooksCommand(&stdout, &stderr, []string{"install"})

	if code != 0 {
		t.Errorf("RunHooksCommand install with existing Stave hook returned %d, want 0: %s", code, stderr.String())
	}
}

func TestRunHooksCommand_Uninstall(t *testing.T) {
	config.ResetGlobal()

	// Create temp directory with git repo
	tmpDir := t.TempDir()

	// Resolve symlinks (macOS /var -> /private/var)
	tmpDir, err := filepath.EvalSymlinks(tmpDir)
	if err != nil {
		t.Fatalf("EvalSymlinks failed: %v", err)
	}

	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("git init failed: %v", err)
	}

	// Install a hook first
	hooksDir := filepath.Join(tmpDir, ".git", "hooks")
	preCommitPath := filepath.Join(hooksDir, "pre-commit")
	if err := hooks.WriteHookScript(preCommitPath, hooks.ScriptParams{HookName: "pre-commit"}); err != nil {
		t.Fatalf("WriteHookScript failed: %v", err)
	}

	configContent := `
hooks:
  pre-commit:
    - target: fmt
`
	configPath := filepath.Join(tmpDir, "stave.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), testConfigPerm); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	testChdir(t, tmpDir)

	var stdout, stderr bytes.Buffer
	code := RunHooksCommand(&stdout, &stderr, []string{"uninstall"})

	if code != 0 {
		t.Errorf("RunHooksCommand uninstall returned %d, want 0: %s", code, stderr.String())
	}

	// Verify hook was removed
	if _, err := os.Stat(preCommitPath); !os.IsNotExist(err) {
		t.Error("pre-commit hook should have been removed")
	}
}

func TestRunHooksCommand_Run_NoHookName(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	code := RunHooksCommand(&stdout, &stderr, []string{"run"})

	if code != 2 {
		t.Errorf("RunHooksCommand run without hook name returned %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "hook name required") {
		t.Error("Error should mention hook name required")
	}
}

func TestRunHooksCommand_Run_Success(t *testing.T) {
	config.ResetGlobal()

	// Create temp directory with config and stavefile
	tmpDir := t.TempDir()

	// Resolve symlinks (macOS /var -> /private/var)
	tmpDir, err := filepath.EvalSymlinks(tmpDir)
	if err != nil {
		t.Fatalf("EvalSymlinks failed: %v", err)
	}

	// Copy go.mod and go.sum for compilation
	copyModFiles(t, tmpDir)

	// Copy the test stavefile
	srcStavefile := filepath.Join("testdata", "hooks", "stavefile.go")
	srcContent, err := os.ReadFile(srcStavefile)
	if err != nil {
		t.Fatalf("ReadFile stavefile failed: %v", err)
	}
	dstStavefile := filepath.Join(tmpDir, "stavefile.go")
	if err := os.WriteFile(dstStavefile, srcContent, testConfigPerm); err != nil {
		t.Fatalf("WriteFile stavefile failed: %v", err)
	}

	configContent := `
hooks:
  pre-commit:
    - target: HookTest
`
	configPath := filepath.Join(tmpDir, "stave.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), testConfigPerm); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	testChdir(t, tmpDir)

	var stdout, stderr bytes.Buffer
	code := RunHooksCommand(&stdout, &stderr, []string{"run", "pre-commit"})

	if code != 0 {
		t.Errorf("RunHooksCommand run pre-commit returned %d, want 0\nstdout: %s\nstderr: %s",
			code, stdout.String(), stderr.String())
	}
}

func TestRunHooksCommand_Run_UnconfiguredHook(t *testing.T) {
	config.ResetGlobal()

	// Create temp directory with config (but no pre-push)
	tmpDir := t.TempDir()
	configContent := `
hooks:
  pre-commit:
    - target: fmt
`
	configPath := filepath.Join(tmpDir, "stave.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), testConfigPerm); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	testChdir(t, tmpDir)

	var stdout, stderr bytes.Buffer
	code := RunHooksCommand(&stdout, &stderr, []string{"run", "pre-push"})

	// Unconfigured hooks should pass
	if code != 0 {
		t.Errorf("RunHooksCommand run unconfigured hook returned %d, want 0", code)
	}
}

func TestRunHooksCommand_Default_ShowsList(t *testing.T) {
	config.ResetGlobal()

	var stdout, stderr bytes.Buffer
	code := RunHooksCommand(&stdout, &stderr, nil)

	// Default behavior is to show list
	if code != 0 {
		t.Errorf("RunHooksCommand with no args returned %d, want 0", code)
	}
}

func TestRunHooksCommand_Run_ExecutesRealTarget(t *testing.T) {
	config.ResetGlobal()

	// Create temp directory
	tmpDir := t.TempDir()

	// Resolve symlinks (macOS /var -> /private/var)
	tmpDir, err := filepath.EvalSymlinks(tmpDir)
	if err != nil {
		t.Fatalf("EvalSymlinks failed: %v", err)
	}

	// Copy go.mod and go.sum for compilation
	copyModFiles(t, tmpDir)

	// Copy the test stavefile
	srcStavefile := filepath.Join("testdata", "hooks", "stavefile.go")
	srcContent, err := os.ReadFile(srcStavefile)
	if err != nil {
		t.Fatalf("ReadFile stavefile failed: %v", err)
	}
	dstStavefile := filepath.Join(tmpDir, "stavefile.go")
	if err := os.WriteFile(dstStavefile, srcContent, testConfigPerm); err != nil {
		t.Fatalf("WriteFile stavefile failed: %v", err)
	}

	// Create marker file path
	markerPath := filepath.Join(tmpDir, "marker.txt")

	// Create stave.yaml with hooks config
	configContent := `
hooks:
  pre-commit:
    - target: HookTest
`
	configPath := filepath.Join(tmpDir, "stave.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), testConfigPerm); err != nil {
		t.Fatalf("WriteFile config failed: %v", err)
	}

	testChdir(t, tmpDir)

	// Set marker env var
	t.Setenv("HOOK_TEST_MARKER", markerPath)

	var stdout, stderr bytes.Buffer
	code := RunHooksCommand(&stdout, &stderr, []string{"run", "pre-commit"})

	if code != 0 {
		t.Errorf("RunHooksCommand run returned %d, want 0\nstdout: %s\nstderr: %s",
			code, stdout.String(), stderr.String())
	}

	// Verify the marker file was created (proving the target actually ran)
	markerContent, err := os.ReadFile(markerPath)
	if err != nil {
		t.Fatalf("Marker file not created - target did not execute: %v\nstdout: %s\nstderr: %s",
			err, stdout.String(), stderr.String())
	}
	if string(markerContent) != "executed" {
		t.Errorf("Marker content = %q, want %q", string(markerContent), "executed")
	}
}

func TestRunHooksCommand_Run_TargetFailure(t *testing.T) {
	config.ResetGlobal()

	// Create temp directory
	tmpDir := t.TempDir()

	// Resolve symlinks (macOS /var -> /private/var)
	tmpDir, err := filepath.EvalSymlinks(tmpDir)
	if err != nil {
		t.Fatalf("EvalSymlinks failed: %v", err)
	}

	// Copy go.mod and go.sum for compilation
	copyModFiles(t, tmpDir)

	// Copy the test stavefile
	srcStavefile := filepath.Join("testdata", "hooks", "stavefile.go")
	srcContent, err := os.ReadFile(srcStavefile)
	if err != nil {
		t.Fatalf("ReadFile stavefile failed: %v", err)
	}
	dstStavefile := filepath.Join(tmpDir, "stavefile.go")
	if err := os.WriteFile(dstStavefile, srcContent, testConfigPerm); err != nil {
		t.Fatalf("WriteFile stavefile failed: %v", err)
	}

	// Create stave.yaml with hooks config targeting the failing target
	configContent := `
hooks:
  pre-commit:
    - target: HookFail
`
	configPath := filepath.Join(tmpDir, "stave.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), testConfigPerm); err != nil {
		t.Fatalf("WriteFile config failed: %v", err)
	}

	testChdir(t, tmpDir)

	var stdout, stderr bytes.Buffer
	code := RunHooksCommand(&stdout, &stderr, []string{"run", "pre-commit"})

	// The target returns an error, so the hook should fail
	if code == 0 {
		t.Errorf("RunHooksCommand run with failing target returned 0, want non-zero\nstdout: %s\nstderr: %s",
			stdout.String(), stderr.String())
	}
}
