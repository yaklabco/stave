package st

import (
	"testing"
)

func TestDependencyPanic(t *testing.T) {
	panicCount := 0
	panicker := func() {
		panicCount++
		panic("boom")
	}

	// First call - should panic
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic on first call, but got none")
			}
		}()
		Deps(panicker)
	}()

	if panicCount != 1 {
		t.Errorf("Expected panicCount 1, got %d", panicCount)
	}

	// Second call - should also panic, ensuring the fix for panic propagation works
	func() {
		defer func() {
			if r := recover(); r == nil {
				// If this fails, the fix is broken
				t.Error("Expected panic on second call (re-propagation), but got none")
			}
		}()
		Deps(panicker)
	}()
}
