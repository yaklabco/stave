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

	cblog "github.com/charmbracelet/log"
	"github.com/yaklabco/stave/config"
	"github.com/yaklabco/stave/internal/hooks"
	"github.com/yaklabco/stave/pkg/st"
	"github.com/yaklabco/stave/pkg/stave/prettylog"
)

// Exit codes for CLI commands.
const (
	exitOK    = 0
	exitError = 1
	exitUsage = 2
)

// HooksParams contains parameters for the hooks command.
type HooksParams struct {
	Debug   bool
	Verbose bool
}

// newStaveTargetRunner creates a TargetRunnerFunc that executes targets using stave.Run.
// This wires the hooks runtime to the real Stave execution engine.
func newStaveTargetRunner(cfg *config.Config) hooks.TargetRunnerFunc {
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

			// Propagate config-level settings
			Debug:    cfg.Debug,
			Verbose:  cfg.Verbose,
			HashFast: cfg.HashFast,
			GoCmd:    cfg.GoCmd,
			CacheDir: cfg.CacheDir,

			// Target invocation: prepend target name to args
			Args: append([]string{target}, args...),
		}

		err := Run(runParams)
		return st.ExitStatus(err), err
	}
}

// RunHooksCommand handles the `stave --hooks` subcommand.
// It returns the exit code.
func RunHooksCommand(stdout, stderr io.Writer, args []string) int {
	return RunHooksCommandContext(context.Background(), stdout, stderr, args)
}

// RunHooksCommandWithParams handles the `stave --hooks` subcommand with debug/verbose params.
// It returns the exit code.
func RunHooksCommandWithParams(ctx context.Context, stdout, stderr io.Writer, params HooksParams, args []string) int {
	// Set up pretty logging with appropriate level
	logHandler := prettylog.SetupPrettyLogger(stdout)
	switch {
	case params.Debug:
		logHandler.SetLevel(cblog.DebugLevel)
	case params.Verbose:
		logHandler.SetLevel(cblog.InfoLevel)
	default:
		logHandler.SetLevel(cblog.WarnLevel)
	}

	return runHooksCommandInternal(ctx, stdout, stderr, args)
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

// RunHooksCommandContext handles the `stave --hooks` subcommand with context.
// It returns the exit code.
//
// Deprecated: Use RunHooksCommandWithParams for proper debug/verbose support.
func RunHooksCommandContext(ctx context.Context, stdout, stderr io.Writer, args []string) int {
	// Set up logging with defaults from environment
	logHandler := prettylog.SetupPrettyLogger(stdout)
	switch {
	case st.Debug():
		logHandler.SetLevel(cblog.DebugLevel)
	case st.Verbose():
		logHandler.SetLevel(cblog.InfoLevel)
	default:
		logHandler.SetLevel(cblog.WarnLevel)
	}

	return runHooksCommandInternal(ctx, stdout, stderr, args)
}

// runHooksCommandInternal is the internal implementation of the hooks command.
func runHooksCommandInternal(ctx context.Context, stdout, stderr io.Writer, args []string) int {
	flagSet := flag.NewFlagSet("hooks", flag.ContinueOnError)
	flagSet.SetOutput(stdout)
	flagSet.Usage = func() {
		hooksUsage(stdout)
	}

	if err := flagSet.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return exitOK
		}
		_, _ = fmt.Fprintf(stderr, "Error: %v\n", err)
		return exitUsage
	}

	subArgs := flagSet.Args()
	if len(subArgs) == 0 {
		// No subcommand, show list
		return runHooksList(ctx, stdout, stderr)
	}

	return dispatchHooksSubcommand(ctx, stdout, stderr, subArgs)
}

