package hooks

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// testGitInit initializes an isolated git repository in the given directory.
// It uses --template= to avoid inheriting hooks from user git templates and
// sets GIT_CONFIG_GLOBAL/SYSTEM to /dev/null so user git config (e.g.
// core.hooksPath) doesn't leak into the test repo. It also writes a local
// git config that unsets core.hooksPath for any subsequent git commands that
// run without the env overrides (e.g. in production code under test).
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

	// Ensure core.hooksPath is not inherited from global config for
	// subsequent git commands run by the code under test.
	unset := exec.Command("git", "config", "--local", "core.hooksPath", "")
	unset.Dir = dir
	_ = unset.Run() //nolint:errcheck // non-zero exit if key absent is fine

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
	_, err := FindGitRepo("")
	require.Error(t, err)
}
