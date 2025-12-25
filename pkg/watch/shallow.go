package watch

import (
	"context"

	"github.com/yaklabco/stave/pkg/watch/mode"
	"github.com/yaklabco/stave/pkg/watch/wctx"
)

// The purpose of the functions in this file is to provide access to functions
// in sub-packages from the `watch` package itself. In pure Go, this would
// be unnecessary, as the relevant functions are public anyway.
// But in the context of stave, this simplifies import alias tracking in
// mainfile.gotmpl, as only the `watch` package (and none of its sub-packages)
// needs to be accessible from that code, and so only its alias needs to be
// tracked.

// NOTE: Due to being used only in the template (and in code generated from
// the template), IDEs and other tools may show these functions as unused;
// do not fall for that & accidentally delete them!

func IsOverallWatchMode() bool {
	return mode.IsOverallWatchMode()
}

func AddRequestedTarget(name string) {
	mode.AddRequestedTarget(name)
}

func RegisterContext(name string, ctx context.Context) { //nolint:revive // This is intentional, we are registering a context by
	wctx.Register(name, ctx)
}

func UnregisterContext(name string) {
	wctx.Unregister(name)
}
