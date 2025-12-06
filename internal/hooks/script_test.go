package hooks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateScript_ContainsMarker(t *testing.T) {
	t.Parallel()

	script := GenerateScript(ScriptParams{HookName: "pre-commit"})

	if !strings.Contains(script, StaveMarker) {
		t.Error("Generated script should contain the Stave marker")
	}
}

func TestGenerateScript_ContainsHookName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		hookName string
	}{
		{"pre-commit"},
		{"pre-push"},
		{"commit-msg"},
		{"prepare-commit-msg"},
	}

	for _, testCase := range tests {
		t.Run(testCase.hookName, func(t *testing.T) {
			t.Parallel()
			script := GenerateScript(ScriptParams{HookName: testCase.hookName})

			// Should appear in both the stave hooks run command and the error message
			if !strings.Contains(script, "stave hooks run "+testCase.hookName) {
				t.Errorf("Generated script should contain 'stave hooks run %s'", testCase.hookName)
			}
			if !strings.Contains(script, "skipping "+testCase.hookName+" hook") {
				t.Errorf("Generated script should contain 'skipping %s hook'", testCase.hookName)
			}
		})
	}
}

func TestGenerateScript_IsPOSIXCompatible(t *testing.T) {
	t.Parallel()

	script := GenerateScript(ScriptParams{HookName: "pre-commit"})

	// Check for shebang
	if !strings.HasPrefix(script, "#!/bin/sh\n") {
		t.Error("Script should start with #!/bin/sh")
	}

	// Check for common bash-isms that would break POSIX sh
	bashisms := []string{
		"[[",        // bash test syntax
		"]]",        // bash test syntax
		"function ", // bash function keyword
		"${var:=",   // bash default value with assignment (we use ${var:-} which is POSIX)
		"source ",   // bash source (we use . which is POSIX)
		"&>",        // bash redirect
		"<<<",       // bash here-string
		"((",        // bash arithmetic
		"))",        // bash arithmetic
		"declare ",  // bash declare
		"local ",    // bash local (not POSIX, but widely supported - we don't use it)
		"$'",        // bash ANSI-C quoting
	}

	for _, bashism := range bashisms {
		if strings.Contains(script, bashism) {
			t.Errorf("Script contains bash-ism %q which may not work in POSIX sh", bashism)
		}
	}
}

func TestGenerateScript_HasEnvControls(t *testing.T) {
	t.Parallel()

	script := GenerateScript(ScriptParams{HookName: "pre-commit"})

	// Check for STAVE_HOOKS=0 handling
	if !strings.Contains(script, `STAVE_HOOKS`) {
		t.Error("Script should handle STAVE_HOOKS environment variable")
	}

	// Check for debug mode
	if !strings.Contains(script, "debug") && !strings.Contains(script, "set -x") {
		t.Error("Script should support debug mode with set -x")
	}
}

func TestGenerateScript_HasUserInit(t *testing.T) {
	t.Parallel()

	script := GenerateScript(ScriptParams{HookName: "pre-commit"})

	// Check for user init script sourcing
	if !strings.Contains(script, "init.sh") {
		t.Error("Script should source user init script")
	}
	if !strings.Contains(script, "XDG_CONFIG_HOME") {
		t.Error("Script should respect XDG_CONFIG_HOME")
	}
}

func TestGenerateScript_UsesExec(t *testing.T) {
	t.Parallel()

	script := GenerateScript(ScriptParams{HookName: "pre-commit"})

	// Check that exec is used to replace the shell process
	if !strings.Contains(script, "exec stave hooks run") {
		t.Error("Script should use exec to replace shell process")
	}
}

func TestIsStaveManaged_True(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	hookPath := filepath.Join(tmpDir, "pre-commit")

	// Write a Stave-managed hook
	content := GenerateScript(ScriptParams{HookName: "pre-commit"})

	if err := os.WriteFile(hookPath, []byte(content), 0o755); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	managed, err := IsStaveManaged(hookPath)
	if err != nil {
		t.Fatalf("IsStaveManaged() error = %v", err)
	}
	if !managed {
		t.Error("IsStaveManaged() = false, want true")
	}
}

func TestIsStaveManaged_False(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	hookPath := filepath.Join(tmpDir, "pre-commit")

	// Write a non-Stave hook
	content := `#!/bin/sh
# Some other hook manager
echo "Hello"
`
	if err := os.WriteFile(hookPath, []byte(content), 0o755); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	managed, err := IsStaveManaged(hookPath)
	if err != nil {
		t.Fatalf("IsStaveManaged() error = %v", err)
	}
	if managed {
		t.Error("IsStaveManaged() = true, want false")
	}
}

func TestIsStaveManaged_FileNotExist(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	hookPath := filepath.Join(tmpDir, "pre-commit")

	managed, err := IsStaveManaged(hookPath)
	if err != nil {
		t.Fatalf("IsStaveManaged() error = %v", err)
	}
	if managed {
		t.Error("IsStaveManaged() = true for non-existent file, want false")
	}
}

func TestWriteHookScript(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	hookPath := filepath.Join(tmpDir, "pre-commit")

	err := WriteHookScript(hookPath, ScriptParams{HookName: "pre-commit"})
	if err != nil {
		t.Fatalf("WriteHookScript() error = %v", err)
	}

	// Verify file exists and is executable
	info, err := os.Stat(hookPath)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}

	// Check executable bit (on Unix)
	if info.Mode()&0o100 == 0 {
		t.Error("Hook script should be executable")
	}

	// Verify content
	content, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !strings.Contains(string(content), "pre-commit") {
		t.Error("Hook script should contain hook name")
	}
}

func TestRemoveHookScript_StaveManaged(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	hookPath := filepath.Join(tmpDir, "pre-commit")

	// Write a Stave-managed hook
	if err := WriteHookScript(hookPath, ScriptParams{HookName: "pre-commit"}); err != nil {
		t.Fatalf("WriteHookScript() error = %v", err)
	}

	removed, err := RemoveHookScript(hookPath)
	if err != nil {
		t.Fatalf("RemoveHookScript() error = %v", err)
	}
	if !removed {
		t.Error("RemoveHookScript() = false, want true")
	}

	// Verify file is gone
	if _, err := os.Stat(hookPath); !os.IsNotExist(err) {
		t.Error("Hook file should have been removed")
	}
}

func TestRemoveHookScript_NotStaveManaged(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	hookPath := filepath.Join(tmpDir, "pre-commit")

	// Write a non-Stave hook
	content := `#!/bin/sh
echo "Hello"
`
	if err := os.WriteFile(hookPath, []byte(content), 0o755); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	removed, err := RemoveHookScript(hookPath)
	if err != nil {
		t.Fatalf("RemoveHookScript() error = %v", err)
	}
	if removed {
		t.Error("RemoveHookScript() = true, want false for non-Stave hook")
	}

	// Verify file still exists
	if _, err := os.Stat(hookPath); os.IsNotExist(err) {
		t.Error("Non-Stave hook file should not have been removed")
	}
}

func TestRemoveHookScript_NotExist(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	hookPath := filepath.Join(tmpDir, "pre-commit")

	removed, err := RemoveHookScript(hookPath)
	if err != nil {
		t.Fatalf("RemoveHookScript() error = %v", err)
	}
	if removed {
		t.Error("RemoveHookScript() = true for non-existent file, want false")
	}
}
