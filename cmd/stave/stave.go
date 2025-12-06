package stave

import (
	"context"
	"os"

	"github.com/charmbracelet/fang"
	"github.com/yaklabco/stave/cmd/stave/version"
	"github.com/yaklabco/stave/pkg/st"
	"github.com/yaklabco/stave/pkg/stave"

	"github.com/spf13/cobra"
)

// exitError is returned when a subcommand fails with a non-zero exit code.
type exitError struct {
	code int
	cmd  string
}

func (e *exitError) Error() string {
	return e.cmd + " command failed"
}

// ExitCode returns the exit code from the failed command.
func (e *exitError) ExitCode() int {
	return e.code
}

const (
	shortDescription = "Stave is a Go-native, make-like command runner. " +
		"It is a fork of Mage. See https://github.com/yaklabco/stave"
)

type rootCmdOptions struct {
	runFunc func(params stave.RunParams) error
}

type Option func(*rootCmdOptions)

// This is intentionally designed to be unusable from outside this package,
// as it exists purely for testing purposes.
func withRunFunc(fn func(params stave.RunParams) error) Option {
	return func(opts *rootCmdOptions) {
		opts.runFunc = fn
	}
}

func NewRootCmd(ctx context.Context, opts ...Option) *cobra.Command {
	rootCmdOpts := &rootCmdOptions{
		runFunc: stave.Run,
	}
	for _, opt := range opts {
		opt(rootCmdOpts)
	}

	var runParams stave.RunParams
	rootCmd := &cobra.Command{
		Use:   "stave [flags] [target]",
		Short: shortDescription,
		Example: `	# Run the default target
		stave

	# Run specific targets
	stave test
	stave build

	# Manage Git hooks
	stave hooks install
	stave hooks list

	# Manage configuration
	stave config show`,
		Version: version.OverallVersionStringColorized(ctx),
		//nolint:contextcheck // context is passed via cmd.Context() to subcommands and via runParams.BaseCtx to main run
		RunE: func(cmd *cobra.Command, args []string) error {
			// Handle built-in subcommands before delegating to target execution
			if len(args) > 0 {
				switch args[0] {
				case "hooks":
					hooksParams := stave.HooksParams{
						Debug:   runParams.Debug,
						Verbose: runParams.Verbose,
					}
					exitCode := stave.RunHooksCommandWithParams(cmd.Context(), os.Stdout, os.Stderr, hooksParams, args[1:])
					if exitCode != 0 {
						return &exitError{code: exitCode, cmd: "hooks"}
					}
					return nil
				case "config":
					exitCode := stave.RunConfigCommandContext(cmd.Context(), os.Stdout, os.Stderr, args[1:])
					if exitCode != 0 {
						return &exitError{code: exitCode, cmd: "config"}
					}
					return nil
				}
			}

			runParams.Args = args
			runParams.WriterForLogger = os.Stdout
			runParams.BaseCtx = cmd.Context() //nolint:fatcontext // intentionally setting context from cmd

			return rootCmdOpts.runFunc(runParams)
		},
	}

	rootCmd.PersistentFlags().BoolVarP(&runParams.Force, "force", "f", false, "force recreation of compiled stavefile")
	rootCmd.PersistentFlags().BoolVarP(&runParams.Debug, "debug", "d", st.Debug(), "turn on debug messages")
	rootCmd.PersistentFlags().BoolVarP(
		&runParams.Verbose, "verbose", "v", st.Verbose(), "show verbose output when running stave targets")
	rootCmd.PersistentFlags().BoolVarP(&runParams.Info, "info", "i", st.Info(), "show docstring for a specific target")
	rootCmd.PersistentFlags().DurationVarP(
		&runParams.Timeout, "timeout", "t", 0, "timeout in duration parsable format (e.g. 5m30s)")
	rootCmd.PersistentFlags().BoolVar(&runParams.Keep, "keep", false, "keep intermediate stave files around after running")
	rootCmd.PersistentFlags().BoolVar(&runParams.DryRun, "dryrun", false, "print commands instead of executing them")
	rootCmd.PersistentFlags().StringVarP(&runParams.Dir, "dir", "C", "", "directory to read stavefiles from")
	rootCmd.PersistentFlags().StringVarP(
		&runParams.WorkDir, "workdir", "w", "", "working directory where stavefiles will run")
	rootCmd.PersistentFlags().StringVar(
		&runParams.GoCmd, "gocmd", st.GoCmd(), "use the given go binary to compile the output")
	rootCmd.PersistentFlags().StringVar(&runParams.GOOS, "goos", "", "set GOOS for binary produced with -compile")
	rootCmd.PersistentFlags().StringVar(&runParams.GOARCH, "goarch", "", "set GOARCH for binary produced with -compile")
	rootCmd.PersistentFlags().StringVar(&runParams.Ldflags, "ldflags", "", "set ldflags for binary produced with -compile")

	// commands below

	rootCmd.PersistentFlags().BoolVarP(&runParams.List, "list", "l", false, "list stave targets in this directory")
	rootCmd.PersistentFlags().BoolVar(&runParams.Init, "init", false, "create a starting template if no stave files exist")
	rootCmd.PersistentFlags().BoolVar(&runParams.Clean, "clean", false, "clean out old generated binaries from CACHE_DIR")
	rootCmd.PersistentFlags().BoolVar(&runParams.Exec, "exec", false, "execute commands under stave")
	rootCmd.PersistentFlags().StringVar(&runParams.CompileOut, "compile", "", "output a static binary to the given path")

	return rootCmd
}

// ExecuteWithFang runs the root Cobra command with Fang-specific options.
// It accepts a context and a root Cobra command as input parameters.
// Returns an error if the command execution fails.
func ExecuteWithFang(ctx context.Context, rootCmd *cobra.Command) error {
	//nolint:wrapcheck // top-level error from cobra, wrapping not needed
	return fang.Execute(
		ctx, rootCmd, fang.WithVersion(rootCmd.Version), fang.WithoutManpage())
}
