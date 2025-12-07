package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHooksConfig_Get(t *testing.T) {
	t.Parallel()

	hooks := HooksConfig{
		"pre-commit": {
			{Target: "fmt"},
			{Target: "lint", Args: []string{"--fast"}},
		},
		"pre-push": {
			{Target: "test"},
		},
	}

	tests := []struct {
		name     string
		hookName string
		wantLen  int
	}{
		{
			name:     "existing hook",
			hookName: "pre-commit",
			wantLen:  2,
		},
		{
			name:     "another existing hook",
			hookName: "pre-push",
			wantLen:  1,
		},
		{
			name:     "non-existent hook",
			hookName: "post-commit",
			wantLen:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := hooks.Get(tt.hookName)
			if len(got) != tt.wantLen {
				t.Errorf("Get(%q) returned %d targets, want %d", tt.hookName, len(got), tt.wantLen)
			}
		})
	}
}

func TestHooksConfig_Get_Nil(t *testing.T) {
	t.Parallel()

	var hooks HooksConfig
	got := hooks.Get("pre-commit")
	if got != nil {
		t.Errorf("Get on nil HooksConfig should return nil, got %v", got)
	}
}

func TestHooksConfig_HookNames(t *testing.T) {
	t.Parallel()

	hooks := HooksConfig{
		"pre-push":   {{Target: "test"}},
		"pre-commit": {{Target: "fmt"}},
		"commit-msg": {{Target: "validate"}},
	}

	names := hooks.HookNames()

	if len(names) != 3 {
		t.Fatalf("HookNames() returned %d names, want 3", len(names))
	}

	// Should be sorted
	expected := []string{"commit-msg", "pre-commit", "pre-push"}
	for i, name := range names {
		if name != expected[i] {
			t.Errorf("HookNames()[%d] = %q, want %q", i, name, expected[i])
		}
	}
}

func TestHooksConfig_HookNames_Nil(t *testing.T) {
	t.Parallel()

	var hooks HooksConfig
	names := hooks.HookNames()
	if names != nil {
		t.Errorf("HookNames on nil HooksConfig should return nil, got %v", names)
	}
}

func TestIsKnownGitHook(t *testing.T) {
	t.Parallel()

	knownHooks := []string{
		"pre-commit",
		"prepare-commit-msg",
		"commit-msg",
		"post-commit",
		"pre-push",
		"pre-rebase",
		"post-checkout",
		"post-merge",
		"pre-receive",
		"update",
		"post-receive",
		"post-update",
	}

	for _, hook := range knownHooks {
		if !IsKnownGitHook(hook) {
			t.Errorf("IsKnownGitHook(%q) = false, want true", hook)
		}
	}

	unknownHooks := []string{
		"pre-foo",
		"custom-hook",
		"not-a-hook",
	}

	for _, hook := range unknownHooks {
		if IsKnownGitHook(hook) {
			t.Errorf("IsKnownGitHook(%q) = true, want false", hook)
		}
	}
}

func TestValidateHooks_ValidConfig(t *testing.T) {
	t.Parallel()

	hooks := HooksConfig{
		"pre-commit": {
			{Target: "fmt"},
			{Target: "lint", Args: []string{"--fast"}},
		},
		"pre-push": {
			{Target: "test", Args: []string{"./..."}},
		},
		"commit-msg": {
			{Target: "validate-commit-message"},
		},
	}

	result := ValidateHooks(hooks)

	if result.HasErrors() {
		t.Errorf("ValidateHooks returned errors for valid config: %s", result.ErrorMessage())
	}
	if result.HasWarnings() {
		t.Errorf("ValidateHooks returned warnings for valid config with known hooks")
	}
}

func TestValidateHooks_EmptyTargetName(t *testing.T) {
	t.Parallel()

	hooks := HooksConfig{
		"pre-commit": {
			{Target: ""},
		},
	}

	result := ValidateHooks(hooks)

	if !result.HasErrors() {
		t.Error("ValidateHooks should return error for empty target name")
	}

	found := false
	for _, err := range result.Errors {
		if err.Field == "hooks.pre-commit[0].target" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Error should reference the specific target field")
	}
}

