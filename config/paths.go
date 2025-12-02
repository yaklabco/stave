// Package config provides XDG-compliant configuration management for Stave.
package config

import (
	"os"
	"path/filepath"
	"runtime"
)

// AppName is the application name used in configuration paths.
const AppName = "stave"

// ConfigFileName is the name of the user configuration file (without extension).
const ConfigFileName = "config"

// ProjectConfigFileName is the name of the project configuration file (without extension).
const ProjectConfigFileName = "stave"

// Platform constants for OS detection.
const (
	osDarwin  = "darwin"
	osWindows = "windows"
)

// XDGPaths holds the resolved XDG base directory paths for the current platform.
type XDGPaths struct {
	ConfigHome string // User configuration directory
	CacheHome  string // User cache directory
	DataHome   string // User data directory
}

// ResolveXDGPaths returns the XDG base directory paths for the current platform.
// It respects XDG environment variables on Linux and uses platform-appropriate
// defaults on macOS and Windows.
func ResolveXDGPaths() XDGPaths {
	return XDGPaths{
		ConfigHome: resolveConfigHome(),
		CacheHome:  resolveCacheHome(),
		DataHome:   resolveDataHome(),
	}
}

// ConfigDir returns the application-specific configuration directory.
func (p XDGPaths) ConfigDir() string {
	return filepath.Join(p.ConfigHome, AppName)
}

// CacheDir returns the application-specific cache directory.
func (p XDGPaths) CacheDir() string {
	return filepath.Join(p.CacheHome, AppName)
}

// DataDir returns the application-specific data directory.
func (p XDGPaths) DataDir() string {
	return filepath.Join(p.DataHome, AppName)
}

// ConfigFilePath returns the full path to the configuration file.
func (p XDGPaths) ConfigFilePath() string {
	return filepath.Join(p.ConfigDir(), ConfigFileName+".yaml")
}

// resolveConfigHome returns the XDG_CONFIG_HOME equivalent for the current platform.
func resolveConfigHome() string {
	// Check XDG_CONFIG_HOME first (works on all platforms if user sets it)
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return dir
	}

	home := userHomeDir()

	switch runtime.GOOS {
	case osDarwin:
		// macOS: prefer ~/Library/Application Support, but ~/.config is also common
		// We use ~/.config for consistency with other CLI tools
		return filepath.Join(home, ".config")
	case osWindows:
		// Windows: use APPDATA
		if appData := os.Getenv("APPDATA"); appData != "" {
			return appData
		}
		return filepath.Join(home, "AppData", "Roaming")
	default:
		// Linux and other Unix: ~/.config
		return filepath.Join(home, ".config")
	}
}

// resolveCacheHome returns the XDG_CACHE_HOME equivalent for the current platform.
func resolveCacheHome() string {
	// Check XDG_CACHE_HOME first
	if dir := os.Getenv("XDG_CACHE_HOME"); dir != "" {
		return dir
	}

	home := userHomeDir()

	switch runtime.GOOS {
	case osDarwin:
		// macOS: ~/Library/Caches
		return filepath.Join(home, "Library", "Caches")
	case osWindows:
		// Windows: use LOCALAPPDATA
		if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
			return filepath.Join(localAppData, "cache")
		}
		return filepath.Join(home, "AppData", "Local", "cache")
	default:
		// Linux and other Unix: ~/.cache
		return filepath.Join(home, ".cache")
	}
}

// resolveDataHome returns the XDG_DATA_HOME equivalent for the current platform.
func resolveDataHome() string {
	// Check XDG_DATA_HOME first
	if dir := os.Getenv("XDG_DATA_HOME"); dir != "" {
		return dir
	}

	home := userHomeDir()

	switch runtime.GOOS {
	case osDarwin:
		// macOS: ~/Library/Application Support
		return filepath.Join(home, "Library", "Application Support")
	case osWindows:
		// Windows: use LOCALAPPDATA
		if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
			return localAppData
		}
		return filepath.Join(home, "AppData", "Local")
	default:
		// Linux and other Unix: ~/.local/share
		return filepath.Join(home, ".local", "share")
	}
}

// userHomeDir returns the user's home directory.
func userHomeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	// Windows fallback
	if home := os.Getenv("USERPROFILE"); home != "" {
		return home
	}
	// Last resort for Windows
	if drive := os.Getenv("HOMEDRIVE"); drive != "" {
		if path := os.Getenv("HOMEPATH"); path != "" {
			return filepath.Join(drive, path)
		}
	}
	return ""
}
