package stave

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompletion(t *testing.T) {
	ctx := t.Context()

	cwd, err := os.Getwd()
	require.NoError(t, err)

	rootCmd := NewRootCmd(ctx)

	// Use testdata/alias which has targets: Status (and st as alias)
	testDir := filepath.Join(cwd, "testdata", "alias")
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		testDir = filepath.Join(cwd, "cmd", "stave", "testdata", "alias")
	}

	// Manually set the flag since we are not calling Execute() which normally parses them
	err = rootCmd.PersistentFlags().Set("dir", testDir)
	require.NoError(t, err)

	targets, directive := rootCmd.ValidArgsFunction(rootCmd, []string{}, "")

	assert.Contains(t, targets, "status")
	assert.Contains(t, targets, "st")
	assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
}
