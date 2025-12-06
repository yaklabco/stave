package stave

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/stave/pkg/st"
	"github.com/yaklabco/stave/pkg/stave"
)

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

func TestAlias(t *testing.T) {
	ctx := t.Context()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	runParams := stave.RunParams{
		BaseCtx: ctx,
		Dir:     "testdata/alias",
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
