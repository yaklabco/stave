package hooks

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// testGitInit initializes a git repository in the given directory.
// It uses --template= to avoid inheriting hooks from user git templates,
// ensuring test isolation regardless of the user's git configuration.
// It also creates the hooks directory since --template= skips creating it.
func testGitInit(t *testing.T, dir string) {
	t.Helper()
	cmd := exec.Command("git", "init", "--template=")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("git init failed: %v", err)
	}
	// Create hooks directory since --template= skips it
	hooksDir := filepath.Join(dir, ".git", "hooks")
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		t.Fatalf("mkdir hooks failed: %v", err)
	}
}

func TestFindGitRepo_Valid(t *testing.T) {
	t.Parallel()

	// Create a temp directory with a git repo
	tmpDir := t.TempDir()

	// Resolve symlinks (macOS /var -> /private/var)
	tmpDir, err := filepath.EvalSymlinks(tmpDir)
	if err != nil {
		t.Fatalf("EvalSymlinks failed: %v", err)
	}

	testGitInit(t, tmpDir)

	repo, err := FindGitRepo(tmpDir)
	if err != nil {
		t.Fatalf("FindGitRepo() error = %v", err)
	}

	if repo.RootDir != tmpDir {
		t.Errorf("RootDir = %q, want %q", repo.RootDir, tmpDir)
	}

	expectedGitDir := filepath.Join(tmpDir, ".git")
	if repo.GitDir != expectedGitDir {
		t.Errorf("GitDir = %q, want %q", repo.GitDir, expectedGitDir)
	}
}

func TestFindGitRepo_NotARepo(t *testing.T) {
	t.Parallel()

	// Create a temp directory without a git repo
	tmpDir := t.TempDir()

	_, err := FindGitRepo(tmpDir)
	if !errors.Is(err, ErrNotGitRepo) {
		t.Errorf("FindGitRepo() error = %v, want %v", err, ErrNotGitRepo)
	}
	// Error should include the path for debugging context
	if err != nil && err.Error() == ErrNotGitRepo.Error() {
		t.Error("FindGitRepo() error should include path context")
	}
}

func TestFindGitRepo_Subdirectory(t *testing.T) {
	t.Parallel()

	// Create a temp directory with a git repo
	tmpDir := t.TempDir()

	// Resolve symlinks (macOS /var -> /private/var)
	tmpDir, err := filepath.EvalSymlinks(tmpDir)
	if err != nil {
		t.Fatalf("EvalSymlinks failed: %v", err)
	}

	testGitInit(t, tmpDir)

	// Create a subdirectory
	subDir := filepath.Join(tmpDir, "subdir", "nested")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	// Find repo from subdirectory
	repo, err := FindGitRepo(subDir)
	if err != nil {
		t.Fatalf("FindGitRepo() error = %v", err)
	}

	if repo.RootDir != tmpDir {
		t.Errorf("RootDir = %q, want %q", repo.RootDir, tmpDir)
	}
}

func TestGitRepo_HooksPath_Default(t *testing.T) {
	t.Parallel()

	// Create a temp directory with a git repo
	tmpDir := t.TempDir()

	// Resolve symlinks (macOS /var -> /private/var)
	tmpDir, err := filepath.EvalSymlinks(tmpDir)
	if err != nil {
		t.Fatalf("EvalSymlinks failed: %v", err)
	}

	testGitInit(t, tmpDir)

	repo, err := FindGitRepo(tmpDir)
	if err != nil {
		t.Fatalf("FindGitRepo() error = %v", err)
	}

	expectedHooksPath := filepath.Join(tmpDir, ".git", "hooks")
	if repo.HooksPath() != expectedHooksPath {
		t.Errorf("HooksPath() = %q, want %q", repo.HooksPath(), expectedHooksPath)
	}

	if repo.HasCustomHooksPath() {
		t.Error("HasCustomHooksPath() = true, want false")
	}
}

