package wctx

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yaklabco/stave/pkg/watch/wtarget"
)

func TestRegister(t *testing.T) {
	ctx := context.Background()
	name := "test-target"
	Register(name, ctx)
	defer Unregister(name)

	assert.Equal(t, ctx, Get(name))
	Unregister(name)
	assert.Nil(t, Get(name))
}

func TestWithCurrent(t *testing.T) {
	ctx := context.Background()
	name := "test-target"
	ctx = WithCurrent(ctx, name)
	assert.Equal(t, name, GetCurrent(ctx))
}

func TestWithConfig(t *testing.T) {
	ctx := context.Background()
	target := &wtarget.Target{Name: "test-target"}
	ctx = WithConfig(ctx, target)
	assert.Equal(t, target, GetConfig(ctx))
}

func TestDisplayName(t *testing.T) {
	assert.Equal(t, "target:name", DisplayName("main.target.name"))
	assert.Equal(t, "other:package:target", DisplayName("other.package.target"))
}

func TestGetActive(t *testing.T) {
	ctx := context.WithValue(t.Context(), "key", "value") //nolint:revive,staticcheck // String-as-context-key, but fine for the purposes of this test.
	name := "github.com/yaklabco/stave/pkg/watch/wctx.TestGetActive"
	// DisplayName will convert it
	displayName := DisplayName(name)
	Register(displayName, ctx)
	defer Unregister(displayName)

	activeCtx := GetActive()
	assert.Equal(t, ctx, activeCtx)
	assert.Equal(t, "value", activeCtx.Value("key"))
}
