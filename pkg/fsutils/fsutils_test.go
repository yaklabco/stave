package fsutils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTruePath(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "stave-fsutils-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a real file
	realFile := filepath.Join(tempDir, "realfile")
	err = os.WriteFile(realFile, []byte("hello"), 0644)
	require.NoError(t, err)

	// Get absolute path of real file
	absRealFile, err := filepath.Abs(realFile)
	require.NoError(t, err)

	// Test with real file
	path, err := TruePath(realFile)
	require.NoError(t, err)
	// On macOS, /var is a symlink to /private/var. EvalSymlinks resolves this.
	// We should compare against the resolved version of absRealFile.
	resolvedAbsRealFile, err := filepath.EvalSymlinks(absRealFile)
	require.NoError(t, err)
	assert.Equal(t, resolvedAbsRealFile, path)

	// Create a symlink
	symlink := filepath.Join(tempDir, "symlink")
	err = os.Symlink(realFile, symlink)
	require.NoError(t, err)

	// Test with symlink
	path, err = TruePath(symlink)
	require.NoError(t, err)
	assert.Equal(t, resolvedAbsRealFile, path)

	// Create a nested symlink
	nestedSymlink := filepath.Join(tempDir, "nested-symlink")
	err = os.Symlink(symlink, nestedSymlink)
	require.NoError(t, err)

	// Test with nested symlink
	path, err = TruePath(nestedSymlink)
	require.NoError(t, err)
	assert.Equal(t, resolvedAbsRealFile, path)
}

func TestTruePath_NonExistent(t *testing.T) {
	path, err := TruePath("/non/existent/path/that/really/should/not/exist")
	require.Error(t, err)
	assert.Empty(t, path)
}

func TestMustRead(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "stave-fsutils-mustread-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	file := filepath.Join(tempDir, "file")
	content := []byte("test content")
	err = os.WriteFile(file, content, 0644)
	require.NoError(t, err)

	assert.NotPanics(t, func() {
		data := MustRead(file)
		assert.Equal(t, content, data)
	})

	assert.Panics(t, func() {
		MustRead(filepath.Join(tempDir, "non-existent"))
	})
}
