package stave

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testDataCyclicDependenciesDir = filepath.Join(testDataDir, "cyclic_dependencies")
)

// TestCyclicDependencyDetection verifies proper detection of cyclic dependencies.
func TestCyclicDependencyDetection(t *testing.T) {
	t.Parallel()
	dataDirForThisTest := testDataCyclicDependenciesDir
	mu := mutexByDir(dataDirForThisTest)
	mu.Lock()
	t.Cleanup(mu.Unlock)

	ctx := t.Context()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	err := Run(RunParams{
		BaseCtx: ctx,
		Dir:     dataDirForThisTest,
		Stdout:  stdout,
		Stderr:  stderr,
		Args:    []string{"Step1"},
	})
	require.Error(t, err)

	expected := "circular dependency detected"
	assert.Contains(t, stderr.String(), expected)
}
