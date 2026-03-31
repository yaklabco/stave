package gitutils

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ErrNotGitRepo is returned when the directory is not inside a Git repository.
var ErrNotGitRepo = errors.New("not a git repository")

// GetRepoRoot returns the absolute path to the root of the current Git repository.
// It works correctly for both regular Git clones and Git worktrees.
// If the current directory is not within a Git repository, it returns ErrNotGitRepo.
func GetRepoRoot() (string, error) {
	return GetRepoRootContext(context.Background())
}

// GetRepoRootContext returns the absolute path to the root of the current Git repository with context.
func GetRepoRootContext(ctx context.Context) (string, error) {
	// Use --show-toplevel to find the root of the current working copy.
	// In a worktree, this returns the root of the worktree.
	// In a regular clone, this returns the root of the clone.
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--show-toplevel")

	// Filter out GIT_DIR and GIT_WORK_TREE to ensure correct behavior in hook contexts.
	cmd.Env = filterGitEnv(os.Environ())

	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrNotGitRepo, err)
	}

	rootDir := strings.TrimSpace(string(out))
	if rootDir == "" {
		return "", ErrNotGitRepo
	}

	// Resolve symlinks (important on macOS where /var is a symlink to /private/var)
	resolved, err := filepath.EvalSymlinks(rootDir)
	if err != nil {
		return filepath.Clean(rootDir), nil //nolint:nilerr // This is an intentional fallback.
	}

	return filepath.Clean(resolved), nil
}

// filterGitEnv removes GIT_DIR and GIT_WORK_TREE from the environment.
func filterGitEnv(env []string) []string {
	var filtered []string
	for _, e := range env {
		if strings.HasPrefix(e, "GIT_DIR=") || strings.HasPrefix(e, "GIT_WORK_TREE=") {
			continue
		}
		filtered = append(filtered, e)
	}
	return filtered
}
