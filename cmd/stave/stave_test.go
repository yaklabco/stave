package stave

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/stave/pkg/st"
	"github.com/yaklabco/stave/pkg/stave"
)

// TestMain pins the stave cache and user config to temp dirs so tests that
// compile stavefiles never write into (or clean out) the developer's real
// cache, and never load the developer's real config.
func TestMain(m *testing.M) {
	os.Exit(runMain(m))
}

func runMain(m *testing.M) int {
	dir, err := os.MkdirTemp("", "stave-cmd-test")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer func() {
		if removeErr := os.RemoveAll(dir); removeErr != nil {
			fmt.Fprintln(os.Stderr, "error removing temp dir:", removeErr)
		}
	}()

	if err := os.Setenv(st.CacheEnv, filepath.Join(dir, "cache")); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if err := os.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "config")); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	return m.Run()
}

func TestVerboseEnv(t *testing.T) {
	ctx := t.Context()
	t.Setenv("STAVEFILE_VERBOSE", "true")
	runFunc := func(params stave.RunParams) error {
		assert.True(t, params.Verbose)
		return nil
	}
	rootCmd := NewRootCmd(ctx, withRunFunc(runFunc))
	require.NoError(t, ExecuteWithFang(ctx, rootCmd))
}

func TestVerboseFalseEnv(t *testing.T) {
	ctx := t.Context()
	t.Setenv("STAVEFILE_VERBOSE", "0")
	runFunc := func(params stave.RunParams) error {
		assert.False(t, params.Verbose)
		return nil
	}
	rootCmd := NewRootCmd(ctx, withRunFunc(runFunc))
	require.NoError(t, ExecuteWithFang(ctx, rootCmd))
}

func TestMultilineEnv(t *testing.T) {
	ctx := t.Context()
	t.Setenv("STAVEFILE_MULTILINE", "true")
	runFunc := func(params stave.RunParams) error {
		assert.True(t, params.Multiline)
		return nil
	}
	rootCmd := NewRootCmd(ctx, withRunFunc(runFunc))
	require.NoError(t, ExecuteWithFang(ctx, rootCmd))
}

func TestMultilineFalseEnv(t *testing.T) {
	ctx := t.Context()
	t.Setenv("STAVEFILE_MULTILINE", "0")
	runFunc := func(params stave.RunParams) error {
		assert.False(t, params.Multiline)
		return nil
	}
	rootCmd := NewRootCmd(ctx, withRunFunc(runFunc))
	require.NoError(t, ExecuteWithFang(ctx, rootCmd))
}

func TestParse(t *testing.T) {
	ctx := t.Context()
	runFunc := func(params stave.RunParams) error {
		assert.False(t, params.Init)
		assert.True(t, params.Debug)
		assert.Equal(t, "dir", params.Dir)
		assert.Equal(t, "foo", params.GoCmd)
		assert.Equal(t, []string{"build", "deploy"}, params.Args)

		return nil
	}
	rootCmd := NewRootCmd(ctx, withRunFunc(runFunc))
	rootCmd.SetArgs([]string{"-v", "--debug", "--gocmd=foo", "-C", "dir", "build", "deploy"})
	require.NoError(t, ExecuteWithFang(ctx, rootCmd))
}

func TestClean(t *testing.T) {
	ctx := t.Context()

	// Pin this test's cache to its own temp dir, isolated from the cache
	// artifacts other tests create under TestMain's shared dir, since this
	// test wipes the dir and asserts on its contents. The env var also keeps
	// st.CacheDir and the config-resolved cache dir pointing at one place.
	t.Setenv(st.CacheEnv, t.TempDir())

	require.NoError(t, os.RemoveAll(st.CacheDir()))

	rootCmd := NewRootCmd(ctx)
	rootCmd.SetArgs([]string{"--clean"})
	require.NoError(t, ExecuteWithFang(ctx, rootCmd))

	TestAlias(t) // make sure we've got something in the CACHE_DIR
	files, err := os.ReadDir(st.CacheDir())
	require.NoError(t, err)
	assert.NotEmpty(t, files)

	runFunc := func(params stave.RunParams) error {
		assert.True(t, params.Clean)
		return nil
	}
	rootCmd = NewRootCmd(ctx, withRunFunc(runFunc))
	rootCmd.SetArgs([]string{"--clean"})
	require.NoError(t, ExecuteWithFang(ctx, rootCmd))

	rootCmd = NewRootCmd(ctx)
	rootCmd.SetArgs([]string{"--clean"})
	require.NoError(t, ExecuteWithFang(ctx, rootCmd))

	infos, err := os.ReadDir(st.CacheDir())
	require.NoError(t, err)

	var names []string
	for _, i := range infos {
		if !i.IsDir() {
			names = append(names, i.Name())
		}
	}

	assert.Empty(t, names)
}

const testDataDir = "testdata"

func TestAlias(t *testing.T) {
	ctx := t.Context()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	runParams := stave.RunParams{
		BaseCtx: ctx,
		Dir:     filepath.Join(testDataDir, "alias"),
		Stdout:  stdout,
		Stderr:  stderr,
		Args:    []string{"status"},
		Debug:   true,
	}

	err := stave.Run(runParams)
	require.NoError(t, err, "stderr was: %s", stderr.String())

	expected := "alias!\n"
	assert.Equal(t, expected, stdout.String())

	stdout.Reset()
	stderr.Reset()
	runParams.Args = []string{"st"}
	err = stave.Run(runParams)
	require.NoError(t, err)

	assert.Equal(t, expected, stdout.String())
}

func TestHooksFlag(t *testing.T) {
	ctx := t.Context()
	runFunc := func(params stave.RunParams) error {
		assert.True(t, params.Hooks)
		assert.Equal(t, []string{"install"}, params.Args)

		return nil
	}
	rootCmd := NewRootCmd(ctx, withRunFunc(runFunc))
	rootCmd.SetArgs([]string{"--hooks", "install"})
	require.NoError(t, ExecuteWithFang(ctx, rootCmd))
}

func TestHooksFlagWithVerbose(t *testing.T) {
	ctx := t.Context()
	runFunc := func(params stave.RunParams) error {
		assert.True(t, params.Hooks)
		assert.True(t, params.Verbose)
		assert.Equal(t, []string{"list"}, params.Args)

		return nil
	}
	rootCmd := NewRootCmd(ctx, withRunFunc(runFunc))
	rootCmd.SetArgs([]string{"--verbose", "--hooks", "list"})
	require.NoError(t, ExecuteWithFang(ctx, rootCmd))
}

func TestConfigFlag(t *testing.T) {
	ctx := t.Context()
	runFunc := func(params stave.RunParams) error {
		assert.True(t, params.Config)
		assert.Equal(t, []string{"show"}, params.Args)

		return nil
	}
	rootCmd := NewRootCmd(ctx, withRunFunc(runFunc))
	rootCmd.SetArgs([]string{"--config", "show"})
	require.NoError(t, ExecuteWithFang(ctx, rootCmd))
}

func TestConfigFlagNoSubcommand(t *testing.T) {
	ctx := t.Context()
	runFunc := func(params stave.RunParams) error {
		assert.True(t, params.Config)
		assert.Empty(t, params.Args)

		return nil
	}
	rootCmd := NewRootCmd(ctx, withRunFunc(runFunc))
	rootCmd.SetArgs([]string{"--config"})
	require.NoError(t, ExecuteWithFang(ctx, rootCmd))
}
