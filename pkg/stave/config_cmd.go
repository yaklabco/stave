package stave

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/yaklabco/stave/config"
)

// ConfigSubcommand represents a config subcommand.
type ConfigSubcommand string

// Config subcommand constants.
const (
	ConfigInit ConfigSubcommand = "init"
	ConfigShow ConfigSubcommand = "show"
	ConfigPath ConfigSubcommand = "path"
)

// RunConfigCommand handles the `stave config` subcommand.
// It returns the exit code.
func RunConfigCommand(stdout, stderr io.Writer, args []string) int {
	fs := flag.NewFlagSet("config", flag.ContinueOnError)
	fs.SetOutput(stdout)
	fs.Usage = func() {
		configUsage(stdout)
	}

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		_, _ = fmt.Fprintf(stderr, "Error: %v\n", err)
		return 2
	}

	subArgs := fs.Args()
	if len(subArgs) == 0 {
		// No subcommand, show effective config
		return runConfigShow(stdout, stderr)
	}

	subcmd := ConfigSubcommand(strings.ToLower(subArgs[0]))
	switch subcmd {
	case ConfigInit:
		return runConfigInit(stdout, stderr)
	case ConfigShow:
		return runConfigShow(stdout, stderr)
	case ConfigPath:
		return runConfigPath(stdout, stderr)
	default:
		_, _ = fmt.Fprintf(stderr, "Error: unknown config subcommand %q\n", subArgs[0])
		configUsage(stderr)
		return 2
	}
}

// runConfigInit creates a default configuration file.
func runConfigInit(stdout, stderr io.Writer) int {
	path, err := config.WriteDefaultConfig()
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}
	_, _ = fmt.Fprintf(stdout, "Created config file: %s\n", path)
	return 0
}

// runConfigShow displays the effective configuration.
func runConfigShow(stdout, stderr io.Writer) int {
	cfg, err := config.Load(nil)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Error loading config: %v\n", err)
		return 1
	}

	_, _ = fmt.Fprintln(stdout, "# Effective Stave Configuration")
	if cfg.ConfigFile() != "" {
		_, _ = fmt.Fprintf(stdout, "# Loaded from: %s\n", cfg.ConfigFile())
	} else {
		_, _ = fmt.Fprintln(stdout, "# (using defaults, no config file found)")
	}
	_, _ = fmt.Fprintln(stdout)
	_, _ = fmt.Fprintf(stdout, "cache_dir: %s\n", cfg.CacheDir)
	_, _ = fmt.Fprintf(stdout, "go_cmd: %s\n", cfg.GoCmd)
	_, _ = fmt.Fprintf(stdout, "verbose: %v\n", cfg.Verbose)
	_, _ = fmt.Fprintf(stdout, "debug: %v\n", cfg.Debug)
	_, _ = fmt.Fprintf(stdout, "hash_fast: %v\n", cfg.HashFast)
	_, _ = fmt.Fprintf(stdout, "ignore_default: %v\n", cfg.IgnoreDefault)
	_, _ = fmt.Fprintf(stdout, "enable_color: %v\n", cfg.EnableColor)
	_, _ = fmt.Fprintf(stdout, "target_color: %s\n", cfg.TargetColor)

	return 0
}

// runConfigPath displays the configuration file paths.
func runConfigPath(stdout, _ io.Writer) int {
	paths := config.ResolveXDGPaths()

	_, _ = fmt.Fprintln(stdout, "Configuration Paths:")
	_, _ = fmt.Fprintf(stdout, "  User config:    %s\n", paths.ConfigFilePath())
	_, _ = fmt.Fprintf(stdout, "  Config dir:     %s\n", paths.ConfigDir())
	_, _ = fmt.Fprintf(stdout, "  Cache dir:      %s\n", paths.CacheDir())
	_, _ = fmt.Fprintf(stdout, "  Data dir:       %s\n", paths.DataDir())

	// Check if user config exists
	cfg, err := config.Load(nil)
	if err == nil && cfg.ConfigFile() != "" {
		_, _ = fmt.Fprintf(stdout, "\nActive config file: %s\n", cfg.ConfigFile())
	} else {
		_, _ = fmt.Fprintln(stdout, "\nNo config file currently loaded (using defaults)")
	}

	return 0
}

// configUsage prints the config command usage.
func configUsage(w io.Writer) {
	_, _ = fmt.Fprint(w, `
stave config [subcommand]

Manage Stave configuration.

Subcommands:
  init    Create a default configuration file
  show    Display effective configuration (default)
  path    Show configuration file paths

Examples:
  stave config           # Show effective configuration
  stave config init      # Create ~/.config/stave/config.yaml
  stave config show      # Same as 'stave config'
  stave config path      # Show config file locations
`[1:])
}
