package sh

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/stave/pkg/watch/wctx"
)

func TestRunRespectsContext(t *testing.T) {
	// Create a context that will be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	// Register it for this test function
	// We need to know what wctx.GetActive() will see.
	// It uses runtime.Callers and DisplayName.

	// Let's just mock the registration for a name we know will be on the stack.
	name := wctx.DisplayName("github.com/yaklabco/stave/pkg/sh.TestRunRespectsContext")
	wctx.Register(name, ctx)
	defer wctx.Unregister(name)

	// Now call sh.Run. It should use the cancelled context and fail immediately.
	err := Run("echo", "should fail")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}
