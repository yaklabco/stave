package hooks

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

// GitRepo holds information about a Git repository.
type GitRepo struct {
	// RootDir is the absolute path to the repository root.
	RootDir string

	// GitDir is the absolute path to the .git directory (or gitdir for worktrees).
	GitDir string

	// customHooksPath is the value of core.hooksPath if set, empty otherwise.
	customHooksPath string
}

// FindGitRepo locates the Git repository from the given directory.
// If dir is empty, the current working directory is used.
func FindGitRepo(dir string) (*GitRepo, error) {
	return FindGitRepoContext(context.Background(), dir)
}

// FindGitRepoContext locates the Git repository from the given directory with context.
// If dir is empty, the current working directory is used.
func FindGitRepoContext(ctx context.Context, dir string) (*GitRepo, error) {
	if dir == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("getting working directory: %w", err)
		}
	}

	// Make dir absolute
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("resolving absolute path: %w", err)
	}

	// Get repository root
	rootDir, err := gitOutput(ctx, absDir, "rev-parse", "--show-toplevel")
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrNotGitRepo, absDir)
	}

	// Get git directory
	gitDir, err := gitOutput(ctx, absDir, "rev-parse", "--git-dir")
	if err != nil {
		return nil, fmt.Errorf("finding git directory: %w", err)
	}

	// Make gitDir absolute if it isn't already
	if !filepath.IsAbs(gitDir) {
		gitDir = filepath.Join(absDir, gitDir)
	}

	// Resolve symlinks to get canonical paths (important on macOS where
	// /var is a symlink to /private/var)
	rootDir, err = filepath.EvalSymlinks(rootDir)
	if err != nil {
		return nil, fmt.Errorf("resolving root dir symlinks: %w", err)
	}
	gitDir, err = filepath.EvalSymlinks(gitDir)
	if err != nil {
		return nil, fmt.Errorf("resolving git dir symlinks: %w", err)
	}

	// Clean paths
	rootDir = filepath.Clean(rootDir)
	gitDir = filepath.Clean(gitDir)

	// Check for custom hooks path (ignoring error since empty is valid)
	customHooksPath, err := gitOutput(ctx, absDir, "config", "--get", "core.hooksPath")
	if err != nil {
		customHooksPath = ""
	}

	return &GitRepo{
		RootDir:         rootDir,
		GitDir:          gitDir,
		customHooksPath: customHooksPath,
	}, nil
}

// HooksPath returns the effective hooks directory for this repository.
// If core.hooksPath is configured, that path is returned (resolved relative to RootDir if relative).
// Otherwise, returns <GitDir>/hooks.
func (r *GitRepo) HooksPath() string {
	if r.customHooksPath != "" {
		// If custom path is relative, resolve it relative to repo root
		if !filepath.IsAbs(r.customHooksPath) {
			return filepath.Join(r.RootDir, r.customHooksPath)
		}
		return r.customHooksPath
	}
	return filepath.Join(r.GitDir, "hooks")
}

// HasCustomHooksPath returns true if core.hooksPath is configured.
func (r *GitRepo) HasCustomHooksPath() bool {
	return r.customHooksPath != ""
}

// gitOutput runs a git command and returns the trimmed stdout.
// Returns an error if the command fails.
func gitOutput(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}

// dirPerm is the permission mode for directories.
const dirPerm = 0o755

// EnsureHooksDir creates the hooks directory if it doesn't exist.
func (r *GitRepo) EnsureHooksDir() error {
	hooksPath := r.HooksPath()
	return os.MkdirAll(hooksPath, dirPerm)
}

// HookPath returns the full path to a specific hook file.
func (r *GitRepo) HookPath(hookName string) string {
	return filepath.Join(r.HooksPath(), hookName)
}