func TestGitRepo_HooksPath_CustomPath(t *testing.T) {
	t.Parallel()

	// Create a temp directory with a git repo
	tmpDir := t.TempDir()

	// Resolve symlinks (macOS /var -> /private/var)
	tmpDir, err := filepath.EvalSymlinks(tmpDir)
	if err != nil {
		t.Fatalf("EvalSymlinks failed: %v", err)
	}

	testGitInit(t, tmpDir)

	// Set custom hooks path
	customPath := ".githooks"
	cmd := exec.Command("git", "config", "core.hooksPath", customPath)
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("git config failed: %v", err)
	}

	repo, err := FindGitRepo(tmpDir)
	if err != nil {
		t.Fatalf("FindGitRepo() error = %v", err)
	}

	expectedHooksPath := filepath.Join(tmpDir, customPath)
	if repo.HooksPath() != expectedHooksPath {
		t.Errorf("HooksPath() = %q, want %q", repo.HooksPath(), expectedHooksPath)
	}

	if !repo.HasCustomHooksPath() {
		t.Error("HasCustomHooksPath() = false, want true")
	}
}

func TestGitRepo_HooksPath_AbsoluteCustomPath(t *testing.T) {
	t.Parallel()

	// Create a temp directory with a git repo
	tmpDir := t.TempDir()

	// Resolve symlinks (macOS /var -> /private/var)
	tmpDir, err := filepath.EvalSymlinks(tmpDir)
	if err != nil {
		t.Fatalf("EvalSymlinks failed: %v", err)
	}

	testGitInit(t, tmpDir)

	// Set absolute custom hooks path
	customPath := filepath.Join(tmpDir, "custom-hooks")
	cmd := exec.Command("git", "config", "core.hooksPath", customPath)
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("git config failed: %v", err)
	}

	repo, err := FindGitRepo(tmpDir)
	if err != nil {
		t.Fatalf("FindGitRepo() error = %v", err)
	}

	if repo.HooksPath() != customPath {
		t.Errorf("HooksPath() = %q, want %q", repo.HooksPath(), customPath)
	}
}

func TestGitRepo_EnsureHooksDir(t *testing.T) {
	t.Parallel()

	// Create a temp directory with a git repo
	tmpDir := t.TempDir()

	// Resolve symlinks (macOS /var -> /private/var)
	tmpDir, err := filepath.EvalSymlinks(tmpDir)
	if err != nil {
		t.Fatalf("EvalSymlinks failed: %v", err)
	}

	testGitInit(t, tmpDir)

	// Remove the hooks directory if it exists
	hooksDir := filepath.Join(tmpDir, ".git", "hooks")
	_ = os.RemoveAll(hooksDir)

	repo, err := FindGitRepo(tmpDir)
	if err != nil {
		t.Fatalf("FindGitRepo() error = %v", err)
	}

	// Ensure hooks dir is created
	if err := repo.EnsureHooksDir(); err != nil {
		t.Fatalf("EnsureHooksDir() error = %v", err)
	}

	// Verify directory exists
	info, err := os.Stat(hooksDir)
	if err != nil {
		t.Fatalf("hooks dir not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("hooks path is not a directory")
	}
}

func TestGitRepo_HookPath(t *testing.T) {
	t.Parallel()

	// Create a temp directory with a git repo
	tmpDir := t.TempDir()

	// Resolve symlinks (macOS /var -> /private/var)
	tmpDir, err := filepath.EvalSymlinks(tmpDir)
	if err != nil {
		t.Fatalf("EvalSymlinks failed: %v", err)
	}

	testGitInit(t, tmpDir)

	repo, err := FindGitRepo(tmpDir)
	if err != nil {
		t.Fatalf("FindGitRepo() error = %v", err)
	}

	hookPath := repo.HookPath("pre-commit")
	expected := filepath.Join(tmpDir, ".git", "hooks", "pre-commit")
	if hookPath != expected {
		t.Errorf("HookPath(pre-commit) = %q, want %q", hookPath, expected)
	}
}

func TestFindGitRepo_EmptyDir(t *testing.T) {
	// Test that empty dir uses current working directory
	// This test doesn't use t.Parallel() because it relies on CWD

	// Get current repo (we're in a git repo)
	repo, err := FindGitRepo("")
	if err != nil {
		t.Fatalf("FindGitRepo('') error = %v", err)
	}

	if repo.RootDir == "" {
		t.Error("RootDir should not be empty")
	}
}
