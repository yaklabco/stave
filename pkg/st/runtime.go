package st

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/yaklabco/stave/internal/dryrun"
	"github.com/yaklabco/stave/internal/env"
)

// CacheEnv is the environment variable that users may set to change the
// location where stave stores its compiled binaries.
const CacheEnv = "STAVEFILE_CACHE"

// VerboseEnv is the environment variable that indicates the user requested
// verbose mode when running a stavefile.
const VerboseEnv = "STAVEFILE_VERBOSE"

// DebugEnv is the environment variable that indicates the user requested
// debug mode when running stave.
const DebugEnv = "STAVEFILE_DEBUG"

// InfoEnv is the environment variable that indicates the user requested
// the docstring for one of the targets.
const InfoEnv = "STAVEFILE_INFO"

// DryRunRequestedEnv is the environment variable that indicates the user requested dryrun mode when running Stave.
const DryRunRequestedEnv = dryrun.RequestedEnv

// DryRunPossibleEnv is the environment variable that indicates we are in a context where a dry run is possible.
const DryRunPossibleEnv = dryrun.PossibleEnv

// GoCmdEnv is the environment variable that indicates the go binary the user
// desires to utilize for Stavefile compilation.
const GoCmdEnv = "STAVEFILE_GOCMD"

// IgnoreDefaultEnv is the environment variable that indicates the user requested
// to ignore the default target specified in the stavefile.
const IgnoreDefaultEnv = "STAVEFILE_IGNOREDEFAULT"

// HashFastEnv is the environment variable that indicates the user requested to
// use a quick hash of stavefiles to determine whether or not the stavefile binary
// needs to be rebuilt. This results in faster runtimes, but means that stave
// will fail to rebuild if a dependency has changed. To force a rebuild, run
// stave with the -f flag.
const HashFastEnv = "STAVEFILE_HASHFAST"

// EnableColorEnv is the environment variable that indicates the user is using
// a terminal which supports a color output. The default is false for backwards
// compatibility. When the value is true and the detected terminal does support colors
// then the list of stave targets will be displayed in ANSI color. When the value
// is true but the detected terminal does not support colors, then the list of
// stave targets will be displayed in the default colors (e.g. black and white).
const EnableColorEnv = "STAVEFILE_ENABLE_COLOR"

// TargetColorEnv is the environment variable that indicates which ANSI color
// should be used to colorize stave targets. This is only applicable when
// the STAVEFILE_ENABLE_COLOR environment variable is true.
// The supported ANSI color names are any of these:
// - Black
// - Red
// - Green
// - Yellow
// - Blue
// - Staventa
// - Cyan
// - White
// - BrightBlack
// - BrightRed
// - BrightGreen
// - BrightYellow
// - BrightBlue
// - BrightStaventa
// - BrightCyan
// - BrightWhite.
const TargetColorEnv = "STAVEFILE_TARGET_COLOR"

// Verbose reports whether a stavefile was run with the verbose flag.
func Verbose() bool {
	return env.ParseBoolEnvDefaultFalse(VerboseEnv)
}

// Debug reports whether a stavefile was run with the debug flag.
func Debug() bool {
	return env.ParseBoolEnvDefaultFalse(DebugEnv)
}

// Info reports whether a stavefile was run with the info flag.
func Info() bool {
	return env.ParseBoolEnvDefaultFalse(InfoEnv)
}

// GoCmd reports the command that Stave will use to build go code.  By default stave runs
// the "go" binary in the PATH.
func GoCmd() string {
	if cmd := os.Getenv(GoCmdEnv); cmd != "" {
		return cmd
	}
	return "go"
}

// HashFast reports whether the user has requested to use the fast hashing
// mechanism rather than rely on go's rebuilding mechanism.
func HashFast() bool {
	return env.ParseBoolEnvDefaultFalse(HashFastEnv)
}

// IgnoreDefault reports whether the user has requested to ignore the default target
// in the stavefile.
func IgnoreDefault() bool {
	return env.ParseBoolEnvDefaultFalse(IgnoreDefaultEnv)
}

// CacheDir returns the directory where stave caches compiled binaries.  It
// defaults to $HOME/.stavefile, but may be overridden by the STAVEFILE_CACHE
// environment variable.
func CacheDir() string {
	d := os.Getenv(CacheEnv)
	if d != "" {
		return d
	}
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("HOMEDRIVE"), os.Getenv("HOMEPATH"), "stavefile")
	default:
		return filepath.Join(os.Getenv("HOME"), ".stavefile")
	}
}

// EnableColor reports whether the user has requested to enable a color output.
func EnableColor() bool {
	return env.ParseBoolEnvDefaultFalse(EnableColorEnv)
}

// TargetColor returns the configured ANSI color name a color output.
func TargetColor() string {
	s, exists := os.LookupEnv(TargetColorEnv)
	if exists {
		if c, ok := getAnsiColor(s); ok {
			return c
		}
	}
	return DefaultTargetAnsiColor
}

// Namespace allows for the grouping of similar commands.
type Namespace struct{}
