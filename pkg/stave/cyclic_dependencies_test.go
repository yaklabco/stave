package stave

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCyclicDependencyDetection verifies proper detection of cyclic dependencies.
func TestCyclicDependencyDetection(t *testing.T) {
	t.Parallel()
	testDataDir := "./testdata/cyclic_dependencies"
	mu := mutexByDir(testDataDir)
	mu.Lock()
	defer mu.Unlock()

	ctx := t.Context()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	err := Run(RunParams{
		BaseCtx: ctx,
		Dir:     testDataDir,
		Stdout:  stdout,
		Stderr:  stderr,
		Args:    []string{"Step1"},
	})
	require.Error(t, err)

	expected := "circular dependency detected"
	assert.Contains(t, stderr.String(), expected)
}
