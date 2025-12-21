package stctx

import (
	"context"
	"testing"
)

func TestContextHelpers(t *testing.T) {
	ctx := context.Background()

	// Test Target
	name := "test-target"
	ctx = ContextWithTarget(ctx, name)
	if GetCurrentTarget(ctx) != name {
		t.Errorf("expected target %q, got %q", name, GetCurrentTarget(ctx))
	}

	// Test TargetState
	state := &struct{ val string }{"my-state"}
	ctx = ContextWithTargetState(ctx, state)
	if GetTargetState(ctx) != state {
		t.Errorf("expected state %v, got %v", state, GetTargetState(ctx))
	}
}
