package wstack

import (
	"runtime"
	"strings"

	"github.com/yaklabco/stave/pkg/stack"
	"github.com/yaklabco/stave/pkg/watch/wctx"
)

func CallerTargetName() string {
	pcs := make([]uintptr, stack.MaxStackDepthToCheck)
	n := runtime.Callers(3, pcs)
	if n == 0 {
		return ""
	}
	frames := runtime.CallersFrames(pcs[:n])
	for {
		frame, more := frames.Next()
		name := frame.Function
		// Skip internal watch package functions, but allow Test functions for testing.
		if strings.HasPrefix(name, "github.com/yaklabco/stave/pkg/watch.") && !strings.Contains(name, ".Test") {
			if !more {
				break
			}
			continue
		}
		return wctx.DisplayName(name)
	}
	return ""
}
