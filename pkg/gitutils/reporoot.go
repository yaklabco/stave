package gitutils

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/yaklabco/stave/internal/hooks"
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
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getting current working directory: %w", err)
	}

	repo, err := hooks.FindGitRepoContext(ctx, cwd)
	if err != nil {
		return "", fmt.Errorf("finding git repo: %w", err)
	}

	return repo.RootDir, nil
}

func IsWorkTree(path string) bool {
	dotGitPath := filepath.Join(path, ".git")
	fileInfo, err := os.Stat(dotGitPath)

	return err == nil && !fileInfo.IsDir()
}
