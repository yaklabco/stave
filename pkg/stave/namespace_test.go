package stave

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNamespaceFallback(t *testing.T) {
	t.Parallel()

	t.Run("NamespaceWithDefault", func(t *testing.T) {
		t.Parallel()
		dataDirForThisTest := filepath.Join(testDataDir, "namespace_default")
		mu := mutexByDir(dataDirForThisTest)
		mu.Lock()
		t.Cleanup(mu.Unlock)

		ctx := t.Context()

		stderr := &bytes.Buffer{}
		stdout := &bytes.Buffer{}

		runParams := RunParams{
			BaseCtx: ctx,
			Dir:     dataDirForThisTest,
			Stderr:  stderr,
			Stdout:  stdout,
			Args:    []string{"build"},
		}

		err := Run(runParams)
		require.NoError(t, err, "stderr was: %s", stderr.String())
		assert.Equal(t, "Build:Default\n", stdout.String())
	})

	t.Run("NamespaceWithoutDefault", func(t *testing.T) {
		t.Parallel()
		dataDirForThisTest := filepath.Join(testDataDir, "namespace_default")
		mu := mutexByDir(dataDirForThisTest)
		mu.Lock()
		t.Cleanup(mu.Unlock)

		ctx := t.Context()

		stderr := &bytes.Buffer{}
		stdout := &bytes.Buffer{}

		runParams := RunParams{
			BaseCtx: ctx,
			Dir:     dataDirForThisTest,
			Stderr:  stderr,
			Stdout:  stdout,
			Args:    []string{"deploy"},
		}

		err := Run(runParams)
		require.Error(t, err)
		assert.Contains(t, stderr.String(), "Target \"deploy\" is a namespace, but it has no Default target.")
	})

	t.Run("ImportedNamespace", func(t *testing.T) {
		t.Parallel()
		dataDirForThisTest := filepath.Join(testDataDir, "import_namespace")
		mu := mutexByDir(dataDirForThisTest)
		mu.Lock()
		t.Cleanup(mu.Unlock)

		ctx := t.Context()

		stderr := &bytes.Buffer{}
		stdout := &bytes.Buffer{}

		runParams := RunParams{
			BaseCtx: ctx,
			Dir:     dataDirForThisTest,
			Stderr:  stderr,
			Stdout:  stdout,
			Args:    []string{"lib:ns"},
		}

		err := Run(runParams)
		require.NoError(t, err, "stderr was: %s", stderr.String())
		assert.Equal(t, "lib:NS:Default\n", stdout.String())
	})
}
