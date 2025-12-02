//nolint:gochecknoglobals // Once/mutex patterns.
package dryrun

import "sync"

var (
	// Once-protected variables for whether the user requested dryrun mode.
	dryRunRequestedValue    bool
	dryRunRequestedEnvValue bool
	dryRunRequestedEnvOnce  sync.Once

	// Once-protected variables for whether dryrun mode is possible.
	dryRunPossible     bool
	dryRunPossibleOnce sync.Once
)
