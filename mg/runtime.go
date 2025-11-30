package mg

import (
	"github.com/yaklabco/stave/config"
)

// Environment variable names for Stave configuration.
// Stave checks for STAVEFILE_* first, then falls back to MAGEFILE_* for backward compatibility.
// These constants are exported for backward compatibility but configuration should be
// managed through the config package.
const (
	// CacheEnv is the environment variable that users may set to change the
	// location where stave stores its compiled binaries.
	CacheEnv = "STAVEFILE_CACHE"

	// VerboseEnv is the environment variable that indicates the user requested
	// verbose mode when running a stavefile.
	VerboseEnv = "STAVEFILE_VERBOSE"

	// DebugEnv is the environment variable that indicates the user requested
	// debug mode when running stave.
	DebugEnv = "STAVEFILE_DEBUG"

	// GoCmdEnv is the environment variable that indicates the go binary the user
	// desires to utilize for Stavefile compilation.
	GoCmdEnv = "STAVEFILE_GOCMD"

	// IgnoreDefaultEnv is the environment variable that indicates the user requested
	// to ignore the default target specified in the stavefile.
	IgnoreDefaultEnv = "STAVEFILE_IGNOREDEFAULT"

	// HashFastEnv is the environment variable that indicates the user requested to
	// use a quick hash of stavefiles to determine whether or not the stavefile binary
	// needs to be rebuilt. This results in faster runtimes, but means that stave
	// will fail to rebuild if a dependency has changed. To force a rebuild, run
	// stave with the -f flag.
	HashFastEnv = "STAVEFILE_HASHFAST"

	// EnableColorEnv is the environment variable that indicates the user is using
	// a terminal which supports a color output. The default is false for backwards
	// compatibility. When the value is true and the detected terminal does support colors
	// then the list of stave targets will be displayed in ANSI color. When the value
	// is true but the detected terminal does not support colors, then the list of
	// stave targets will be displayed in the default colors (e.g. black and white).
	EnableColorEnv = "STAVEFILE_ENABLE_COLOR"

	// TargetColorEnv is the environment variable that indicates which ANSI color
	// should be used to colorize stave targets. This is only applicable when
	// the STAVEFILE_ENABLE_COLOR environment variable is true.
	// The supported ANSI color names are any of these:
	// - Black
	// - Red
	// - Green
	// - Yellow
	// - Blue
	// - Magenta
	// - Cyan
	// - White
	// - BrightBlack
	// - BrightRed
	// - BrightGreen
	// - BrightYellow
	// - BrightBlue
	// - BrightMagenta
	// - BrightCyan
	// - BrightWhite
	TargetColorEnv = "STAVEFILE_TARGET_COLOR"
)

// Legacy environment variable names for backward compatibility with mage.
// These are checked if the STAVEFILE_* equivalents are not set.
const (
	LegacyCacheEnv         = "MAGEFILE_CACHE"
	LegacyVerboseEnv       = "MAGEFILE_VERBOSE"
	LegacyDebugEnv         = "MAGEFILE_DEBUG"
	LegacyGoCmdEnv         = "MAGEFILE_GOCMD"
	LegacyIgnoreDefaultEnv = "MAGEFILE_IGNOREDEFAULT"
	LegacyHashFastEnv      = "MAGEFILE_HASHFAST"
	LegacyEnableColorEnv   = "MAGEFILE_ENABLE_COLOR"
	LegacyTargetColorEnv   = "MAGEFILE_TARGET_COLOR"
)

// Verbose reports whether a stavefile was run with the verbose flag.
// This delegates to the global config loaded by the config package.
func Verbose() bool {
	return config.Global().Verbose
}

// Debug reports whether a stavefile was run with the debug flag.
// This delegates to the global config loaded by the config package.
func Debug() bool {
	return config.Global().Debug
}

// GoCmd reports the command that Stave will use to build go code. By default stave runs
// the "go" binary in the PATH.
// This delegates to the global config loaded by the config package.
func GoCmd() string {
	return config.Global().GoCmd
}

// HashFast reports whether the user has requested to use the fast hashing
// mechanism rather than rely on go's rebuilding mechanism.
// This delegates to the global config loaded by the config package.
func HashFast() bool {
	return config.Global().HashFast
}

// IgnoreDefault reports whether the user has requested to ignore the default target
// in the stavefile.
// This delegates to the global config loaded by the config package.
func IgnoreDefault() bool {
	return config.Global().IgnoreDefault
}

// CacheDir returns the directory where stave caches compiled binaries. It
// defaults to the XDG cache directory, but may be overridden by the STAVEFILE_CACHE
// environment variable (or MAGEFILE_CACHE for backward compatibility) or config file.
// This delegates to the global config loaded by the config package.
func CacheDir() string {
	return config.Global().CacheDir
}

// EnableColor reports whether the user has requested to enable a color output.
// This delegates to the global config loaded by the config package.
func EnableColor() bool {
	return config.Global().EnableColor
}

// TargetColor returns the configured ANSI color name for color output.
// This delegates to the global config loaded by the config package.
func TargetColor() string {
	cfg := config.Global()
	// Use the color helper to ensure we return a valid ANSI color
	if c, ok := getAnsiColor(cfg.TargetColor); ok {
		return c
	}
	return DefaultTargetAnsiColor
}

// Namespace allows for the grouping of similar commands
type Namespace struct{}
