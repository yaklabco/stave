package mg_test

import (
	"os"
	"runtime"
	"testing"

	"github.com/yaklabco/stave/mg"
)

func TestVerbose(t *testing.T) {
	tests := []struct {
		name     string
		staveEnv string
		mageEnv  string
		want     bool
	}{
		{"stavefile set true", "1", "", true},
		{"stavefile set false", "0", "", false},
		{"stavefile set true string", "true", "", true},
		{"stavefile set false string", "false", "", false},
		{"magefile fallback true", "", "1", true},
		{"magefile fallback false", "", "0", false},
		{"stavefile takes precedence", "0", "1", false},
		{"neither set", "", "", false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// Clean env
			os.Unsetenv("STAVEFILE_VERBOSE")
			os.Unsetenv("MAGEFILE_VERBOSE")
			t.Cleanup(func() {
				os.Unsetenv("STAVEFILE_VERBOSE")
				os.Unsetenv("MAGEFILE_VERBOSE")
			})

			if tt.staveEnv != "" {
				os.Setenv("STAVEFILE_VERBOSE", tt.staveEnv)
			}
			if tt.mageEnv != "" {
				os.Setenv("MAGEFILE_VERBOSE", tt.mageEnv)
			}

			if got := mg.Verbose(); got != tt.want {
				t.Errorf("Verbose() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDebug(t *testing.T) {
	tests := []struct {
		name     string
		staveEnv string
		mageEnv  string
		want     bool
	}{
		{"stavefile set true", "1", "", true},
		{"stavefile set false", "0", "", false},
		{"stavefile set true string", "true", "", true},
		{"magefile fallback true", "", "1", true},
		{"magefile fallback false", "", "0", false},
		{"stavefile takes precedence", "0", "1", false},
		{"neither set", "", "", false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// Clean env
			os.Unsetenv("STAVEFILE_DEBUG")
			os.Unsetenv("MAGEFILE_DEBUG")
			t.Cleanup(func() {
				os.Unsetenv("STAVEFILE_DEBUG")
				os.Unsetenv("MAGEFILE_DEBUG")
			})

			if tt.staveEnv != "" {
				os.Setenv("STAVEFILE_DEBUG", tt.staveEnv)
			}
			if tt.mageEnv != "" {
				os.Setenv("MAGEFILE_DEBUG", tt.mageEnv)
			}

			if got := mg.Debug(); got != tt.want {
				t.Errorf("Debug() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGoCmd(t *testing.T) {
	tests := []struct {
		name     string
		staveEnv string
		mageEnv  string
		want     string
	}{
		{"stavefile set custom", "custom-go", "", "custom-go"},
		{"magefile fallback", "", "legacy-go", "legacy-go"},
		{"stavefile takes precedence", "custom-go", "legacy-go", "custom-go"},
		{"neither set defaults to go", "", "", "go"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// Clean env
			os.Unsetenv("STAVEFILE_GOCMD")
			os.Unsetenv("MAGEFILE_GOCMD")
			t.Cleanup(func() {
				os.Unsetenv("STAVEFILE_GOCMD")
				os.Unsetenv("MAGEFILE_GOCMD")
			})

			if tt.staveEnv != "" {
				os.Setenv("STAVEFILE_GOCMD", tt.staveEnv)
			}
			if tt.mageEnv != "" {
				os.Setenv("MAGEFILE_GOCMD", tt.mageEnv)
			}

			if got := mg.GoCmd(); got != tt.want {
				t.Errorf("GoCmd() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHashFast(t *testing.T) {
	tests := []struct {
		name     string
		staveEnv string
		mageEnv  string
		want     bool
	}{
		{"stavefile true", "true", "", true},
		{"stavefile false", "false", "", false},
		{"stavefile 1", "1", "", true},
		{"stavefile 0", "0", "", false},
		{"magefile fallback", "", "1", true},
		{"stavefile precedence", "0", "1", false},
		{"neither set", "", "", false},
		{"stavefile empty string", "", "", false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// Clean env
			os.Unsetenv("STAVEFILE_HASHFAST")
			os.Unsetenv("MAGEFILE_HASHFAST")
			t.Cleanup(func() {
				os.Unsetenv("STAVEFILE_HASHFAST")
				os.Unsetenv("MAGEFILE_HASHFAST")
			})

			if tt.staveEnv != "" {
				os.Setenv("STAVEFILE_HASHFAST", tt.staveEnv)
			}
			if tt.mageEnv != "" {
				os.Setenv("MAGEFILE_HASHFAST", tt.mageEnv)
			}

			if got := mg.HashFast(); got != tt.want {
				t.Errorf("HashFast() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIgnoreDefault(t *testing.T) {
	tests := []struct {
		name     string
		staveEnv string
		mageEnv  string
		want     bool
	}{
		{"stavefile true", "true", "", true},
		{"stavefile false", "false", "", false},
		{"stavefile 1", "1", "", true},
		{"stavefile 0", "0", "", false},
		{"magefile fallback", "", "1", true},
		{"stavefile precedence", "0", "1", false},
		{"neither set", "", "", false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// Clean env
			os.Unsetenv("STAVEFILE_IGNOREDEFAULT")
			os.Unsetenv("MAGEFILE_IGNOREDEFAULT")
			t.Cleanup(func() {
				os.Unsetenv("STAVEFILE_IGNOREDEFAULT")
				os.Unsetenv("MAGEFILE_IGNOREDEFAULT")
			})

			if tt.staveEnv != "" {
				os.Setenv("STAVEFILE_IGNOREDEFAULT", tt.staveEnv)
			}
			if tt.mageEnv != "" {
				os.Setenv("MAGEFILE_IGNOREDEFAULT", tt.mageEnv)
			}

			if got := mg.IgnoreDefault(); got != tt.want {
				t.Errorf("IgnoreDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCacheDir(t *testing.T) {
	tests := []struct {
		name     string
		staveEnv string
		mageEnv  string
		wantPath string // substring that should be in the path
	}{
		{"stavefile set", "/tmp/stave-cache", "", "/tmp/stave-cache"},
		{"magefile fallback", "", "/tmp/mage-cache", "/tmp/mage-cache"},
		{"stavefile precedence", "/tmp/stave-cache", "/tmp/mage-cache", "/tmp/stave-cache"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// Clean env
			os.Unsetenv("STAVEFILE_CACHE")
			os.Unsetenv("MAGEFILE_CACHE")
			t.Cleanup(func() {
				os.Unsetenv("STAVEFILE_CACHE")
				os.Unsetenv("MAGEFILE_CACHE")
			})

			if tt.staveEnv != "" {
				os.Setenv("STAVEFILE_CACHE", tt.staveEnv)
			}
			if tt.mageEnv != "" {
				os.Setenv("MAGEFILE_CACHE", tt.mageEnv)
			}

			got := mg.CacheDir()
			if got != tt.wantPath {
				t.Errorf("CacheDir() = %q, want %q", got, tt.wantPath)
			}
		})
	}

	// Test default path logic
	t.Run("default path", func(t *testing.T) {
		os.Unsetenv("STAVEFILE_CACHE")
		os.Unsetenv("MAGEFILE_CACHE")
		t.Cleanup(func() {
			os.Unsetenv("STAVEFILE_CACHE")
			os.Unsetenv("MAGEFILE_CACHE")
		})

		got := mg.CacheDir()
		if runtime.GOOS == "windows" {
			// Should contain "stavefile"
			if len(got) == 0 || got[len(got)-9:] != "stavefile" {
				t.Errorf("CacheDir() on Windows = %q, want to end with 'stavefile'", got)
			}
		} else {
			// Should contain ".stavefile"
			if len(got) == 0 || got[len(got)-10:] != ".stavefile" {
				t.Errorf("CacheDir() on Unix = %q, want to end with '.stavefile'", got)
			}
		}
	})
}

func TestEnableColor(t *testing.T) {
	tests := []struct {
		name     string
		staveEnv string
		mageEnv  string
		want     bool
	}{
		{"stavefile true", "true", "", true},
		{"stavefile false", "false", "", false},
		{"stavefile 1", "1", "", true},
		{"stavefile 0", "0", "", false},
		{"magefile fallback", "", "1", true},
		{"stavefile precedence", "0", "1", false},
		{"neither set", "", "", false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// Clean env
			os.Unsetenv("STAVEFILE_ENABLE_COLOR")
			os.Unsetenv("MAGEFILE_ENABLE_COLOR")
			t.Cleanup(func() {
				os.Unsetenv("STAVEFILE_ENABLE_COLOR")
				os.Unsetenv("MAGEFILE_ENABLE_COLOR")
			})

			if tt.staveEnv != "" {
				os.Setenv("STAVEFILE_ENABLE_COLOR", tt.staveEnv)
			}
			if tt.mageEnv != "" {
				os.Setenv("MAGEFILE_ENABLE_COLOR", tt.mageEnv)
			}

			if got := mg.EnableColor(); got != tt.want {
				t.Errorf("EnableColor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTargetColorFallback(t *testing.T) {
	t.Run("stavefile takes precedence over magefile", func(t *testing.T) {
		os.Setenv("STAVEFILE_TARGET_COLOR", "Red")
		os.Setenv("MAGEFILE_TARGET_COLOR", "Green")
		t.Cleanup(func() {
			os.Unsetenv("STAVEFILE_TARGET_COLOR")
			os.Unsetenv("MAGEFILE_TARGET_COLOR")
		})

		got := mg.TargetColor()
		// Should be the red ANSI code
		if got != "\u001b[31m" {
			t.Errorf("TargetColor() with STAVEFILE_TARGET_COLOR=Red = %q, want red ANSI code", got)
		}
	})

	t.Run("magefile fallback", func(t *testing.T) {
		os.Unsetenv("STAVEFILE_TARGET_COLOR")
		os.Setenv("MAGEFILE_TARGET_COLOR", "Yellow")
		t.Cleanup(func() {
			os.Unsetenv("MAGEFILE_TARGET_COLOR")
		})

		got := mg.TargetColor()
		// Should be the yellow ANSI code
		if got != "\u001b[33m" {
			t.Errorf("TargetColor() with MAGEFILE_TARGET_COLOR=Yellow = %q, want yellow ANSI code", got)
		}
	})
}

