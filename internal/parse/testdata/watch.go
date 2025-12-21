//go:build stave

package main

import (
	"github.com/yaklabco/stave/pkg/watch"
)

func WatchTarget() {
	watch.Watch("*.go")
}

func NonWatchTarget() {
}

func WatchDepsTarget() {
	watch.Deps(NonWatchTarget)
}
