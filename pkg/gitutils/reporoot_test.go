package gitutils

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetRepoRoot(t *testing.T) {
	// Create a temp directory for the test
	tmpDir := t.TempDir()
	tmpDir, err := filepath.EvalSymlinks(tmpDir)
	require.NoError(t, err)

	t.Chdir(tmpDir)
	_, err = GetRepoRoot()
	require.Error(t, err)
	require.ErrorIs(t, err, ErrNotGitRepo)

	// In a regular git repo
	mainRepoDir := filepath.Join(tmpDir, "main-repo")
	require.NoError(t, os.MkdirAll(mainRepoDir, 0o755))
	testGitInit(t, mainRepoDir)
	t.Chdir(mainRepoDir)

	root, err := GetRepoRoot()
	require.NoError(t, err)
	require.Equal(t, mainRepoDir, root)

	// In a subdirectory of a git repo
	subDir := filepath.Join(mainRepoDir, "subdir")
	require.NoError(t, os.MkdirAll(subDir, 0o755))
	t.Chdir(subDir)
	root, err = GetRepoRoot()
	require.NoError(t, err)
	require.Equal(t, mainRepoDir, root)

	// In a worktree
	// We need a commit to create a worktree
	runGit(t, mainRepoDir, "config", "user.email", "test@example.com")
	runGit(t, mainRepoDir, "config", "user.name", "Test User")
	require.NoError(t, os.WriteFile(filepath.Join(mainRepoDir, "file.txt"), []byte("hello"), 0o644))
	runGit(t, mainRepoDir, "add", "file.txt")
	runGit(t, mainRepoDir, "commit", "-m", "initial commit")

	worktreeDir := filepath.Join(tmpDir, "worktree")
	runGit(t, mainRepoDir, "worktree", "add", worktreeDir)

	t.Chdir(worktreeDir)
	root, err = GetRepoRoot()
	require.NoError(t, err)
	require.Equal(t, worktreeDir, root)
}

func TestGetRepoRoot_HookContext(t *testing.T) {
	// Create a temp directory for the test
	tmpDir := t.TempDir()
	tmpDir, err := filepath.EvalSymlinks(tmpDir)
	require.NoError(t, err)

	// In a regular git repo
	mainRepoDir := filepath.Join(tmpDir, "main-repo")
	require.NoError(t, os.MkdirAll(mainRepoDir, 0o755))
	testGitInit(t, mainRepoDir)

	// In a subdirectory of a git repo
	subDir := filepath.Join(mainRepoDir, "subdir")
	require.NoError(t, os.MkdirAll(subDir, 0o755))

	// Simulate a hook context where GIT_DIR and GIT_WORK_TREE are set.
	// This usually happens when git runs a hook.
	// If we're in 'subdir' and GIT_DIR is '../.git', we want to make sure
	// GetRepoRoot still returns the correct root.

	// First, let's see what happens without environment variables
	t.Chdir(subDir)
	root, err := GetRepoRoot()
	require.NoError(t, err)
	require.Equal(t, mainRepoDir, root)

	// Now set environment variables as if we're in a hook
	// We'll set GIT_DIR to point to the .git directory relative to the current dir
	t.Setenv("GIT_DIR", filepath.Join("..", ".git"))
	t.Setenv("GIT_WORK_TREE", "..")

	// If we are in 'subdir', but we set GIT_WORK_TREE to '.', we want to see if
	// rev-parse returns the current directory instead of the main repo root.
	t.Setenv("GIT_WORK_TREE", ".")

	root, err = GetRepoRoot()
	require.NoError(t, err)
	// If the fix is NOT applied, GetRepoRoot() might return 'subDir'
	// because git rev-parse --show-toplevel respects GIT_WORK_TREE.
	require.Equal(t, mainRepoDir, root)
}

func testGitInit(t *testing.T, dir string) {
	t.Helper()
	cmd := exec.Command("git", "init", "--template=")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_CONFIG_GLOBAL="+os.DevNull,
		"GIT_CONFIG_SYSTEM="+os.DevNull,
	)
	if err := cmd.Run(); err != nil {
		t.Fatalf("git init failed: %v", err)
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_CONFIG_GLOBAL="+os.DevNull,
		"GIT_CONFIG_SYSTEM="+os.DevNull,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\nOutput: %s", args, err, string(out))
	}
}
