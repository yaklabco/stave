package hooks

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/yaklabco/stave/config"
)

// mockRunner creates a TargetRunnerFunc that returns the specified exit codes
// in order for each invocation.
func mockRunner(exitCodes ...int) TargetRunnerFunc {
	idx := 0
	return func(_ context.Context, _ string, _ []string, _ io.Reader, _, _ io.Writer) (int, error) {
		if idx >= len(exitCodes) {
			return 0, nil
		}
		code := exitCodes[idx]
		idx++
		return code, nil
	}
}

// mockRunnerWithError creates a TargetRunnerFunc that returns an error.
func mockRunnerWithError(err error) TargetRunnerFunc {
	return func(_ context.Context, _ string, _ []string, _ io.Reader, _, _ io.Writer) (int, error) {
		return 0, err
	}
}

// mockRunnerCapture creates a TargetRunnerFunc that captures the target name and args.
type targetCall struct {
	target string
	args   []string
}

func mockRunnerCapture(calls *[]targetCall) TargetRunnerFunc {
	return func(_ context.Context, target string, args []string, _ io.Reader, _, _ io.Writer) (int, error) {
		*calls = append(*calls, targetCall{target: target, args: args})
		return 0, nil
	}
}

func TestRuntime_Run_HooksDisabled(t *testing.T) {
	t.Setenv(EnvStaveHooks, "0")

	var stderr bytes.Buffer
	runtime := &Runtime{
		Config: &config.Config{
			Hooks: config.HooksConfig{
				"pre-commit": {
					{Target: "fmt"},
				},
			},
		},
		Stderr: &stderr,
	}

	result, err := runtime.Run(context.Background(), "pre-commit", nil)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if !result.Disabled {
		t.Error("Result.Disabled = false, want true")
	}
	if !result.Success() {
		t.Error("Result should be successful when hooks are disabled")
	}
	if len(result.Targets) != 0 {
		t.Error("No targets should be run when hooks are disabled")
	}
}

func TestRuntime_Run_NoConfig(t *testing.T) {
	t.Parallel()

	runtime := &Runtime{
		Config: nil,
	}

	result, err := runtime.Run(context.Background(), "pre-commit", nil)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if !result.Success() {
		t.Error("Result should be successful with no config")
	}
}

func TestRuntime_Run_NoHooksConfig(t *testing.T) {
	t.Parallel()

	runtime := &Runtime{
		Config: &config.Config{
			Hooks: nil,
		},
	}

	result, err := runtime.Run(context.Background(), "pre-commit", nil)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if !result.Success() {
		t.Error("Result should be successful with no hooks config")
	}
}

func TestRuntime_Run_UnconfiguredHook(t *testing.T) {
	t.Parallel()

	runtime := &Runtime{
		Config: &config.Config{
			Hooks: config.HooksConfig{
				"pre-push": {
					{Target: "test"},
				},
			},
		},
	}

	result, err := runtime.Run(context.Background(), "pre-commit", nil)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if !result.Success() {
		t.Error("Result should be successful for unconfigured hook")
	}
	if len(result.Targets) != 0 {
		t.Error("No targets should be run for unconfigured hook")
	}
}

func TestRuntime_Run_SingleTarget_Success(t *testing.T) {
	t.Parallel()

	runtime := &Runtime{
		Config: &config.Config{
			Hooks: config.HooksConfig{
				"pre-commit": {
					{Target: "fmt"},
				},
			},
		},
		TargetRunner: mockRunner(0),
	}

	result, err := runtime.Run(context.Background(), "pre-commit", nil)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if !result.Success() {
		t.Error("Result should be successful")
	}
	if len(result.Targets) != 1 {
		t.Fatalf("Expected 1 target result, got %d", len(result.Targets))
	}
	if result.Targets[0].Name != "fmt" {
		t.Errorf("Target name = %q, want %q", result.Targets[0].Name, "fmt")
	}
	if result.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", result.ExitCode)
	}
}

func TestRuntime_Run_SingleTarget_Failure(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer
	runtime := &Runtime{
		Config: &config.Config{
			Hooks: config.HooksConfig{
				"pre-commit": {
					{Target: "lint"},
				},
			},
		},
		Stderr:       &stderr,
		TargetRunner: mockRunner(1),
	}

	result, err := runtime.Run(context.Background(), "pre-commit", nil)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if result.Success() {
		t.Error("Result should not be successful")
	}
	if result.ExitCode != 1 {
		t.Errorf("ExitCode = %d, want 1", result.ExitCode)
	}
	if len(result.Targets) != 1 {
		t.Fatalf("Expected 1 target result, got %d", len(result.Targets))
	}
	if result.Targets[0].ExitCode != 1 {
		t.Errorf("Target ExitCode = %d, want 1", result.Targets[0].ExitCode)
	}
}

