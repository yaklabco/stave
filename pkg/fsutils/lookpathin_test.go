package fsutils

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLookPathIn(t *testing.T) {
	tempDir := t.TempDir()

	// Create a sub-directory for our test commands
	binDir := filepath.Join(tempDir, "bin")
	err := os.Mkdir(binDir, 0755)
	require.NoError(t, err)

	cmdName := "my-test-cmd"
	if runtime.GOOS == "windows" {
		cmdName += ".exe"
	}
	cmdPath := filepath.Join(binDir, cmdName)

	// Create a dummy executable
	err = os.WriteFile(cmdPath, []byte("dummy"), 0755)
	require.NoError(t, err)

	t.Run("CommandFoundInPath", func(t *testing.T) {
		path, err := LookPathIn("my-test-cmd", binDir)
		require.NoError(t, err)
		assert.Equal(t, cmdPath, path)
	})

	t.Run("CommandNotFoundInPath", func(t *testing.T) {
		_, err := LookPathIn("non-existent-cmd", binDir)
		require.Error(t, err)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("EmptyPathList", func(t *testing.T) {
		_, err := LookPathIn("my-test-cmd", "")
		assert.Error(t, err)
	})

	t.Run("MultiplePaths", func(t *testing.T) {
		otherDir := filepath.Join(tempDir, "other")
		err := os.Mkdir(otherDir, 0755)
		require.NoError(t, err)

		pathList := strings.Join([]string{otherDir, binDir}, string(os.PathListSeparator))
		path, err := LookPathIn("my-test-cmd", pathList)
		require.NoError(t, err)
		assert.Equal(t, cmdPath, path)
	})
}
