package parallelism

import (
	"os"
	"runtime"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApply(t *testing.T) {
	// Save original env vars and GOMAXPROCS
	origStaveNumProcessors, origStaveNumProcessorsOK := os.LookupEnv(StaveNumProcessorsEnvVar)
	origGoMaxProcs, origGoMaxProcsOK := os.LookupEnv(GoMaxProcsEnvVar)
	origMaxProcs := runtime.GOMAXPROCS(0)
	defer func() {
		revertEnvVar(t, origStaveNumProcessors, origStaveNumProcessorsOK)
		revertEnvVar(t, origGoMaxProcs, origGoMaxProcsOK)
		runtime.GOMAXPROCS(origMaxProcs)
	}()

	t.Run("Default", func(t *testing.T) {
		origVal, origValOK := os.LookupEnv(StaveNumProcessorsEnvVar)
		t.Cleanup(func() {
			revertEnvVar(t, origVal, origValOK)
		})

		require.NoError(t, os.Unsetenv(StaveNumProcessorsEnvVar))
		envMap := make(map[string]string)
		err := Apply(envMap)
		require.NoError(t, err)

		expectedNum := runtime.NumCPU()
		assert.Equal(t, strconv.Itoa(expectedNum), envMap[StaveNumProcessorsEnvVar])
		assert.Equal(t, strconv.Itoa(expectedNum), envMap[GoMaxProcsEnvVar])
	})

	t.Run("FromEnv", func(t *testing.T) {
		t.Setenv(StaveNumProcessorsEnvVar, "4")
		envMap := make(map[string]string)
		err := Apply(envMap)
		require.NoError(t, err)

		assert.Equal(t, "4", envMap[StaveNumProcessorsEnvVar])
		assert.Equal(t, "4", envMap[GoMaxProcsEnvVar])
		assert.Equal(t, 4, runtime.GOMAXPROCS(0))
	})

	t.Run("InvalidEnv", func(t *testing.T) {
		t.Setenv(StaveNumProcessorsEnvVar, "invalid")
		envMap := make(map[string]string)
		err := Apply(envMap)
		assert.Error(t, err)
	})
}

func revertEnvVar(t *testing.T, origVal string, origValOK bool) {
	t.Helper()

	if origValOK {
		setenvErr := os.Setenv(StaveNumProcessorsEnvVar, origVal)
		if setenvErr != nil {
			t.Errorf("Failed to revert env var %q: %v", StaveNumProcessorsEnvVar, setenvErr)
		}
	} else {
		unsetenvErr := os.Unsetenv(StaveNumProcessorsEnvVar)
		if unsetenvErr != nil {
			t.Errorf("Failed to revert env var %q: %v", GoMaxProcsEnvVar, unsetenvErr)
		}
	}
}