func TestRuntime_Run_MultipleTargets_AllPass(t *testing.T) {
	t.Parallel()

	runtime := &Runtime{
		Config: &config.Config{
			Hooks: config.HooksConfig{
				"pre-commit": {
					{Target: "fmt"},
					{Target: "lint"},
					{Target: "vet"},
				},
			},
		},
		TargetRunner: mockRunner(0, 0, 0),
	}

	result, err := runtime.Run(context.Background(), "pre-commit", nil)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if !result.Success() {
		t.Error("Result should be successful")
	}
	if len(result.Targets) != 3 {
		t.Fatalf("Expected 3 target results, got %d", len(result.Targets))
	}
	for i, tr := range result.Targets {
		if !tr.Success() {
			t.Errorf("Target %d should be successful", i)
		}
	}
}

func TestRuntime_Run_MultipleTargets_FailFast(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer
	runtime := &Runtime{
		Config: &config.Config{
			Hooks: config.HooksConfig{
				"pre-commit": {
					{Target: "fmt"},
					{Target: "lint"},
					{Target: "vet"},
				},
			},
		},
		Stderr:       &stderr,
		TargetRunner: mockRunner(0, 1, 0), // lint fails
	}

	result, err := runtime.Run(context.Background(), "pre-commit", nil)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if result.Success() {
		t.Error("Result should not be successful")
	}
	// Should stop at lint (second target), not run vet
	if len(result.Targets) != 2 {
		t.Fatalf("Expected 2 target results (fail-fast), got %d", len(result.Targets))
	}
	if result.Targets[0].Name != "fmt" {
		t.Errorf("First target = %q, want %q", result.Targets[0].Name, "fmt")
	}
	if result.Targets[1].Name != "lint" {
		t.Errorf("Second target = %q, want %q", result.Targets[1].Name, "lint")
	}
}

func TestRuntime_Run_WithArgs(t *testing.T) {
	t.Parallel()

	var calls []targetCall
	runtime := &Runtime{
		Config: &config.Config{
			Hooks: config.HooksConfig{
				"pre-push": {
					{Target: "test", Args: []string{"--short"}},
				},
			},
		},
		TargetRunner: mockRunnerCapture(&calls),
	}

	hookArgs := []string{"origin", "main"}
	_, err := runtime.Run(context.Background(), "pre-push", hookArgs)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if len(calls) != 1 {
		t.Fatalf("Expected 1 call, got %d", len(calls))
	}

	expectedArgs := []string{"--short", "origin", "main"}
	if len(calls[0].args) != len(expectedArgs) {
		t.Fatalf("Args length = %d, want %d", len(calls[0].args), len(expectedArgs))
	}
	for i, arg := range expectedArgs {
		if calls[0].args[i] != arg {
			t.Errorf("Args[%d] = %q, want %q", i, calls[0].args[i], arg)
		}
	}
}

func TestRuntime_Run_WithError(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer
	testErr := errors.New("target execution failed")
	runtime := &Runtime{
		Config: &config.Config{
			Hooks: config.HooksConfig{
				"pre-commit": {
					{Target: "fmt"},
				},
			},
		},
		Stderr:       &stderr,
		TargetRunner: mockRunnerWithError(testErr),
	}

	result, err := runtime.Run(context.Background(), "pre-commit", nil)
	if err != nil {
		t.Fatalf("Run() error = %v (should not propagate target errors)", err)
	}

	if result.Success() {
		t.Error("Result should not be successful when runner returns error")
	}
	if result.ExitCode != 1 {
		t.Errorf("ExitCode = %d, want 1", result.ExitCode)
	}
}

func TestIsHooksDisabled(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{"disabled", "0", true},
		{"enabled", "1", false},
		{"empty", "", false},
		{"debug", "debug", false},
		{"random", "yes", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(EnvStaveHooks, tc.value)

			got := IsHooksDisabled()
			if got != tc.want {
				t.Errorf("IsHooksDisabled() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestIsDebugMode(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{"debug lowercase", "debug", true},
		{"debug mixed case", "Debug", true},
		{"debug uppercase", "DEBUG", true},
		{"disabled", "0", false},
		{"enabled", "1", false},
		{"empty", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(EnvStaveHooks, tc.value)

			got := IsDebugMode()
			if got != tc.want {
				t.Errorf("IsDebugMode() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestNewRuntime(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Hooks: config.HooksConfig{
			"pre-commit": {
				{Target: "fmt"},
			},
		},
	}

	runtime := NewRuntime(cfg)

	if runtime.Config != cfg {
		t.Error("Config should be set")
	}
	if runtime.Stdout == nil {
		t.Error("Stdout should be set")
	}
	if runtime.Stderr == nil {
		t.Error("Stderr should be set")
	}
}

func TestTargetResult_Success(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		result   TargetResult
		expected bool
	}{
		{
			name:     "success",
			result:   TargetResult{ExitCode: 0, Error: nil},
			expected: true,
		},
		{
			name:     "non-zero exit",
			result:   TargetResult{ExitCode: 1, Error: nil},
			expected: false,
		},
		{
			name:     "error with zero exit",
			result:   TargetResult{ExitCode: 0, Error: errors.New("failed")},
			expected: false,
		},
		{
			name:     "error with non-zero exit",
			result:   TargetResult{ExitCode: 1, Error: errors.New("failed")},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.result.Success(); got != tt.expected {
				t.Errorf("Success() = %v, want %v", got, tt.expected)
			}
		})
	}
}
