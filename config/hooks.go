package config

import (
	"fmt"
	"sort"
	"strings"

	"github.com/samber/lo"
)

// HookTarget represents a single target to run for a Git hook.
type HookTarget struct {
	// Target is the name of the Stave target to run.
	Target string `mapstructure:"target"`

	// Args are additional CLI arguments passed to the target invocation.
	Args []string `mapstructure:"args,omitempty"`

	// WorkDir is the working directory for the target invocation; if empty, current dir is assumed.
	WorkDir string `mapstructure:"workdir,omitempty"`
}

// HooksConfig maps Git hook names to their configured targets.
type HooksConfig map[string][]HookTarget

// knownGitHooks is the set of standard Git hook names.
// Used to warn on unrecognized hook names.
//
//nolint:gochecknoglobals // package-level lookup table for hook validation
var knownGitHooks = lo.Keyify([]string{
	"applypatch-msg",
	"pre-applypatch",
	"post-applypatch",
	"pre-commit",
	"prepare-commit-msg",
	"commit-msg",
	"post-commit",
	"pre-rebase",
	"post-checkout",
	"post-merge",
	"pre-push",
	"pre-receive",
	"update",
	"post-receive",
	"post-update",
	"push-to-checkout",
	"pre-auto-gc",
	"post-rewrite",
	"sendemail-validate",
	"fsmonitor-watchman",
	"p4-pre-submit",
	"p4-changelist",
	"p4-prepare-changelist",
	"p4-post-changelist",
	"post-index-change",
})

// IsKnownGitHook returns true if the hook name is a recognized Git hook.
func IsKnownGitHook(name string) bool {
	return lo.HasKey(knownGitHooks, name)
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
