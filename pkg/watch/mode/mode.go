package mode

import (
	"strings"
	"sync"
)

var (
	overallWatchMode bool            //nolint:gochecknoglobals // These are intentionally global, and part of a sync.Mutex pattern.
	requestedTargets map[string]bool //nolint:gochecknoglobals // These are intentionally global, and part of a sync.Mutex pattern.
	primaryTarget    string          //nolint:gochecknoglobals // These are intentionally global, and part of a sync.Mutex pattern.
	watchModeMu      sync.Mutex      //nolint:gochecknoglobals // These are intentionally global, and part of a sync.Mutex pattern.
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

// AddRequestedTarget adds a target to the list of requested targets.
func AddRequestedTarget(name string) {
	watchModeMu.Lock()
	defer watchModeMu.Unlock()
	if requestedTargets == nil {
		requestedTargets = make(map[string]bool)
	}
	if primaryTarget == "" {
		primaryTarget = name
	}
	requestedTargets[strings.ToLower(name)] = true
}

// ResetForTest resets the global state for testing purposes.
func ResetForTest() {
	watchModeMu.Lock()
	defer watchModeMu.Unlock()
	overallWatchMode = false
	requestedTargets = nil
	primaryTarget = ""
}

// IsRequestedTarget returns whether the given target name was requested on the command line.
func IsRequestedTarget(name string) bool {
	watchModeMu.Lock()
	defer watchModeMu.Unlock()
	return requestedTargets[strings.ToLower(name)]
}

// GetOutermostTarget returns the name of the primary outermost target.
func GetOutermostTarget() string {
	watchModeMu.Lock()
	defer watchModeMu.Unlock()
	return primaryTarget
}
