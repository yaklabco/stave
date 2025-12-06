package hooks

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/yaklabco/stave/config"
	"github.com/yaklabco/stave/internal/log"
	"github.com/yaklabco/stave/pkg/st"
)

// Environment variable names for hooks control.
const (
	EnvStaveHooks = "STAVE_HOOKS"
)

// ErrHooksDisabled is returned when hooks are disabled via STAVE_HOOKS=0.
var ErrHooksDisabled = errors.New("hooks disabled via STAVE_HOOKS=0")

// Runtime executes hook targets.
type Runtime struct {
	// Config is the Stave configuration containing hook definitions.
	Config *config.Config

	// Stdout is where target output is written.
	Stdout io.Writer

	// Stderr is where error messages are written.
	Stderr io.Writer

	// TargetRunner is the function that runs a Stave target.
	// Production code should always set this; if nil, a no-op test stub is used.
	TargetRunner TargetRunnerFunc
}

// TargetRunnerFunc runs a Stave target and returns its exit code.
type TargetRunnerFunc func(
	ctx context.Context,
	target string,
	args []string,
	stdin io.Reader,
	stdout, stderr io.Writer,
) (int, error)

// RunResult holds the outcome of running a hook.
type RunResult struct {
	// Hook is the name of the hook that was run.
	Hook string

	// Targets contains the results for each target that was executed.
	Targets []TargetResult

	// ExitCode is the overall exit code (0 for success, first non-zero for failure).
	ExitCode int

	// TotalTime is the total duration of hook execution.
	TotalTime time.Duration

	// Disabled is true if hooks were disabled via environment variable.
	Disabled bool
}

// TargetResult holds the result of running a single target.
type TargetResult struct {
	// Name is the target name.
	Name string

	// Args are the arguments passed to the target.
	Args []string

	// ExitCode is the exit code from running the target.
	ExitCode int

	// Duration is how long the target took to run.
	Duration time.Duration

	// Error is any error that occurred (may be nil even with non-zero exit).
	Error error
}

// Success returns true if the target completed successfully.
func (r TargetResult) Success() bool {
	return r.ExitCode == 0 && r.Error == nil
}

// Success returns true if the hook completed successfully (all targets passed or hooks disabled).
func (r RunResult) Success() bool {
	return r.ExitCode == 0
}

// Run executes all configured targets for the given hook.
// It returns a RunResult with details about the execution.
//
// Behavior:
//   - If STAVE_HOOKS=0, returns immediately with success (Disabled=true).
//   - If no targets are configured for the hook, returns success.
//   - Executes targets sequentially in order.
//   - Stops on first failure (fail-fast).
func (r *Runtime) Run(ctx context.Context, hookName string, args []string) (*RunResult, error) {
	startTime := time.Now()
	result := &RunResult{
		Hook:    hookName,
		Targets: []TargetResult{},
	}

	// Check if hooks are disabled
	if IsHooksDisabled() {
		slog.Debug("hooks disabled via environment",
			slog.String("hook", hookName),
			slog.String("env", EnvStaveHooks))
		result.Disabled = true
		result.TotalTime = time.Since(startTime)
		if r.Stderr != nil {
			_, _ = fmt.Fprintf(r.Stderr, "stave: hooks disabled (STAVE_HOOKS=0)\n")
		}
		return result, nil
	}

	// Get hook configuration
	if r.Config == nil || r.Config.Hooks == nil {
		slog.Debug("no hooks configured",
			slog.String("hook", hookName))
		result.TotalTime = time.Since(startTime)
		return result, nil
	}

	targets := r.Config.Hooks.Get(hookName)
	if len(targets) == 0 {
		slog.Debug("no targets for hook",
			slog.String("hook", hookName))
		result.TotalTime = time.Since(startTime)
		return result, nil
	}

	slog.Debug("hook execution starting",
		slog.String("hook", hookName),
		slog.Int("target_count", len(targets)))

	// Ensure we have a target runner
	runner := r.TargetRunner
	if runner == nil {
		runner = defaultTargetRunner
	}

	// Execute each target sequentially
	for _, target := range targets {
		targetStart := time.Now()

		// Combine configured args with any args passed to the hook.
		// Copy to avoid mutating the original slice.
		targetArgs := make([]string, 0, len(target.Args)+len(args))
		targetArgs = append(targetArgs, target.Args...)
		targetArgs = append(targetArgs, args...)

		slog.Debug("target starting",
			slog.String("hook", hookName),
			slog.String("target", target.Target),
			slog.Any("args", targetArgs))

		// Determine stdin handling
		var stdin io.Reader
		if target.PassStdin {
			stdin = os.Stdin
		}

		// Run the target
		exitCode, err := runner(ctx, target.Target, targetArgs, stdin, r.Stdout, r.Stderr)

		targetResult := TargetResult{
			Name:     target.Target,
			Args:     targetArgs,
			ExitCode: exitCode,
			Duration: time.Since(targetStart),
			Error:    err,
		}
		result.Targets = append(result.Targets, targetResult)

		slog.Debug("target completed",
			slog.String("target", target.Target),
			slog.Int("exit_code", exitCode),
			slog.Duration("duration", targetResult.Duration))

		// Fail-fast on first failure
		if exitCode != 0 || err != nil {
			result.ExitCode = exitCode
			if result.ExitCode == 0 && err != nil {
				result.ExitCode = 1
			}
			result.TotalTime = time.Since(startTime)

			// Print failure message
			if r.Stderr != nil {
				_, _ = fmt.Fprintf(r.Stderr, "stave: hook %s failed at target %s (exit %d)\n",
					hookName, target.Target, result.ExitCode)
			}
			return result, nil
		}
	}

	result.TotalTime = time.Since(startTime)

	if st.Verbose() {
		log.SimpleConsoleLogger.Printf("Hook completed: %s (%d targets, %v)",
			hookName, len(result.Targets), result.TotalTime)
	}

	return result, nil
}

// IsHooksDisabled returns true if hooks are disabled via STAVE_HOOKS=0.
func IsHooksDisabled() bool {
	val := os.Getenv(EnvStaveHooks)
	return val == "0"
}

// IsDebugMode returns true if STAVE_HOOKS=debug.
func IsDebugMode() bool {
	val := os.Getenv(EnvStaveHooks)
	return strings.ToLower(val) == "debug"
}

// defaultTargetRunner is a no-op stub for testing purposes only.
// Production code should always inject a real runner via Runtime.TargetRunner.
func defaultTargetRunner(_ context.Context, _ string, _ []string, _ io.Reader, _, _ io.Writer) (int, error) {
	return 0, nil
}

// NewRuntime creates a new Runtime with the given configuration.
func NewRuntime(cfg *config.Config) *Runtime {
	return &Runtime{
		Config: cfg,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
}
