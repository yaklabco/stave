package config

import (
	"os"
	"runtime"
	"testing"
)

func TestResolveConfigHome_WithXDGEnv(t *testing.T) {
	// Save original
	orig := os.Getenv("XDG_CONFIG_HOME")
	defer func() {
		if orig == "" {
			os.Unsetenv("XDG_CONFIG_HOME")
		} else {
			os.Setenv("XDG_CONFIG_HOME", orig)
		}
	}()

	os.Setenv("XDG_CONFIG_HOME", "/custom/xdg/config")

	result := resolveConfigHome()
	if result != "/custom/xdg/config" {
		t.Errorf("resolveConfigHome() = %q, want %q", result, "/custom/xdg/config")
	}
}

func TestResolveCacheHome_WithXDGEnv(t *testing.T) {
	// Save original
	orig := os.Getenv("XDG_CACHE_HOME")
	defer func() {
		if orig == "" {
			os.Unsetenv("XDG_CACHE_HOME")
		} else {
			os.Setenv("XDG_CACHE_HOME", orig)
		}
	}()

	os.Setenv("XDG_CACHE_HOME", "/custom/xdg/cache")

	result := resolveCacheHome()
	if result != "/custom/xdg/cache" {
		t.Errorf("resolveCacheHome() = %q, want %q", result, "/custom/xdg/cache")
	}
}

func TestResolveDataHome_WithXDGEnv(t *testing.T) {
	// Save original
	orig := os.Getenv("XDG_DATA_HOME")
	defer func() {
		if orig == "" {
			os.Unsetenv("XDG_DATA_HOME")
		} else {
			os.Setenv("XDG_DATA_HOME", orig)
		}
	}()

	os.Setenv("XDG_DATA_HOME", "/custom/xdg/data")

	result := resolveDataHome()
	if result != "/custom/xdg/data" {
		t.Errorf("resolveDataHome() = %q, want %q", result, "/custom/xdg/data")
	}
}

func TestUserHomeDir(t *testing.T) {
	home := userHomeDir()
	if home == "" {
		t.Skip("Could not determine home directory")
	}

	// Should be an absolute path
	if home[0] != '/' && (runtime.GOOS != osWindows || (len(home) < 2 || home[1] != ':')) {
		t.Errorf("userHomeDir() = %q, should be absolute path", home)
	}
}

func TestXDGPaths_Methods(t *testing.T) {
	paths := XDGPaths{
		ConfigHome: "/config",
		CacheHome:  "/cache",
		DataHome:   "/data",
	}

	tests := []struct {
		name     string
		method   func() string
		expected string
	}{
		{"ConfigDir", paths.ConfigDir, "/config/stave"},
		{"CacheDir", paths.CacheDir, "/cache/stave"},
		{"DataDir", paths.DataDir, "/data/stave"},
		{"ConfigFilePath", paths.ConfigFilePath, "/config/stave/config.yaml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.method(); got != tt.expected {
				t.Errorf("%s() = %q, want %q", tt.name, got, tt.expected)
			}
		})
	}
}
