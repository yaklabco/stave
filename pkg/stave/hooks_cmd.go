package stave

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sort"
	"strings"

	"github.com/samber/lo"
	"github.com/yaklabco/stave/config"
	"github.com/yaklabco/stave/internal/hooks"
	"github.com/yaklabco/stave/pkg/st"
)

// Exit codes for CLI commands.
const (
	exitOK    = 0
	exitError = 1
	exitUsage = 2
)

const (
	HooksAreRunningEnv = "STAVEFILE_HOOKS_RUNNING"
)

// printErr writes "Error: <message>\n" to w and returns exitError.
func printErr(w io.Writer, err error) int {
	_, _ = fmt.Fprintf(w, "Error: %v\n", err)
	return exitError
}

// printUsageErr writes "Error: <message>\n" to w and returns exitUsage.
func printUsageErr(w io.Writer, err error) int {
	_, _ = fmt.Fprintf(w, "Error: %v\n", err)
	return exitUsage
}

// printConfigErr writes "Error loading config: <message>\n" to w and returns exitError.
func printConfigErr(w io.Writer, err error) int {
	_, _ = fmt.Fprintf(w, "Error loading config: %v\n", err)
	return exitError
}

// HooksParams contains parameters for the hooks command.
type HooksParams struct {
	Debug   bool
	Verbose bool
}

// newStaveTargetRunner creates a TargetRunnerFunc that executes targets using stave.Run.
// This wires the hooks runtime to the real Stave execution engine.
func newStaveTargetRunner(cfg *config.Config, workingDir string) hooks.TargetRunnerFunc {
	return func(
		ctx context.Context,
		target string,
		args []string,
		stdin io.Reader,
		stdout, stderr io.Writer,
	) (int, error) {
		runParams := RunParams{
			BaseCtx: ctx,
			Stdin:   stdin,
			Stdout:  stdout,
			Stderr:  stderr,
			Dir:     workingDir,

			// Propagate config-level settings
			Debug:    cfg.Debug,
			Verbose:  cfg.Verbose,
			HashFast: cfg.HashFast,
			GoCmd:    cfg.GoCmd,
			CacheDir: cfg.CacheDir,

			// Target invocation: prepend target name to args
			Args: append([]string{target}, args...),

			HooksAreRunning: true,
		}

		err := Run(runParams)
		return st.ExitStatus(err), err
	}
}

// RunHooksCommand handles the `stave --hooks` subcommand with debug/verbose params.
// It returns the exit code.
func RunHooksCommand(ctx context.Context, params RunParams) int {
	flagSet := flag.NewFlagSet("hooks", flag.ContinueOnError)
	flagSet.SetOutput(params.Stdout)
	flagSet.Usage = func() {
		hooksUsage(params.Stdout)
	}

	if err := flagSet.Parse(params.Args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return exitOK
		}
		_, _ = fmt.Fprintf(params.Stderr, "Error: %v\n", err)
		return exitUsage
	}

	subArgs := flagSet.Args()
	if len(subArgs) == 0 {
		// No subcommand, show list
		return runHooksList(ctx, params)
	}

	return dispatchHooksSubcommand(ctx, params, subArgs)
}

// HooksSubcommand represents a hooks subcommand.
type HooksSubcommand string

// Hooks subcommand constants.
const (
	HooksInit      HooksSubcommand = "init"
	HooksInstall   HooksSubcommand = "install"
	HooksUninstall HooksSubcommand = "uninstall"
	HooksList      HooksSubcommand = "list"
	HooksRun       HooksSubcommand = "run"
)

func dispatchHooksSubcommand(ctx context.Context, params RunParams, subArgs []string) int {
	subcmd := HooksSubcommand(strings.ToLower(subArgs[0]))

	slog.Debug("hooks subcommand dispatching",
		slog.String("subcommand", string(subcmd)))

	switch subcmd {
	case HooksInit:
		return runHooksInit(ctx, params)
	case HooksInstall:
		return runHooksInstall(ctx, params, subArgs[1:])
	case HooksUninstall:
		return runHooksUninstall(ctx, params, subArgs[1:])
	case HooksList:
		return runHooksList(ctx, params)
	case HooksRun:
		return runHooksRun(ctx, params, subArgs[1:])
	default:
		slog.Debug("unknown hooks subcommand",
			slog.String("subcommand", subArgs[0]))
		_, _ = fmt.Fprintf(params.Stderr, "Error: unknown hooks subcommand %q\n", subArgs[0])
		hooksUsage(params.Stderr)
		return exitUsage
	}
}

