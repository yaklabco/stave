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
	"github.com/yaklabco/stave/pkg/env"
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

// SetConfigFile sets the path to the configuration file.
// This is primarily for testing.
func (c *Config) SetConfigFile(path string) {
	c.configFile = path
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
	opts = normalizeLoadOptions(opts)

	viperInstance := viper.New()
	setDefaults(viperInstance)
	viperInstance.SetConfigType("yaml")

	configFileUsed, err := loadConfigFiles(viperInstance, opts)
	if err != nil {
		return nil, err
	}

	cfg, err := unmarshalConfig(viperInstance, opts, configFileUsed)
	if err != nil {
		return nil, err
	}

	return validateAndFinalize(cfg, opts)
}

// normalizeLoadOptions ensures opts is non-nil and has defaults applied.
func normalizeLoadOptions(opts *LoadOptions) *LoadOptions {
	if opts == nil {
		opts = &LoadOptions{}
	}
	if opts.Stderr == nil {
		opts.Stderr = os.Stderr
	}
	return opts
}

// loadConfigFiles loads user and project config files into viper.
// Returns the path to the most recently loaded config file.
func loadConfigFiles(viperInstance *viper.Viper, opts *LoadOptions) (string, error) {
	var configFileUsed string

	if !opts.SkipUserConfig {
		usedFile, err := loadUserConfig(viperInstance)
		if err != nil {
			return "", err
		}
		if usedFile != "" {
			configFileUsed = usedFile
		}
	}

	if !opts.SkipProjectConfig {
		usedFile, err := loadProjectConfig(viperInstance, opts.ProjectDir)
		if err != nil {
			return "", err
		}
		if usedFile != "" {
			configFileUsed = usedFile
		}
	}

	return configFileUsed, nil
}

// loadUserConfig loads user config from XDG path (~/.config/stave/config.yaml).
func loadUserConfig(viperInstance *viper.Viper) (string, error) {
	paths := ResolveXDGPaths()
	viperInstance.SetConfigName(ConfigFileName)
	viperInstance.AddConfigPath(paths.ConfigDir())

	if err := viperInstance.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return "", fmt.Errorf("failed to read user config file: %w", err)
		}
		return "", nil
	}
	return viperInstance.ConfigFileUsed(), nil
}

// loadProjectConfig loads project config (./stave.yaml) and merges with existing config.
func loadProjectConfig(viperInstance *viper.Viper, projectDir string) (string, error) {
	if projectDir == "" {
		var err error
		projectDir, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get working directory: %w", err)
		}
	}

	projectConfigPath := filepath.Join(projectDir, ProjectConfigFileName+".yaml")
	if _, statErr := os.Stat(projectConfigPath); statErr != nil {
		// File doesn't exist, which is fine - return empty path with no error
		return "", nil //nolint:nilerr // stat error means file missing, not a failure
	}

	viperInstance.SetConfigFile(projectConfigPath)
	if err := viperInstance.MergeInConfig(); err != nil {
		return "", fmt.Errorf("failed to read project config file: %w", err)
	}
	return projectConfigPath, nil
}

// unmarshalConfig unmarshals viper config into a Config struct and applies env overrides.
func unmarshalConfig(
	viperInstance *viper.Viper,
	opts *LoadOptions,
	configFileUsed string,
) (*Config, error) {
	var cfg Config
	if err := viperInstance.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if !opts.SkipEnv {
		applyEnvironmentOverrides(&cfg)
	}

	cfg.configFile = configFileUsed
	return &cfg, nil
}

// validateAndFinalize applies defaults, expands paths, and validates the config.
func validateAndFinalize(cfg *Config, opts *LoadOptions) (*Config, error) {
	if cfg.CacheDir == "" {
		cfg.CacheDir = ResolveXDGPaths().CacheDir()
	}

	if strings.HasPrefix(cfg.CacheDir, "~/") {
		home := userHomeDir()
		cfg.CacheDir = filepath.Join(home, cfg.CacheDir[2:])
	}

	result := cfg.Validate()
	if result.HasWarnings() {
		result.WriteWarnings(opts.Stderr)
	}
	if result.HasErrors() {
		return nil, errors.New(result.ErrorMessage())
	}

	return cfg, nil
}

// applyEnvironmentOverrides applies environment variable overrides to the config.
// Environment variables take precedence over config file values.
func applyEnvironmentOverrides(cfg *Config) {
	applyStringEnv("STAVEFILE_CACHE", &cfg.CacheDir)
	applyStringEnv("STAVEFILE_GOCMD", &cfg.GoCmd)
	applyStringEnv("STAVEFILE_TARGET_COLOR", &cfg.TargetColor)

	applyBoolEnv("STAVEFILE_VERBOSE", &cfg.Verbose)
	applyBoolEnv("STAVEFILE_DEBUG", &cfg.Debug)
	applyBoolEnv("STAVEFILE_HASHFAST", &cfg.HashFast)
	applyBoolEnv("STAVEFILE_IGNOREDEFAULT", &cfg.IgnoreDefault)
	applyBoolEnv("STAVEFILE_ENABLE_COLOR", &cfg.EnableColor)
}

// applyStringEnv applies an environment variable value to a string pointer if set.
func applyStringEnv(envVar string, target *string) {
	if v := os.Getenv(envVar); v != "" {
		*target = v
	}
}

// applyBoolEnv applies an environment variable value to a bool pointer if set.
// Unset, empty, or invalid values leave the config value unchanged.
func applyBoolEnv(envVar string, target *bool) {
	v, ok := os.LookupEnv(envVar)
	if !ok || v == "" {
		return
	}
	b, err := env.ParseBool(v)
	if err != nil {
		return
	}
	*target = b
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

// File permission constants.
const (
	dirPermission  = 0o755
	filePermission = 0o600
)

// WriteDefaultConfig writes a default configuration file to the user's config directory.
func WriteDefaultConfig() (string, error) {
	paths := ResolveXDGPaths()
	configDir := paths.ConfigDir()

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, dirPermission); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := paths.ConfigFilePath()

	// Check if file already exists
	if _, err := os.Stat(configPath); err == nil {
		return "", fmt.Errorf("config file already exists: %s", configPath)
	}

	// Write default config with restricted permissions for security
	content := defaultConfigYAML()
	if err := os.WriteFile(configPath, []byte(content), filePermission); err != nil {
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
