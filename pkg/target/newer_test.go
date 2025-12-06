package target

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// testFilePermission is the permission mode for test files.
const testFilePermission = 0o644

// setupTestDir creates a temp dir with test files and returns cleanup function.
func setupTestDir(t *testing.T, files []string) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "")
	require.NoError(t, err, "creating temp dir")
	t.Cleanup(func() { _ = os.RemoveAll(dir) })

	for _, name := range files {
		out := filepath.Join(dir, name)
		require.NoError(t, os.WriteFile(out, []byte("hi!"), testFilePermission))
	}
	return dir
}

// appendToFile appends content to a file.
func appendToFile(t *testing.T, path string) {
	t.Helper()
	fh, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, testFilePermission)
	require.NoError(t, err, "opening file to append")
	_, err = fh.WriteString("\nbye!\n")
	require.NoError(t, err, "appending to file")
	require.NoError(t, fh.Close(), "closing file")
}

func TestNewestModTime(t *testing.T) {
	t.Parallel()
	dir := setupTestDir(t, []string{"a", "b", "c", "d"})

	time.Sleep(10 * time.Millisecond)
	outName := filepath.Join(dir, "c")
	appendToFile(t, outName)

	afi, err := os.Stat(filepath.Join(dir, "a"))
	require.NoError(t, err, "stating unmodified file")

	cfi, err := os.Stat(outName)
	require.NoError(t, err, "stating modified file")
	require.False(t, afi.ModTime().Equal(cfi.ModTime()), "modified and unmodified file mtimes equal")

	newest, err := NewestModTime(dir)
	require.NoError(t, err, "finding newest mod time")
	require.True(t, newest.Equal(cfi.ModTime()), "expected newest mod time to match c")
}

func TestOldestModTime(t *testing.T) {
	t.Parallel()
	dir := setupTestDir(t, []string{"a", "b", "c", "d"})

	time.Sleep(10 * time.Millisecond)
	for _, name := range []string{"a", "b", "d"} {
		appendToFile(t, filepath.Join(dir, name))
	}

	afi, err := os.Stat(filepath.Join(dir, "a"))
	require.NoError(t, err, "stating modified file")

	outName := filepath.Join(dir, "c")
	cfi, err := os.Stat(outName)
	require.NoError(t, err, "stating unmodified file")
	require.False(t, afi.ModTime().Equal(cfi.ModTime()), "modified and unmodified file mtimes equal")

	oldest, err := OldestModTime(dir)
	require.NoError(t, err, "finding oldest mod time")
	require.True(t, oldest.Equal(cfi.ModTime()), "expected oldest mod time to match c")
}
