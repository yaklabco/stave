package config

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

// Config holds all Stave configuration values.
type Config struct {
	// CacheDir is the directory where stave caches compiled binaries.
	// If empty, defaults to the XDG cache directory.
	CacheDir string `mapstructure:"cache_dir"`

	// GoCmd is the Go command to use for compilation.
	GoCmd string `mapstructure:"go_cmd"`

	// Verbose enables verbose output when running targets.
	Verbose bool `mapstructure:"verbose"`

	// Debug enables debug messages.
	Debug bool `mapstructure:"debug"`

	// HashFast uses quick hashing instead of relying on GOCACHE.
	HashFast bool `mapstructure:"hash_fast"`

	// IgnoreDefault ignores the default target in stavefiles.
	IgnoreDefault bool `mapstructure:"ignore_default"`

	// EnableColor enables colored output in terminal.
	EnableColor bool `mapstructure:"enable_color"`

	// TargetColor is the ANSI color name for target names.
	TargetColor string `mapstructure:"target_color"`

	// Hooks defines Git hooks and the Stave targets they should run.
	Hooks HooksConfig `mapstructure:"hooks"`

	// configFile is the path to the config file that was loaded (if any).
	configFile string
}

// ConfigFile returns the path to the configuration file that was loaded,
// or an empty string if no file was loaded.
func (c *Config) ConfigFile() string {
	return c.configFile
}

// globalConfig holds the singleton global configuration.
// These globals are intentional for the singleton pattern.
//
//nolint:gochecknoglobals // singleton pattern requires package-level state
var (
	globalConfig       *Config
	globalConfigLoaded bool
	globalConfigMu     sync.RWMutex
)

// Global returns the global configuration singleton.
// It loads the configuration on first access.
func Global() *Config {
	globalConfigMu.RLock()
	if globalConfigLoaded {
		cfg := globalConfig
		globalConfigMu.RUnlock()
		return cfg
	}
	globalConfigMu.RUnlock()

	// Need to load config
	globalConfigMu.Lock()
	defer globalConfigMu.Unlock()

	// Double-check after acquiring write lock
	if globalConfigLoaded {
		return globalConfig
	}

	cfg, err := Load(nil)
	if err != nil {
		// Fall back to defaults on error
		cfg = &Config{
			GoCmd:       DefaultGoCmd,
			TargetColor: DefaultTargetColor,
		}
	}
	globalConfig = cfg
	globalConfigLoaded = true
	return globalConfig
}

// SetGlobal sets the global configuration.
// This is primarily useful for testing.
func SetGlobal(cfg *Config) {
	globalConfigMu.Lock()
	defer globalConfigMu.Unlock()
	globalConfig = cfg
	globalConfigLoaded = true
}

// ResetGlobal resets the global configuration to be reloaded on next access.
// This is primarily useful for testing.
func ResetGlobal() {
	globalConfigMu.Lock()
	defer globalConfigMu.Unlock()
	globalConfig = nil
	globalConfigLoaded = false
}

// LoadOptions configures how configuration is loaded.
type LoadOptions struct {
	// ProjectDir is the directory to search for project-level config.
	// If empty, the current working directory is used.
	ProjectDir string

	// Stderr is where warnings are written.
	// If nil, os.Stderr is used.
	Stderr io.Writer

	// SkipProjectConfig skips loading project-level configuration.
	SkipProjectConfig bool

	// SkipUserConfig skips loading user-level configuration.
	SkipUserConfig bool

	// SkipEnv skips reading environment variables.
	SkipEnv bool
}

