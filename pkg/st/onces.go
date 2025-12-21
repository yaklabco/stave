//nolint:gochecknoglobals // Once/mutex patterns.
package st

import "sync"

var onces = &onceMap{
	mu: &sync.Mutex{},
	m:  map[onceKey]*onceFun{},
}

// ResetOnces clears the global map of once-run dependencies.
// This is primarily used in watch mode to allow dependencies to re-run.
func ResetOnces() {
	onces.mu.Lock()
	defer onces.mu.Unlock()
	// fmt.Println("[DEBUG] ResetOnces called")
	onces.m = make(map[onceKey]*onceFun)
}

// ResetSpecificOnces clears specific functions from the global map of once-run dependencies.
func ResetSpecificOnces(fns ...interface{}) {
	onces.mu.Lock()
	defer onces.mu.Unlock()
	for _, fInterface := range fns {
		var theFunc Fn
		if ff, ok := fInterface.(Fn); ok {
			theFunc = ff
		} else {
			theFunc = F(fInterface)
		}
		key := onceKey{
			Name: theFunc.Name(),
			ID:   theFunc.ID(),
		}
		delete(onces.m, key)
	}
}

// ResetOncesByName clears functions from the global map of once-run dependencies by their display names.
func ResetOncesByName(names ...string) {
	onces.mu.Lock()
	defer onces.mu.Unlock()
	nameMap := make(map[string]bool)
	for _, n := range names {
		nameMap[n] = true
	}
	for key := range onces.m {
		if nameMap[DisplayName(key.Name)] {
			delete(onces.m, key)
		}
	}
}
