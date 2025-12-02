package config

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestResolveXDGPaths(t *testing.T) {
	paths := ResolveXDGPaths()

	if paths.ConfigHome == "" {
		t.Error("ConfigHome should not be empty")
	}
	if paths.CacheHome == "" {
		t.Error("CacheHome should not be empty")
	}
	if paths.DataHome == "" {
		t.Error("DataHome should not be empty")
	}
}

func TestXDGPaths_ConfigDir(t *testing.T) {
	paths := ResolveXDGPaths()
	configDir := paths.ConfigDir()

	if !filepath.IsAbs(configDir) {
		t.Error("ConfigDir should return an absolute path")
	}
	if filepath.Base(configDir) != AppName {
		t.Errorf("ConfigDir should end with %q, got %q", AppName, filepath.Base(configDir))
	}
}

func TestXDGPaths_CacheDir(t *testing.T) {
	paths := ResolveXDGPaths()
	cacheDir := paths.CacheDir()

	if !filepath.IsAbs(cacheDir) {
		t.Error("CacheDir should return an absolute path")
	}
	if filepath.Base(cacheDir) != AppName {
		t.Errorf("CacheDir should end with %q, got %q", AppName, filepath.Base(cacheDir))
	}
}

func TestXDGConfigHomeOverride(t *testing.T) {
	testDir := "/custom/config/path"
	t.Setenv("XDG_CONFIG_HOME", testDir)

	paths := ResolveXDGPaths()
	if paths.ConfigHome != testDir {
		t.Errorf("Expected ConfigHome to be %q, got %q", testDir, paths.ConfigHome)
	}
}

func TestLoad_Defaults(t *testing.T) {
	// Reset global state
	ResetGlobal()

	// Load with all sources disabled to get pure defaults
	cfg, err := Load(&LoadOptions{
		SkipUserConfig:    true,
		SkipProjectConfig: true,
		SkipEnv:           true,
	})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.GoCmd != DefaultGoCmd {
		t.Errorf("GoCmd = %q, want %q", cfg.GoCmd, DefaultGoCmd)
	}
	if cfg.Verbose != DefaultVerbose {
		t.Errorf("Verbose = %v, want %v", cfg.Verbose, DefaultVerbose)
	}
	if cfg.Debug != DefaultDebug {
		t.Errorf("Debug = %v, want %v", cfg.Debug, DefaultDebug)
	}
	if cfg.HashFast != DefaultHashFast {
		t.Errorf("HashFast = %v, want %v", cfg.HashFast, DefaultHashFast)
	}
	if cfg.IgnoreDefault != DefaultIgnoreDefault {
		t.Errorf("IgnoreDefault = %v, want %v", cfg.IgnoreDefault, DefaultIgnoreDefault)
	}
	if cfg.EnableColor != DefaultEnableColor {
		t.Errorf("EnableColor = %v, want %v", cfg.EnableColor, DefaultEnableColor)
	}
	if cfg.TargetColor != DefaultTargetColor {
		t.Errorf("TargetColor = %q, want %q", cfg.TargetColor, DefaultTargetColor)
	}
}

func TestLoad_EnvironmentVariables(t *testing.T) {
	// Reset global state
	ResetGlobal()

	// Set test values using t.Setenv (auto-cleanup)
	t.Setenv("STAVEFILE_VERBOSE", "true")
	t.Setenv("STAVEFILE_DEBUG", "1")
	t.Setenv("STAVEFILE_GOCMD", "/custom/go")

	cfg, err := Load(&LoadOptions{
		SkipUserConfig:    true,
		SkipProjectConfig: true,
	})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if !cfg.Verbose {
		t.Error("Verbose should be true from STAVEFILE_VERBOSE")
	}
	if !cfg.Debug {
		t.Error("Debug should be true from STAVEFILE_DEBUG")
	}
	if cfg.GoCmd != "/custom/go" {
		t.Errorf("GoCmd = %q, want %q", cfg.GoCmd, "/custom/go")
	}
}

func TestLoad_LegacyEnvironmentVariables(t *testing.T) {
	// Reset global state
	ResetGlobal()

	// t.Setenv clears any existing value, so STAVEFILE_VERBOSE will be unset
	// Set STAVEFILE_ to test legacy fallback
	t.Setenv("STAVEFILE_VERBOSE", "true")

	cfg, err := Load(&LoadOptions{
		SkipUserConfig:    true,
		SkipProjectConfig: true,
	})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if !cfg.Verbose {
		t.Error("Verbose should be true from legacy STAVEFILE_VERBOSE")
	}
}

func TestLoad_ProjectConfig(t *testing.T) {
	// Reset global state
	ResetGlobal()

	// Create temp directory with config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "stave.yaml")
	configContent := `
verbose: true
go_cmd: /project/go
target_color: Red
`
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
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

	if !cfg.Verbose {
		t.Error("Verbose should be true from project config")
	}
	if cfg.GoCmd != "/project/go" {
		t.Errorf("GoCmd = %q, want %q", cfg.GoCmd, "/project/go")
	}
	if cfg.TargetColor != "Red" {
		t.Errorf("TargetColor = %q, want %q", cfg.TargetColor, "Red")
	}
}

func TestConfig_Validate_InvalidColor(t *testing.T) {
	cfg := &Config{
		TargetColor: "InvalidColor",
	}

	result := cfg.Validate()
	if !result.HasErrors() {
		t.Error("Expected validation error for invalid color")
	}
}

func TestConfig_Validate_ValidColors(t *testing.T) {
	validColors := []string{
		"Black", "Red", "Green", "Yellow", "Blue", "Magenta", "Cyan", "White",
		"BrightBlack", "BrightRed", "BrightGreen", "BrightYellow",
		"BrightBlue", "BrightMagenta", "BrightCyan", "BrightWhite",
		"black", "red", "CYAN", // case insensitive
	}

	for _, color := range validColors {
		cfg := &Config{TargetColor: color}
		result := cfg.Validate()
		if result.HasErrors() {
			t.Errorf("Color %q should be valid, got error: %s", color, result.ErrorMessage())
		}
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.GoCmd != DefaultGoCmd {
		t.Errorf("GoCmd = %q, want %q", cfg.GoCmd, DefaultGoCmd)
	}
	if cfg.CacheDir == "" {
		t.Error("CacheDir should not be empty")
	}
}

func TestGlobal_Singleton(t *testing.T) {
	// Reset global state
	ResetGlobal()

	// Get global twice, should be same instance
	cfg1 := Global()
	cfg2 := Global()

	if cfg1 != cfg2 {
		t.Error("Global() should return the same instance")
	}
}

func TestSetGlobal(t *testing.T) {
	// Reset global state
	ResetGlobal()

	customCfg := &Config{
		GoCmd:   "/custom/go",
		Verbose: true,
	}

	SetGlobal(customCfg)

	if Global() != customCfg {
		t.Error("SetGlobal should set the global config")
	}

	// Reset for other tests
	ResetGlobal()
}

func TestValidationResults_WriteWarnings(t *testing.T) {
	result := ValidationResults{
		Warnings: []ValidationWarning{
			{Field: "test", Message: "warning 1"},
			{Field: "test2", Message: "warning 2"},
		},
	}

	var buf bytes.Buffer
	result.WriteWarnings(&buf)

	output := buf.String()
	if output == "" {
		t.Error("WriteWarnings should produce output")
	}
}