// runHooksInit initializes hooks configuration in stave.yaml.
func runHooksInit(ctx context.Context, params RunParams) int {
	slog.Debug("loading hooks configuration")

	// First ensure config exists
	cfg, err := config.Load(&config.LoadOptions{ProjectDir: params.Dir})
	if err != nil {
		return printConfigErr(params.Stderr, err)
	}

	// Check if hooks are already configured
	if len(cfg.Hooks) > 0 {
		slog.Debug("hooks already configured",
			slog.Int("hook_count", len(cfg.Hooks)))
		_, _ = fmt.Fprintln(params.Stdout, "Hooks configuration already exists in stave.yaml")
		return runHooksList(ctx, params)
	}

	printHooksInitInstructions(params.Stdout)
	return exitOK
}

func printHooksInitInstructions(out io.Writer) {
	_, _ = fmt.Fprintln(out, staveInitText)
}

// runHooksInstall installs hook scripts to the Git repository.
func runHooksInstall(ctx context.Context, params RunParams, args []string) int {
	flagSet := flag.NewFlagSet("install", flag.ContinueOnError)
	flagSet.SetOutput(params.Stdout)
	force := flagSet.Bool("force", false, "overwrite existing non-Stave hooks")

	if err := flagSet.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return exitOK
		}
		return printUsageErr(params.Stderr, err)
	}

	slog.Debug("hooks install starting",
		slog.Bool("force", *force))

	// Find Git repository
	repo, err := hooks.FindGitRepoContext(ctx, params.Dir)
	if err != nil {
		if errors.Is(err, hooks.ErrNotGitRepo) {
			_, _ = fmt.Fprintf(params.Stderr, "Error: %v\n", err)
			_, _ = fmt.Fprintln(params.Stderr, "Run this command from within a Git repository.")
			return exitError
		}
		return printErr(params.Stderr, err)
	}

	// Load configuration
	slog.Debug("loading hooks configuration")
	cfg, err := config.Load(&config.LoadOptions{ProjectDir: params.Dir})
	if err != nil {
		return printConfigErr(params.Stderr, err)
	}

	// Check if hooks are configured
	if len(cfg.Hooks) == 0 {
		slog.Debug("no hooks configured in config")
		_, _ = fmt.Fprintln(params.Stderr, "No hooks configured in stave.yaml")
		_, _ = fmt.Fprintln(params.Stderr, "Run 'stave --hooks init' for setup instructions.")
		return exitError
	}

	return installHooks(repo, cfg, *force, params)
}

func installHooks(repo *hooks.GitRepo, cfg *config.Config, force bool, params RunParams) int {
	// Ensure hooks directory exists
	if err := repo.EnsureHooksDir(); err != nil {
		_, _ = fmt.Fprintf(params.Stderr, "Error: creating hooks directory: %v\n", err)
		return exitError
	}

	// Install each configured hook
	hookNames := cfg.Hooks.HookNames()
	slog.Debug("installing hooks",
		slog.Int("hook_count", len(hookNames)),
		slog.String("directory", repo.HooksPath()))

	installed := 0
	for _, hookName := range hookNames {
		if code := installSingleHook(repo, hookName, force, params.Stdout, params.Stderr); code != exitOK {
			return code
		}
		installed++
	}

	slog.Debug("hooks installed",
		slog.Int("count", installed),
		slog.String("directory", repo.HooksPath()))

	_, _ = fmt.Fprintf(params.Stdout, "\nInstalled %d hook(s) to %s\n", installed, repo.HooksPath())
	return exitOK
}

