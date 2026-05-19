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

const (
	HookNameApplypatchMsg       = "applypatch-msg"
	HookNamePreApplypatch       = "pre-applypatch"
	HookNamePostApplypatch      = "post-applypatch"
	HookNamePreCommit           = "pre-commit"
	HookNamePrepareCommitMsg    = "prepare-commit-msg"
	HookNameCommitMsg           = "commit-msg"
	HookNamePostCommit          = "post-commit"
	HookNamePreRebase           = "pre-rebase"
	HookNamePostCheckout        = "post-checkout"
	HookNamePostMerge           = "post-merge"
	HookNamePrePush             = "pre-push"
	HookNamePreReceive          = "pre-receive"
	HookNameUpdate              = "update"
	HookNamePostReceive         = "post-receive"
	HookNamePostUpdate          = "post-update"
	HookNamePushToCheckout      = "push-to-checkout"
	HookNamePreAutoGc           = "pre-auto-gc"
	HookNamePostRewrite         = "post-rewrite"
	HookNameSendemailValidate   = "sendemail-validate"
	HookNameFsmonitorWatchman   = "fsmonitor-watchman"
	HookNameP4PreSubmit         = "p4-pre-submit"
	HookNameP4Changelist        = "p4-changelist"
	HookNameP4PrepareChangelist = "p4-prepare-changelist"
	HookNameP4PostChangelist    = "p4-post-changelist"
	HookNamePostIndexChange     = "post-index-change"
)

// knownGitHooks is the set of standard Git hook names.
// Used to warn on unrecognized hook names.
//
//nolint:gochecknoglobals // package-level lookup table for hook validation
var knownGitHooks = lo.Keyify([]string{
	HookNameApplypatchMsg,
	HookNamePreApplypatch,
	HookNamePostApplypatch,
	HookNamePreCommit,
	HookNamePrepareCommitMsg,
	HookNameCommitMsg,
	HookNamePostCommit,
	HookNamePreRebase,
	HookNamePostCheckout,
	HookNamePostMerge,
	HookNamePrePush,
	HookNamePreReceive,
	HookNameUpdate,
	HookNamePostReceive,
	HookNamePostUpdate,
	HookNamePushToCheckout,
	HookNamePreAutoGc,
	HookNamePostRewrite,
	HookNameSendemailValidate,
	HookNameFsmonitorWatchman,
	HookNameP4PreSubmit,
	HookNameP4Changelist,
	HookNameP4PrepareChangelist,
	HookNameP4PostChangelist,
	HookNamePostIndexChange,
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