// Load reads configuration from all sources and returns a Config struct.
// Configuration is loaded in the following order (later sources override earlier):
//  1. Defaults
//  2. User config file (~/.config/stave/config.yaml)
//  3. Project config file (./stave.yaml)
//  4. Environment variables (STAVEFILE_* and STAVEFILE_*)
//
// If opts is nil, default options are used.
func Load(opts *LoadOptions) (*Config, error) {
	if opts == nil {
		opts = &LoadOptions{}
	}

	if opts.Stderr == nil {
		opts.Stderr = os.Stderr
	}

	viperInstance := viper.New()

	// Set defaults
	setDefaults(viperInstance)
	viperInstance.SetConfigType("yaml")

	var configFileUsed string

	// Load user config from XDG path (~/.config/stave/config.yaml)
	if !opts.SkipUserConfig {
		paths := ResolveXDGPaths()
		viperInstance.SetConfigName(ConfigFileName)
		viperInstance.AddConfigPath(paths.ConfigDir())

		if err := viperInstance.ReadInConfig(); err != nil {
			var configFileNotFoundError viper.ConfigFileNotFoundError
			if !errors.As(err, &configFileNotFoundError) {
				return nil, fmt.Errorf("failed to read user config file: %w", err)
			}
		} else {
			configFileUsed = viperInstance.ConfigFileUsed()
		}
	}

	// Load project config (./stave.yaml) - merges with/overrides user config
	if !opts.SkipProjectConfig {
		projectDir := opts.ProjectDir
		if projectDir == "" {
			var err error
			projectDir, err = os.Getwd()
			if err != nil {
				return nil, fmt.Errorf("failed to get working directory: %w", err)
			}
		}

		projectConfigPath := filepath.Join(projectDir, ProjectConfigFileName+".yaml")
		if _, err := os.Stat(projectConfigPath); err == nil {
			viperInstance.SetConfigFile(projectConfigPath)
			if err := viperInstance.MergeInConfig(); err != nil {
				return nil, fmt.Errorf("failed to read project config file: %w", err)
			}
			configFileUsed = projectConfigPath
		}
	}

	// Unmarshal into struct
	var cfg Config
	if err := viperInstance.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Apply environment variable overrides (env vars take precedence over config files)
	if !opts.SkipEnv {
		applyEnvironmentOverrides(&cfg)
	}

	// Record which config file was used (project config takes precedence for display)
	cfg.configFile = configFileUsed

	// Apply cache directory default if not set
	if cfg.CacheDir == "" {
		cfg.CacheDir = ResolveXDGPaths().CacheDir()
	}

	// Expand ~ in cache_dir
	if strings.HasPrefix(cfg.CacheDir, "~/") {
		home := userHomeDir()
		cfg.CacheDir = filepath.Join(home, cfg.CacheDir[2:])
	}

	// Validate configuration
	result := cfg.Validate()
	if result.HasWarnings() {
		result.WriteWarnings(opts.Stderr)
	}
	if result.HasErrors() {
		return nil, errors.New(result.ErrorMessage())
	}

	return &cfg, nil
}

// applyEnvironmentOverrides applies environment variable overrides to the config.
// Environment variables take precedence over config file values.
func applyEnvironmentOverrides(cfg *Config) {
	// parseBool interprets a string as a boolean value.
	parseBool := func(v string) bool {
		return v == "1" || v == "true" || v == "TRUE" || v == "True"
	}

	// Apply overrides
	if v := os.Getenv("STAVEFILE_CACHE"); v != "" {
		cfg.CacheDir = v
	}
	if v := os.Getenv("STAVEFILE_GOCMD"); v != "" {
		cfg.GoCmd = v
	}
	if v := os.Getenv("STAVEFILE_VERBOSE"); v != "" {
		cfg.Verbose = parseBool(v)
	}
	if v := os.Getenv("STAVEFILE_DEBUG"); v != "" {
		cfg.Debug = parseBool(v)
	}
	if v := os.Getenv("STAVEFILE_HASHFAST"); v != "" {
		cfg.HashFast = parseBool(v)
	}
	if v := os.Getenv("STAVEFILE_IGNOREDEFAULT"); v != "" {
		cfg.IgnoreDefault = parseBool(v)
	}
	if v := os.Getenv("STAVEFILE_ENABLE_COLOR"); v != "" {
		cfg.EnableColor = parseBool(v)
	}
	if v := os.Getenv("STAVEFILE_TARGET_COLOR"); v != "" {
		cfg.TargetColor = v
	}
}

// DefaultConfig returns a Config with all default values.
func DefaultConfig() *Config {
	return &Config{
		CacheDir:      ResolveXDGPaths().CacheDir(),
		GoCmd:         DefaultGoCmd,
		Verbose:       DefaultVerbose,
		Debug:         DefaultDebug,
		HashFast:      DefaultHashFast,
		IgnoreDefault: DefaultIgnoreDefault,
		EnableColor:   DefaultEnableColor,
		TargetColor:   DefaultTargetColor,
	}
}

// WriteDefaultConfig writes a default configuration file to the user's config directory.
func WriteDefaultConfig() (string, error) {
	paths := ResolveXDGPaths()
	configDir := paths.ConfigDir()

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := paths.ConfigFilePath()

	// Check if file already exists
	if _, err := os.Stat(configPath); err == nil {
		return "", fmt.Errorf("config file already exists: %s", configPath)
	}

	// Write default config with 0600 permissions for security
	content := defaultConfigYAML()
	if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
		return "", fmt.Errorf("failed to write config file: %w", err)
	}

	return configPath, nil
}

// defaultConfigYAML returns the default configuration as YAML.
func defaultConfigYAML() string {
	return `# Stave Configuration
# See https://github.com/yaklabco/stave for documentation

# Directory where stave caches compiled binaries.
# Defaults to XDG cache directory if not set.
# cache_dir: ~/.cache/stave

# Go command to use for compilation.
go_cmd: go

# Enable verbose output when running targets.
verbose: false

# Enable debug messages.
debug: false

# Use quick hashing instead of relying on GOCACHE.
# Faster but may miss transitive dependency changes.
hash_fast: false

# Ignore the default target in stavefiles.
ignore_default: false

# Enable colored output in terminal.
enable_color: false

# ANSI color for target names.
# Options: Black, Red, Green, Yellow, Blue, Magenta, Cyan, White,
#          BrightBlack, BrightRed, BrightGreen, BrightYellow,
#          BrightBlue, BrightMagenta, BrightCyan, BrightWhite
target_color: Cyan
`
}