func installSingleHook(repo *hooks.GitRepo, hookName string, force bool, stdout, stderr io.Writer) int {
	hookPath := repo.HookPath(hookName)

	slog.Debug("hook installation check",
		slog.String("hook", hookName),
		slog.String("path", hookPath))

	// Check for existing non-Stave hook
	managed, err := hooks.IsStaveManaged(hookPath)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Error: checking %s: %v\n", hookName, err)
		return exitError
	}

	slog.Debug("hook managed status",
		slog.String("hook", hookName),
		slog.Bool("managed", managed))

	// Check if file exists and is not Stave-managed
	if !managed {
		if _, statErr := os.Stat(hookPath); statErr == nil {
			if !force {
				slog.Debug("existing non-stave hook found",
					slog.String("hook", hookName))
				_, _ = fmt.Fprintf(stderr, "Error: %s already exists and was not installed by Stave\n", hookName)
				_, _ = fmt.Fprintln(stderr, "Use --force to overwrite, or remove the existing hook first.")
				return exitError
			}
			slog.Debug("overwriting existing hook",
				slog.String("hook", hookName))
			_, _ = fmt.Fprintf(stdout, "Overwriting existing %s hook\n", hookName)
		}
	}

	// Write the hook script
	if err := hooks.WriteHookScript(hookPath, hooks.ScriptParams{HookName: hookName}); err != nil {
		_, _ = fmt.Fprintf(stderr, "Error: writing %s: %v\n", hookName, err)
		return exitError
	}
	_, _ = fmt.Fprintf(stdout, "Installed %s\n", hookName)
	return exitOK
}

// runHooksUninstall removes Stave-managed hook scripts.
func runHooksUninstall(ctx context.Context, params RunParams, args []string) int {
	flagSet := flag.NewFlagSet("uninstall", flag.ContinueOnError)
	flagSet.SetOutput(params.Stdout)
	all := flagSet.Bool("all", false, "uninstall all Stave-managed hooks")

	if err := flagSet.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return exitOK
		}
		return printUsageErr(params.Stderr, err)
	}

	// Find Git repository
	repo, err := hooks.FindGitRepoContext(ctx, params.Dir)
	if err != nil {
		return printErr(params.Stderr, err)
	}

	// Load configuration for hook names
	cfg, err := config.Load(&config.LoadOptions{ProjectDir: params.Dir})
	if err != nil {
		return printConfigErr(params.Stderr, err)
	}

	hookNames := getHookNamesToUninstall(*all, cfg)
	if len(hookNames) == 0 {
		_, _ = fmt.Fprintln(params.Stdout, "No hooks to uninstall.")
		return exitOK
	}

	return uninstallHooks(repo, hookNames, params.Stdout, params.Stderr)
}

func getHookNamesToUninstall(all bool, cfg *config.Config) []string {
	if all {
		return config.KnownGitHookNames()
	}
	if cfg.Hooks != nil {
		return cfg.Hooks.HookNames()
	}
	return nil
}

func uninstallHooks(repo *hooks.GitRepo, hookNames []string, stdout, stderr io.Writer) int {
	slog.Debug("uninstalling hooks",
		slog.Int("hook_count", len(hookNames)))

	removed := 0
	for _, hookName := range hookNames {
		hookPath := repo.HookPath(hookName)
		wasRemoved, err := hooks.RemoveHookScript(hookPath)
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "Error: removing %s: %v\n", hookName, err)
			continue
		}
		if wasRemoved {
			_, _ = fmt.Fprintf(stdout, "Removed %s\n", hookName)
			removed++
		}
	}

	if removed == 0 {
		slog.Debug("no stave-managed hooks found to remove")
		_, _ = fmt.Fprintln(stdout, "No Stave-managed hooks found to remove.")
	} else {
		slog.Info("hooks removed",
			slog.Int("count", removed))
		_, _ = fmt.Fprintf(stdout, "\nRemoved %d hook(s)\n", removed)
	}
	return exitOK
}

// runHooksList displays configured hooks.
func runHooksList(ctx context.Context, params RunParams) int {
	slog.Debug("loading hooks configuration for list")

	cfg, err := config.Load(&config.LoadOptions{ProjectDir: params.Dir})
	if err != nil {
		return printConfigErr(params.Stderr, err)
	}

	if len(cfg.Hooks) == 0 {
		slog.Debug("no hooks configured")
		_, _ = fmt.Fprintln(params.Stdout, "No hooks configured.")
		_, _ = fmt.Fprintln(params.Stdout, "Run 'stave --hooks init' for setup instructions.")
		return exitOK
	}

	slog.Debug("listing configured hooks",
		slog.Int("hook_count", len(cfg.Hooks)))

	printConfiguredHooks(cfg, params.Stdout)

	// Check installation status
	repo, repoErr := hooks.FindGitRepoContext(ctx, params.Dir)
	if repoErr == nil {
		printInstallationStatus(repo, cfg.Hooks.HookNames(), params.Stdout)
	}

	return exitOK
}