func TestValidateHooks_EmptyHookName(t *testing.T) {
	t.Parallel()

	hooks := HooksConfig{
		"": {
			{Target: "fmt"},
		},
	}

	result := ValidateHooks(hooks)

	if !result.HasErrors() {
		t.Error("ValidateHooks should return error for empty hook name")
	}
}

func TestValidateHooks_UnknownHookWarning(t *testing.T) {
	t.Parallel()

	hooks := HooksConfig{
		"custom-hook": {
			{Target: "something"},
		},
	}

	result := ValidateHooks(hooks)

	if result.HasErrors() {
		t.Errorf("ValidateHooks should not return errors for unknown hook: %s", result.ErrorMessage())
	}
	if !result.HasWarnings() {
		t.Error("ValidateHooks should return warning for unknown hook name")
	}
}

func TestValidateHooks_MultipleErrors(t *testing.T) {
	t.Parallel()

	hooks := HooksConfig{
		"pre-commit": {
			{Target: ""},
			{Target: ""},
		},
		"pre-push": {
			{Target: ""},
		},
	}

	result := ValidateHooks(hooks)

	if len(result.Errors) != 3 {
		t.Errorf("ValidateHooks returned %d errors, want 3", len(result.Errors))
	}
}

func TestLoad_WithHooksConfig(t *testing.T) {
	// Reset global state
	ResetGlobal()

	// Create temp directory with config file containing hooks
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "stave.yaml")
	configContent := `
hooks:
  pre-commit:
    - target: fmt
    - target: lint
      args: ["--fast"]
  pre-push:
    - target: test
      args: ["./..."]
  commit-msg:
    - target: validate
      passStdin: true
`
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := Load(&LoadOptions{
		ProjectDir:     tmpDir,
		SkipUserConfig: true,
		SkipEnv:        true,
	})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify hooks were loaded
	if cfg.Hooks == nil {
		t.Fatal("Hooks should not be nil")
	}

	preCommit := cfg.Hooks.Get("pre-commit")
	if len(preCommit) != 2 {
		t.Errorf("pre-commit should have 2 targets, got %d", len(preCommit))
	}
	if preCommit[0].Target != "fmt" {
		t.Errorf("pre-commit[0].Target = %q, want %q", preCommit[0].Target, "fmt")
	}
	if preCommit[1].Target != "lint" {
		t.Errorf("pre-commit[1].Target = %q, want %q", preCommit[1].Target, "lint")
	}
	if len(preCommit[1].Args) != 1 || preCommit[1].Args[0] != "--fast" {
		t.Errorf("pre-commit[1].Args = %v, want [--fast]", preCommit[1].Args)
	}

	prePush := cfg.Hooks.Get("pre-push")
	if len(prePush) != 1 {
		t.Errorf("pre-push should have 1 target, got %d", len(prePush))
	}

	commitMsg := cfg.Hooks.Get("commit-msg")
	if len(commitMsg) != 1 {
		t.Errorf("commit-msg should have 1 target, got %d", len(commitMsg))
	}
}

func TestConfig_Validate_WithInvalidHooks(t *testing.T) {
	cfg := &Config{
		Hooks: HooksConfig{
			"pre-commit": {
				{Target: ""},
			},
		},
	}

	result := cfg.Validate()
	if !result.HasErrors() {
		t.Error("Config.Validate should return errors for invalid hooks config")
	}
}

func TestConfig_Validate_WithUnknownHookWarning(t *testing.T) {
	cfg := &Config{
		TargetColor: "Cyan",
		Hooks: HooksConfig{
			"unknown-hook": {
				{Target: "something"},
			},
		},
	}

	result := cfg.Validate()
	if result.HasErrors() {
		t.Errorf("Config.Validate should not error for unknown hook: %s", result.ErrorMessage())
	}
	if !result.HasWarnings() {
		t.Error("Config.Validate should warn for unknown hook name")
	}
}
