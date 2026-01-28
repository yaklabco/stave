package stave

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/fang"
	"github.com/yaklabco/stave/cmd/stave/version"
	"github.com/yaklabco/stave/pkg/st"
	"github.com/yaklabco/stave/pkg/stave"

	"github.com/spf13/cobra"
)

const (
	shortDescription = "Stave is a Go-native, make-like command runner. " +
		"It is a fork of mage. See https://github.com/yaklabco/stave"
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
	stave --hooks install
	stave --hooks list

	# Delegate to direnv
	stave --direnv <args>

	# Manage configuration
	stave --config show`,
		Version: version.OverallVersionStringColorized(ctx),
		ValidArgsFunction: func(cmd *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			dir, err := cmd.Root().PersistentFlags().GetString("dir")
			if err != nil {
				return nil, cobra.ShellCompDirectiveError
			}

			targets, err := stave.TargetNames(cmd.Context(), dir)
			if err != nil {
				return nil, cobra.ShellCompDirectiveError
			}
			return targets, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			runParams.Args = args
			runParams.WriterForLogger = os.Stdout
			runParams.BaseCtx = cmd.Context() //nolint:fatcontext // intentionally setting context from cmd

			return rootCmdOpts.runFunc(runParams)
		},
	}

	// Flags.
	rootCmd.PersistentFlags().BoolVarP(&runParams.Debug, "debug", "d", st.Debug(), "turn on debug messages")
	rootCmd.PersistentFlags().StringVarP(&runParams.Dir, "dir", "C", "", "directory to read stavefiles from")
	rootCmd.PersistentFlags().BoolVar(&runParams.DryRun, "dryrun", false, "print commands instead of executing them")
	rootCmd.PersistentFlags().BoolVarP(&runParams.Force, "force", "f", false, "force recreation of compiled stavefile")
	rootCmd.PersistentFlags().StringVar(&runParams.GOARCH, "goarch", "", "set GOARCH for binary produced with --compile")
	rootCmd.PersistentFlags().StringVar(&runParams.GoCmd, "gocmd", st.GoCmd(), "use the given go binary to compile the output")
	rootCmd.PersistentFlags().StringVar(&runParams.GOOS, "goos", "", "set GOOS for binary produced with --compile")
	rootCmd.PersistentFlags().BoolVarP(&runParams.Info, "info", "i", st.Info(), "show docstring for a specific target")
	rootCmd.PersistentFlags().BoolVar(&runParams.Keep, "keep", false, "keep intermediate stave files around after running")
	rootCmd.PersistentFlags().StringVar(&runParams.Ldflags, "ldflags", "", "set ldflags for binary produced with --compile")
	rootCmd.PersistentFlags().DurationVarP(&runParams.Timeout, "timeout", "t", 0, "timeout in duration parsable format (e.g. 5m30s)")
	rootCmd.PersistentFlags().BoolVarP(&runParams.Verbose, "verbose", "v", st.Verbose(), "show verbose output when running stave targets")
	rootCmd.PersistentFlags().StringVarP(&runParams.WorkDir, "workdir", "w", "", "working directory where stavefiles will run")

	// Flags that are actually commands ("pseudo-flags").
	rootCmd.PersistentFlags().BoolVar(&runParams.Clean, "clean", false, "clean out old generated binaries from CACHE_DIR")
	rootCmd.PersistentFlags().StringVar(&runParams.CompileOut, "compile", "", "output a static binary to the given path")
	rootCmd.PersistentFlags().BoolVar(&runParams.Config, "config", false, "manage stave configuration")
	rootCmd.PersistentFlags().BoolVar(&runParams.DirEnv, "direnv", false, "delegate to direnv for managing environment variables")
	rootCmd.PersistentFlags().BoolVar(&runParams.Exec, "exec", false, "execute commands under stave")
	rootCmd.PersistentFlags().BoolVar(&runParams.Hooks, "hooks", false, "manage git hooks (install, list, run, etc.)")
	rootCmd.PersistentFlags().BoolVar(&runParams.Init, "init", false, "create a starting template if no stave files exist")
	rootCmd.PersistentFlags().BoolVarP(&runParams.List, "list", "l", false, "list stave targets in this directory")

	// Mark --exec as hidden for now, since it doesn't do anything interesting (yet!), and users may therefore be confused by its existence.
	// Revisit this as Stave's functionality expands.
	err := rootCmd.PersistentFlags().MarkHidden("exec")
	if err != nil {
		panic(fmt.Errorf("failed to mark --exec as hidden: %w", err))
	}

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
