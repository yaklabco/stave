package config

import (
	"time"

	"github.com/spf13/viper"
)

// Default configuration values.
const (
	// DefaultGoCmd is the default Go command to use for compilation.
	DefaultGoCmd = "go"

	// DefaultVerbose is the default verbose setting.
	DefaultVerbose = false

	// DefaultDebug is the default debug setting.
	DefaultDebug = false

	// DefaultHashFast is the default hash fast setting.
	DefaultHashFast = false

	// DefaultIgnoreDefault is the default ignore default target setting.
	DefaultIgnoreDefault = false

	// DefaultEnableColor is the default color output setting.
	DefaultEnableColor = false

	// DefaultTargetColor is the default ANSI color for target names.
	DefaultTargetColor = "Cyan"

	// DefaultUpdateCheckEnabled controls whether update checks are enabled by default.
	DefaultUpdateCheckEnabled = true
)

// DefaultUpdateCheckInterval is the default duration between update checks.
var DefaultUpdateCheckInterval = 24 * time.Hour //nolint:gochecknoglobals // default configuration value

// setDefaults configures default values in the viper instance.
func setDefaults(viperInstance *viper.Viper) {
	viperInstance.SetDefault("cache_dir", "")
	viperInstance.SetDefault("go_cmd", DefaultGoCmd)
	viperInstance.SetDefault("verbose", DefaultVerbose)
	viperInstance.SetDefault("debug", DefaultDebug)
	viperInstance.SetDefault("hash_fast", DefaultHashFast)
	viperInstance.SetDefault("ignore_default", DefaultIgnoreDefault)
	viperInstance.SetDefault("enable_color", DefaultEnableColor)
	viperInstance.SetDefault("target_color", DefaultTargetColor)
	viperInstance.SetDefault("update_check.enabled", DefaultUpdateCheckEnabled)
	viperInstance.SetDefault("update_check.interval", DefaultUpdateCheckInterval)
}