func dispatchHooksSubcommand(ctx context.Context, stdout, stderr io.Writer, subArgs []string) int {
	subcmd := HooksSubcommand(strings.ToLower(subArgs[0]))

	slog.Debug("hooks subcommand dispatching",
		slog.String("subcommand", string(subcmd)))

	switch subcmd {
	case HooksInit:
		return runHooksInit(ctx, stdout, stderr)
	case HooksInstall:
		return runHooksInstall(ctx, stdout, stderr, subArgs[1:])
	case HooksUninstall:
		return runHooksUninstall(ctx, stdout, stderr, subArgs[1:])
	case HooksList:
		return runHooksList(ctx, stdout, stderr)
	case HooksRun:
		return runHooksRun(ctx, stdout, stderr, subArgs[1:])
	default:
		slog.Debug("unknown hooks subcommand",
			slog.String("subcommand", subArgs[0]))
		_, _ = fmt.Fprintf(stderr, "Error: unknown hooks subcommand %q\n", subArgs[0])
		hooksUsage(stderr)
		return exitUsage
	}
}

// runHooksInit initializes hooks configuration in stave.yaml.
func runHooksInit(ctx context.Context, stdout, stderr io.Writer) int {
	slog.Debug("loading hooks configuration")

	// First ensure config exists
	cfg, err := config.Load(nil)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Error loading config: %v\n", err)
		return exitError
	}

	// Check if hooks are already configured
	if len(cfg.Hooks) > 0 {
		slog.Debug("hooks already configured",
			slog.Int("hook_count", len(cfg.Hooks)))
		_, _ = fmt.Fprintln(stdout, "Hooks configuration already exists in stave.yaml")
		return runHooksList(ctx, stdout, stderr)
	}

	printHooksInitInstructions(stdout)
	return exitOK
}

func printHooksInitInstructions(out io.Writer) {
	_, _ = fmt.Fprintln(out, "Add hooks configuration to your stave.yaml file:")
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, "hooks:")
	_, _ = fmt.Fprintln(out, "  pre-commit:")
	_, _ = fmt.Fprintln(out, "    - target: fmt")
	_, _ = fmt.Fprintln(out, "    - target: lint")
	_, _ = fmt.Fprintln(out, "  pre-push:")
	_, _ = fmt.Fprintln(out, "    - target: test")
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, "Then run: stave --hooks install")
}

// runHooksInstall installs hook scripts to the Git repository.
func runHooksInstall(ctx context.Context, stdout, stderr io.Writer, args []string) int {
	flagSet := flag.NewFlagSet("install", flag.ContinueOnError)
	flagSet.SetOutput(stdout)
	force := flagSet.Bool("force", false, "overwrite existing non-Stave hooks")

	if err := flagSet.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return exitOK
		}
		_, _ = fmt.Fprintf(stderr, "Error: %v\n", err)
		return exitUsage
	}

	slog.Debug("hooks install starting",
		slog.Bool("force", *force))

	// Find Git repository
	repo, err := hooks.FindGitRepoContext(ctx, "")
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Error: %v\n", err)
		if errors.Is(err, hooks.ErrNotGitRepo) {
			_, _ = fmt.Fprintln(stderr, "Run this command from within a Git repository.")
		}
		return exitError
	}

	// Load configuration
	slog.Debug("loading hooks configuration")
	cfg, err := config.Load(nil)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Error loading config: %v\n", err)
		return exitError
	}

	// Check if hooks are configured
	if len(cfg.Hooks) == 0 {
		slog.Debug("no hooks configured in config")
		_, _ = fmt.Fprintln(stderr, "No hooks configured in stave.yaml")
		_, _ = fmt.Fprintln(stderr, "Run 'stave --hooks init' for setup instructions.")
		return exitError
	}

	return installHooks(repo, cfg, *force, stdout, stderr)
}

func installHooks(repo *hooks.GitRepo, cfg *config.Config, force bool, stdout, stderr io.Writer) int {
	// Ensure hooks directory exists
	if err := repo.EnsureHooksDir(); err != nil {
		_, _ = fmt.Fprintf(stderr, "Error creating hooks directory: %v\n", err)
		return exitError
	}

	// Install each configured hook
	hookNames := cfg.Hooks.HookNames()
	slog.Debug("installing hooks",
		slog.Int("hook_count", len(hookNames)),
		slog.String("directory", repo.HooksPath()))

	installed := 0
	for _, hookName := range hookNames {
		if code := installSingleHook(repo, hookName, force, stdout, stderr); code != exitOK {
			return code
		}
		installed++
	}

	slog.Info("hooks installed",
		slog.Int("count", installed),
		slog.String("directory", repo.HooksPath()))

	_, _ = fmt.Fprintf(stdout, "\nInstalled %d hook(s) to %s\n", installed, repo.HooksPath())
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
		_, _ = fmt.Fprintf(stderr, "Error checking %s: %v\n", hookName, err)
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
		_, _ = fmt.Fprintf(stderr, "Error writing %s: %v\n", hookName, err)
		return exitError
	}
	_, _ = fmt.Fprintf(stdout, "Installed %s\n", hookName)
	return exitOK
}

