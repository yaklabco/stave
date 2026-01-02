package wtarget

import (
	"context"
	"sync"

	"github.com/gobwas/glob"
)

type Target struct {
	Name        string
	Patterns    []string
	Globs       []glob.Glob
	Deps        []any
	Watchers    []string
	DepIDs      map[string]struct{}
	CancelFuncs []context.CancelFunc
	Mu          sync.Mutex
	RerunChan   chan struct{}
}
