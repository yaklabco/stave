package stave

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWatchFunctionality(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping watch test in short mode")
	}

	absRoot, err := filepath.Abs("../..")
	require.NoError(t, err)
	staveBin := buildStave(t, absRoot, "stave_test")

	tmpDir := t.TempDir()
	watchDir := filepath.Join(tmpDir, "watchme")
	require.NoError(t, os.Mkdir(watchDir, 0755))

	setupStaveProject(t, tmpDir, absRoot, fmt.Sprintf(`
func WatchTarget() error {
	watch.Watch("%s/**")
	fmt.Println("RUNNING_TARGET")
	return nil
}
`, watchDir))

	ctx := t.Context()
	handle := startWatch(t, ctx, staveBin, tmpDir, "watchtarget")
	defer handle.stop()

	handle.wait("RUNNING_TARGET")

	// 4. Modify a file in watchDir
	testFile := filepath.Join(watchDir, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("hello"), 0644))

	// 5. Check if it re-runs
	handle.wait("RUNNING_TARGET")
}

func TestWatchDepsFunctionality(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping watch test in short mode")
	}

	absRoot, err := filepath.Abs("../..")
	require.NoError(t, err)
	staveBin := buildStave(t, absRoot, "stave_test")

	tmpDir := t.TempDir()
	watchDir := filepath.Join(tmpDir, "watchme")
	require.NoError(t, os.Mkdir(watchDir, 0755))

	setupStaveProject(t, tmpDir, absRoot, fmt.Sprintf(`
func DepTarget() {
	fmt.Println("RUNNING_DEP")
}

func WatchTarget() error {
	watch.Watch("%s/**")
	watch.Deps(DepTarget)
	fmt.Println("RUNNING_TARGET")
	return nil
}
`, watchDir))

	ctx := t.Context()
	handle := startWatch(t, ctx, staveBin, tmpDir, "watchtarget")
	defer handle.stop()

	handle.wait("RUNNING_DEP")
	handle.wait("RUNNING_TARGET")

	// 4. Modify a file in watchDir
	testFile := filepath.Join(watchDir, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("hello"), 0644))

	// 5. Check if it re-runs
	handle.wait("RUNNING_DEP")
	handle.wait("RUNNING_TARGET")
}