// runHooksUninstall removes Stave-managed hook scripts.
func runHooksUninstall(ctx context.Context, stdout, stderr io.Writer, args []string) int {
	flagSet := flag.NewFlagSet("uninstall", flag.ContinueOnError)
	flagSet.SetOutput(stdout)
	all := flagSet.Bool("all", false, "uninstall all Stave-managed hooks")

	if err := flagSet.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return exitOK
		}
		_, _ = fmt.Fprintf(stderr, "Error: %v\n", err)
		return exitUsage
	}

	// Find Git repository
	repo, err := hooks.FindGitRepoContext(ctx, "")
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Error: %v\n", err)
		return exitError
	}

	// Load configuration for hook names
	cfg, err := config.Load(nil)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Error loading config: %v\n", err)
		return exitError
	}

	hookNames := getHookNamesToUninstall(*all, cfg)
	if len(hookNames) == 0 {
		_, _ = fmt.Fprintln(stdout, "No hooks to uninstall.")
		return exitOK
	}

	return uninstallHooks(repo, hookNames, stdout, stderr)
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
			_, _ = fmt.Fprintf(stderr, "Error removing %s: %v\n", hookName, err)
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
func runHooksList(ctx context.Context, stdout, stderr io.Writer) int {
	slog.Debug("loading hooks configuration for list")

	cfg, err := config.Load(nil)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Error loading config: %v\n", err)
		return exitError
	}

	if len(cfg.Hooks) == 0 {
		slog.Debug("no hooks configured")
		_, _ = fmt.Fprintln(stdout, "No hooks configured.")
		_, _ = fmt.Fprintln(stdout, "Run 'stave --hooks init' for setup instructions.")
		return exitOK
	}

	slog.Debug("listing configured hooks",
		slog.Int("hook_count", len(cfg.Hooks)))

	printConfiguredHooks(cfg, stdout)

	// Check installation status
	repo, repoErr := hooks.FindGitRepoContext(ctx, "")
	if repoErr == nil {
		printInstallationStatus(repo, cfg.Hooks.HookNames(), stdout)
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
func runHooksRun(ctx context.Context, stdout, stderr io.Writer, args []string) int {
	flagSet := flag.NewFlagSet("run", flag.ContinueOnError)
	flagSet.SetOutput(stdout)

	if err := flagSet.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return exitOK
		}
		_, _ = fmt.Fprintf(stderr, "Error: %v\n", err)
		return exitUsage
	}

	remaining := flagSet.Args()
	if len(remaining) == 0 {
		_, _ = fmt.Fprintln(stderr, "Error: hook name required")
		_, _ = fmt.Fprintln(stderr, "Usage: stave --hooks run <hook-name> [-- args...]")
		return exitUsage
	}

	hookName := remaining[0]
	hookArgs := parseHookArgs(remaining[1:])

	slog.Debug("hooks run starting",
		slog.String("hook", hookName),
		slog.Any("args", hookArgs))

	// Load configuration
	slog.Debug("loading hooks configuration")
	cfg, err := config.Load(nil)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Error loading config: %v\n", err)
		return exitError
	}

	// Create runtime and execute with real target runner
	runtime := &hooks.Runtime{
		Config:       cfg,
		Stdout:       stdout,
		Stderr:       stderr,
		TargetRunner: newStaveTargetRunner(cfg),
	}

	result, err := runtime.Run(ctx, hookName, hookArgs)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Error: %v\n", err)
		return exitError
	}

	return result.ExitCode
}

func parseHookArgs(args []string) []string {
	for i, arg := range args {
		if arg == "--" {
			return args[i+1:]
		}
	}
	return nil
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
