// Package dryrun implements the conditional checks for Stave's dryrun mode.
//
// For IsDryRun() to be true, two things have to be true:
// 1. IsPossible() must be true
//   - This can only happen if the env var `STAVEFILE_DRYRUN_POSSIBLE` was set at the
//     point of the first call to IsPossible()
//
// 2. IsRequested() must be true
//   - This can happen under one of two conditions:
//     i.  The env var `STAVEFILE_DRYRUN` was set at the point of the first call to IsRequested()
//     ii. SetRequested(true) was called at some point prior to the IsPossible() call.
//
// This enables the "top-level" Stave run, which compiles the stavefile into a
// binary, to always be carried out regardless of `-dryrun` (because
// `STAVEFILE_DRYRUN_POSSIBLE` will not be set in that situation), while still
// enabling true dryrun functionality for "inner" Stave runs (i.e., runs of the
// compiled stavefile binary).
package dryrun

import (
	"context"
	"os"
	"os/exec"
)

// RequestedEnv is the environment variable that indicates the user requested dryrun mode when running stave.
const RequestedEnv = "STAVEFILE_DRYRUN"

// PossibleEnv is the environment variable that indicates we are in a context where a dry run is possible.
const PossibleEnv = "STAVEFILE_DRYRUN_POSSIBLE"

// SetRequested sets the dryrun requested state to the specified boolean value.
func SetRequested(value bool) {
	dryRunRequestedValue = value
}

// SetPossible sets the dryrun possible value to the specified boolean value.
func SetPossible(value bool) {
	dryRunPossible = value
}

// IsRequested checks if dry-run mode was requested, either explicitly or via an environment variable.
func IsRequested() bool {
	dryRunRequestedEnvOnce.Do(func() {
		if os.Getenv(RequestedEnv) != "" {
			dryRunRequestedEnvValue = true
		}
	})

	return dryRunRequestedEnvValue || dryRunRequestedValue
}

// IsPossible checks if dry-run mode is supported in the current context.
func IsPossible() bool {
	dryRunPossibleOnce.Do(func() {
		dryRunPossible = dryRunPossible || os.Getenv(PossibleEnv) != ""
	})

	return dryRunPossible
}

// Wrap creates an *exec.Cmd to run a command or simulate it in dry-run mode.
// If not in dry-run mode, it returns exec.Command(cmd, args...).
// In dry-run mode, it returns a command that prints the simulated command.
func Wrap(ctx context.Context, cmd string, args ...string) *exec.Cmd {
	if !IsDryRun() {
		return exec.CommandContext(ctx, cmd, args...)
	}

	// Return an *exec.Cmd that just prints the command that would have been run.
	return exec.CommandContext(ctx, "echo", append([]string{"DRYRUN: " + cmd}, args...)...) //nolint:gosec // It's echo!
}

// IsDryRun determines if dry-run mode is both possible and requested.
func IsDryRun() bool {
	possible := IsPossible()
	requested := IsRequested()

	return possible && requested
}
