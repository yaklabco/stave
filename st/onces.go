//nolint:gochecknoglobals // Once/mutex patterns.
package st

import "sync"

var onces = &onceMap{
	mu: &sync.Mutex{},
	m:  map[onceKey]*onceFun{},
}
