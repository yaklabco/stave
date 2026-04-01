package hooks

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFindGitRepo_Worktree(t *testing.T) {
	// Create a temp directory for the main repo
	tmpDir := t.TempDir()
	tmpDir, err := filepath.EvalSymlinks(tmpDir)
	require.NoError(t, err)

	mainRepoDir := filepath.Join(tmpDir, "main-repo")
	require.NoError(t, os.MkdirAll(mainRepoDir, 0o755))

	// Initialize main repo
	testGitInit(t, mainRepoDir)

	// We need at least one commit to create a worktree
	testRunGit(t, mainRepoDir, "config", "user.email", "test@example.com")
	testRunGit(t, mainRepoDir, "config", "user.name", "Test User")
	require.NoError(t, os.WriteFile(filepath.Join(mainRepoDir, "file.txt"), []byte("hello"), 0o644))
	testRunGit(t, mainRepoDir, "add", "file.txt")
	testRunGit(t, mainRepoDir, "commit", "-m", "initial commit")

	// Create a worktree
	worktreeDir := filepath.Join(tmpDir, "worktree")
	testRunGit(t, mainRepoDir, "worktree", "add", worktreeDir)

	// Find repo from worktree
	repo, err := FindGitRepo(worktreeDir)
	require.NoError(t, err)

	// RootDir should be the worktree directory
	require.Equal(t, worktreeDir, repo.RootDir)

	// HooksPath should point to the main repo's hooks directory
	expectedHooksPath := filepath.Join(mainRepoDir, ".git", "hooks")
	require.Equal(t, expectedHooksPath, repo.HooksPath(), "HooksPath should point to main repo's hooks")
}
