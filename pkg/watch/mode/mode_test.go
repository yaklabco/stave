package mode

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWatchMode(t *testing.T) {
	defer ResetForTest()

	t.Run("InitialState", func(t *testing.T) {
		ResetForTest()
		assert.False(t, IsOverallWatchMode())
		assert.Empty(t, GetOutermostTarget())
		assert.False(t, IsRequestedTarget("target1"))
	})

	t.Run("SetOverallWatchMode", func(t *testing.T) {
		ResetForTest()
		SetOverallWatchMode(true)
		assert.True(t, IsOverallWatchMode())
		SetOverallWatchMode(false)
		assert.False(t, IsOverallWatchMode())
	})

	t.Run("RequestedTargets", func(t *testing.T) {
		ResetForTest()
		AddRequestedTarget("Target1")
		assert.True(t, IsRequestedTarget("Target1"))
		assert.True(t, IsRequestedTarget("target1"))
		assert.Equal(t, "Target1", GetOutermostTarget())

		AddRequestedTarget("Target2")
		assert.True(t, IsRequestedTarget("target2"))
		assert.Equal(t, "Target1", GetOutermostTarget()) // First one added is primary
	})
}
