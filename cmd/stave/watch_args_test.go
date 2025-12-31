package stave

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestWatchWithArgs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping watch test in short mode")
	}

	absRoot, err := filepath.Abs("../..")
	require.NoError(t, err)
	staveBin := buildStave(t, absRoot, "stave_watch_args_test")

	tmpDir := t.TempDir()
	watchDir := filepath.Join(tmpDir, "watchme")
	require.NoError(t, os.Mkdir(watchDir, 0755))

	setupStaveProject(t, tmpDir, absRoot, `
func WatchDir(dir string) {
	watch.Watch(fmt.Sprintf("%s/**", dir))
	fmt.Printf("WATCHING_DIR: %s\n", dir)
}
`)

	ctx := t.Context()
	handle := startWatch(t, ctx, staveBin, tmpDir, "-v", "watchdir", watchDir)
	defer handle.stop()

	handle.wait("WATCHING_DIR")

	// Trigger a file change
	time.Sleep(1 * time.Second) // Give it a moment to start watching
	require.NoError(t, os.WriteFile(filepath.Join(watchDir, "test.txt"), []byte("hi"), 0644))

	// Wait for re-run
	handle.wait("WATCHING_DIR")
}