func printConfiguredHooks(cfg *config.Config, out io.Writer) {
	_, _ = fmt.Fprintln(out, "Configured Git hooks:")
	_, _ = fmt.Fprintln(out)

	hookNames := cfg.Hooks.HookNames()
	for _, hookName := range hookNames {
		targets := cfg.Hooks.Get(hookName)
		_, _ = fmt.Fprintf(out, "  %s:\n", hookName)
		for _, target := range targets {
			if len(target.Args) > 0 {
				_, _ = fmt.Fprintf(out, "    - %s %s\n", target.Target, strings.Join(target.Args, " "))
			} else {
				_, _ = fmt.Fprintf(out, "    - %s\n", target.Target)
			}
		}
	}
}

func printInstallationStatus(repo *hooks.GitRepo, hookNames []string, out io.Writer) {
	_, _ = fmt.Fprintln(out)
	installed := 0
	var missing []string
	for _, hookName := range hookNames {
		hookPath := repo.HookPath(hookName)
		managed, err := hooks.IsStaveManaged(hookPath)
		if err != nil {
			managed = false
		}
		if managed {
			installed++
		} else {
			missing = append(missing, hookName)
		}
	}

	if installed == len(hookNames) {
		_, _ = fmt.Fprintf(out, "All %d hook(s) installed.\n", installed)
	} else {
		_, _ = fmt.Fprintf(out, "%d of %d hook(s) installed.\n", installed, len(hookNames))
		if len(missing) > 0 {
			sort.Strings(missing)
			_, _ = fmt.Fprintf(out, "Missing: %s\n", strings.Join(missing, ", "))
			_, _ = fmt.Fprintln(out, "Run 'stave --hooks install' to install missing hooks.")
		}
	}
}

// runHooksRun executes the targets for a specific hook.
func runHooksRun(ctx context.Context, params RunParams, args []string) int {
	flagSet := flag.NewFlagSet("run", flag.ContinueOnError)
	flagSet.SetOutput(params.Stdout)

	if err := flagSet.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return exitOK
		}
		return printUsageErr(params.Stderr, err)
	}

	remaining := flagSet.Args()
	if len(remaining) == 0 {
		_, _ = fmt.Fprintln(params.Stderr, "Error: hook name required")
		_, _ = fmt.Fprintln(params.Stderr, "Usage: stave --hooks run <hook-name> [-- args...]")
		return exitUsage
	}

	hookName := remaining[0]
	hookArgs := parseHookArgs(remaining[1:])

	slog.Debug("hooks run starting",
		slog.String("hook", hookName),
		slog.Any("args", hookArgs))

	// Load configuration
	slog.Debug("loading hooks configuration")
	cfg, err := config.Load(&config.LoadOptions{ProjectDir: params.Dir})
	if err != nil {
		return printConfigErr(params.Stderr, err)
	}

	// Create runtime and execute with real target runner
	runtime := &hooks.Runtime{
		Config:       cfg,
		Stdin:        params.Stdin,
		Stdout:       params.Stdout,
		Stderr:       params.Stderr,
		TargetRunner: newStaveTargetRunner(cfg, params.Dir),
	}

	result, err := runtime.Run(ctx, hookName, hookArgs)
	if err != nil {
		return printErr(params.Stderr, err)
	}

	return result.ExitCode
}

func parseHookArgs(args []string) []string {
	return lo.Without(args, "--")
}

// hooksUsage prints the hooks command usage.
func hooksUsage(w io.Writer) {
	_, _ = fmt.Fprint(w, `
stave --hooks [subcommand]

Manage Git hooks for this repository.

Subcommands:
  init        Show instructions for configuring hooks
  install     Install hook scripts to .git/hooks
  uninstall   Remove Stave-managed hook scripts
  list        List configured hooks and their targets (default)
  run         Execute targets for a specific hook

Flags for install:
  --force     Overwrite existing non-Stave hooks

Flags for uninstall:
  --all       Remove all Stave-managed hooks (not just configured ones)

Environment Variables:
  STAVE_HOOKS=0      Disable all hooks
  STAVE_HOOKS=debug  Enable debug output in hook scripts

Examples:
  stave --hooks                    # List configured hooks
  stave --hooks init               # Show setup instructions
  stave --hooks install            # Install all configured hooks
  stave --hooks install --force    # Overwrite existing hooks
  stave --hooks uninstall          # Remove configured hooks
  stave --hooks uninstall --all    # Remove all Stave hooks
  stave --hooks run pre-commit     # Execute pre-commit targets
`[1:])
}
