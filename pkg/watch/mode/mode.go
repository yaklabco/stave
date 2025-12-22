package mode

import "sync"

var (
	overallWatchMode bool       //nolint:gochecknoglobals // These are intentionally global, and part of a sync.Mutex pattern.
	outermostTarget  string     //nolint:gochecknoglobals // These are intentionally global, and part of a sync.Mutex pattern.
	watchModeMu      sync.Mutex //nolint:gochecknoglobals // These are intentionally global, and part of a sync.Mutex pattern.
)

// SetOverallWatchMode sets whether we are in overall watch mode.
func SetOverallWatchMode(b bool) {
	watchModeMu.Lock()
	overallWatchMode = b
	watchModeMu.Unlock()
}

// IsOverallWatchMode returns whether we are in overall watch mode.
func IsOverallWatchMode() bool {
	watchModeMu.Lock()
	defer watchModeMu.Unlock()
	return overallWatchMode
}

// SetOutermostTarget sets the name of the outermost target.
func SetOutermostTarget(name string) {
	watchModeMu.Lock()
	outermostTarget = name
	watchModeMu.Unlock()
}

// GetOutermostTarget returns the name of the outermost target.
func GetOutermostTarget() string {
	watchModeMu.Lock()
	defer watchModeMu.Unlock()
	return outermostTarget
}
