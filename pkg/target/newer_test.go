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

func TestDirNewerWithEmptyDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// No files inside dir — nothing newer than any target time
	newer, err := DirNewer(time.Now(), dir)
	if err != nil {
		t.Fatal(err)
	}
	// The directory itself was just created, its modtime is very recent
	// so it should be newer than time.Time{} but we pass time.Now()
	// The dir's modtime should be ≤ now, so it should not be newer
	if newer {
		t.Fatal("expected empty dir to not be newer than now")
	}
}

func TestDirNewerMissingSource(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	_, err := DirNewer(time.Now(), filepath.Join(dir, "nonexistent"))
	if err == nil {
		t.Fatal("expected error for missing source")
	}
}

func TestNewestModTimeEmptyDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// An empty directory still has the directory entry itself
	newest, err := NewestModTime(dir)
	if err != nil {
		t.Fatal(err)
	}
	// Since directory's own timestamp is ignored, this should return time.Time{}
	info, err := os.Stat(dir)
	require.NoError(t, err)
	if !newest.Equal(time.Time{}) {
		t.Fatalf("expected dir modtime, got %v vs %v", newest, info.ModTime())
	}
}

func TestOldestModTimeEmptyDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	oldest, err := OldestModTime(dir)
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(dir)
	require.NoError(t, err)
	if !oldest.Equal(farFutureTime) {
		t.Fatalf("expected dir modtime, got %v vs %v", oldest, info.ModTime())
	}
}

func TestPathNewerMissingSource(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	_, err := PathNewer(time.Now(), filepath.Join(dir, "nonexistent"))
	if err == nil {
		t.Fatal("expected error for missing source")
	}
}

func TestGlobNewerNoMatch(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	_, err := GlobNewer(time.Now(), filepath.Join(dir, "*.nonexistent"))
	if err == nil {
		t.Fatal("expected error for glob with no matches")
	}
}
