package config

import (
	"fmt"
	"sort"
	"strings"
)

// HookTarget represents a single target to run for a Git hook.
type HookTarget struct {
	// Target is the name of the Stave target to run.
	Target string `mapstructure:"target"`

	// Args are additional CLI arguments passed to the target invocation.
	Args []string `mapstructure:"args,omitempty"`
}

// HooksConfig maps Git hook names to their configured targets.
type HooksConfig map[string][]HookTarget

// knownGitHooks is the set of standard Git hook names.
// Used to warn on unrecognized hook names.
//
//nolint:gochecknoglobals // package-level lookup table for hook validation
var knownGitHooks = map[string]bool{
	"applypatch-msg":        true,
	"pre-applypatch":        true,
	"post-applypatch":       true,
	"pre-commit":            true,
	"prepare-commit-msg":    true,
	"commit-msg":            true,
	"post-commit":           true,
	"pre-rebase":            true,
	"post-checkout":         true,
	"post-merge":            true,
	"pre-push":              true,
	"pre-receive":           true,
	"update":                true,
	"post-receive":          true,
	"post-update":           true,
	"push-to-checkout":      true,
	"pre-auto-gc":           true,
	"post-rewrite":          true,
	"sendemail-validate":    true,
	"fsmonitor-watchman":    true,
	"p4-pre-submit":         true,
	"p4-changelist":         true,
	"p4-prepare-changelist": true,
	"p4-post-changelist":    true,
	"post-index-change":     true,
}

// IsKnownGitHook returns true if the hook name is a recognized Git hook.
func IsKnownGitHook(name string) bool {
	return knownGitHooks[name]
}

// ValidateHooks validates the hooks configuration and returns any errors or warnings.
func ValidateHooks(hooks HooksConfig) ValidationResults {
	var result ValidationResults

	// Iterate in sorted order for deterministic output
	hookNames := hooks.HookNames()
	for _, hookName := range hookNames {
		targets := hooks[hookName]
		// Validate hook name is not empty
		if strings.TrimSpace(hookName) == "" {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "hooks",
				Message: "hook name cannot be empty",
			})
			continue
		}

		// Warn on unrecognized hook names (non-blocking)
		if !IsKnownGitHook(hookName) {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Field:   "hooks." + hookName,
				Message: fmt.Sprintf("unrecognized Git hook name %q", hookName),
			})
		}

		// Validate each target in the hook
		for i, target := range targets {
			if strings.TrimSpace(target.Target) == "" {
				result.Errors = append(result.Errors, ValidationError{
					Field:   fmt.Sprintf("hooks.%s[%d].target", hookName, i),
					Message: "target name cannot be empty",
				})
			}
		}
	}

	return result
}

// Get returns the targets configured for the given hook name.
// Returns nil if the hook is not configured.
func (h HooksConfig) Get(hookName string) []HookTarget {
	if h == nil {
		return nil
	}
	return h[hookName]
}

// HookNames returns all configured hook names in sorted order.
func (h HooksConfig) HookNames() []string {
	if h == nil {
		return nil
	}
	names := make([]string, 0, len(h))
	for name := range h {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// KnownGitHookNames returns all known Git hook names in sorted order.
func KnownGitHookNames() []string {
	names := make([]string, 0, len(knownGitHooks))
	for name := range knownGitHooks {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
