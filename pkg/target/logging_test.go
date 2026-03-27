package target

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLogging(t *testing.T) {
	// Setup a temporary directory for our test files
	dir := t.TempDir()

	dst := filepath.Join(dir, "dst.txt")
	src := filepath.Join(dir, "src.txt")

	// Create dst and src with known timestamps
	now := time.Now().Truncate(time.Second)
	require.NoError(t, os.WriteFile(dst, []byte("dst"), 0644))
	require.NoError(t, os.WriteFile(src, []byte("src"), 0644))

	// Set dst older than src
	require.NoError(t, os.Chtimes(dst, now.Add(-time.Hour), now.Add(-time.Hour)))
	require.NoError(t, os.Chtimes(src, now, now))

	// Capture slog output
	var buf bytes.Buffer
	h := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	l := slog.New(h)

	// We need to set the global logger or use a way to inject it if possible.
	// Since target package uses global slog, we replace it temporarily.
	old := slog.Default()
	slog.SetDefault(l)
	defer slog.SetDefault(old)

	t.Run("Path newer", func(t *testing.T) {
		buf.Reset()
		newer, err := Path(dst, src)
		require.NoError(t, err)
		require.True(t, newer)

		output := buf.String()
		require.Contains(t, output, "target.Path")
		require.Contains(t, output, "\"verdict\":true")
		require.Contains(t, output, "dst.txt")
		require.Contains(t, output, "src.txt")
	})

	t.Run("Path older", func(t *testing.T) {
		buf.Reset()
		// Swap them in logic: dst (newer) vs src (older)
		newer, err := Path(src, dst)
		require.NoError(t, err)
		require.False(t, newer)

		output := buf.String()
		require.Contains(t, output, "target.Path")
		require.Contains(t, output, "\"verdict\":false")
	})

	t.Run("Path missing dst", func(t *testing.T) {
		buf.Reset()
		missingDst := filepath.Join(dir, "missing.txt")
		newer, err := Path(missingDst, src)
		require.NoError(t, err)
		require.True(t, newer)

		output := buf.String()
		require.Contains(t, output, "target.Path")
		require.Contains(t, output, "destination does not exist")
	})

	t.Run("Dir", func(t *testing.T) {
		buf.Reset()
		newer, err := Dir(dst, src)
		require.NoError(t, err)
		require.True(t, newer)

		output := buf.String()
		require.Contains(t, output, "target.Dir")
		require.Contains(t, output, "\"verdict\":true")
	})

	t.Run("Glob", func(t *testing.T) {
		buf.Reset()
		newer, err := Glob(dst, filepath.Join(dir, "src*.txt"))
		require.NoError(t, err)
		require.True(t, newer)

		output := buf.String()
		require.Contains(t, output, "target.Glob")
		require.Contains(t, output, "\"verdict\":true")
	})
}
