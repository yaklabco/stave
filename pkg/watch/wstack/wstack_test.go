package wstack

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yaklabco/stave/pkg/watch/wctx"
)

func TestCallerTargetName(t *testing.T) {
	wrapper := func() string { //nolint:gocritic // Intentionally ignoring `unlambda` linter complaint here.
		return CallerTargetName()
	}

	name := wrapper()

	assert.Equal(t, wctx.DisplayName("github.com/yaklabco/stave/pkg/watch/wstack.TestCallerTargetName"), name)
}
