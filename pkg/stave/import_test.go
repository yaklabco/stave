//nolint:lll // Long string-literals.
package stave

import (
	"bytes"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testdataStaveimportDir = "./testdata/staveimport"

func TestStaveImportsList(t *testing.T) {
	t.Parallel()
	testDataDir := testdataStaveimportDir
	mu := mutexByDir(testDataDir)
	mu.Lock()
	defer mu.Unlock()

	ctx := t.Context()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	runParams := RunParams{
		BaseCtx: ctx,
		Dir:     testDataDir,
		Stdout:  stdout,
		Stderr:  stderr,
		List:    true,
	}

	err := Run(runParams)
	require.NoError(t, err, "stderr was: %s", stderr.String())
	expected := `
Targets:
  buildSubdir        Builds stuff.
  ns:deploy          deploys stuff.
  root               
  zz:buildSubdir2    Builds stuff.
  zz:ns:deploy2*     deploys stuff.

* default target
`[1:]

	assert.Equal(t, expected, stdout.String())
}

func TestStaveImportsHelp(t *testing.T) {
	t.Parallel()
	testDataDir := testdataStaveimportDir
	mu := mutexByDir(testDataDir)
	mu.Lock()
	defer mu.Unlock()

	ctx := t.Context()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	runParams := RunParams{
		BaseCtx: ctx,
		Dir:     testDataDir,
		Stdout:  stdout,
		Stderr:  stderr,
		Info:    true,
		Args:    []string{"buildSubdir"},
	}

	err := Run(runParams)
	require.NoError(t, err, "stderr was: %s", stderr.String())
	expected := `
BuildSubdir Builds stuff.

Usage:

	stave buildsubdir

`[1:]

	assert.Equal(t, expected, stdout.String())
}

func TestStaveImportsHelpNamed(t *testing.T) {
	t.Parallel()
	testDataDir := testdataStaveimportDir
	mu := mutexByDir(testDataDir)
	mu.Lock()
	defer mu.Unlock()

	ctx := t.Context()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	runParams := RunParams{
		BaseCtx: ctx,
		Dir:     testDataDir,
		Stdout:  stdout,
		Stderr:  stderr,
		Info:    true,
		Args:    []string{"zz:buildSubdir2"},
	}

	err := Run(runParams)
	require.NoError(t, err, "stderr was: %s", stderr.String())
	expected := `
BuildSubdir2 Builds stuff.

Usage:

	stave zz:buildsubdir2

`[1:]

	assert.Equal(t, expected, stdout.String())
}

func TestStaveImportsHelpNamedNS(t *testing.T) {
	t.Parallel()
	testDataDir := testdataStaveimportDir
	mu := mutexByDir(testDataDir)
	mu.Lock()
	defer mu.Unlock()

	ctx := t.Context()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	runParams := RunParams{
		BaseCtx: ctx,
		Dir:     testDataDir,
		Stdout:  stdout,
		Stderr:  stderr,
		Info:    true,
		Args:    []string{"zz:ns:deploy2"},
	}

	err := Run(runParams)
	require.NoError(t, err, "stderr was: %s", stderr.String())
	expected := `
Deploy2 deploys stuff.

Usage:

	stave zz:ns:deploy2

Aliases: nsd2

`[1:]

	assert.Equal(t, expected, stdout.String())
}

func TestStaveImportsRoot(t *testing.T) {
	t.Parallel()
	testDataDir := testdataStaveimportDir
	mu := mutexByDir(testDataDir)
	mu.Lock()
	defer mu.Unlock()

	ctx := t.Context()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	runParams := RunParams{
		BaseCtx: ctx,
		Dir:     testDataDir,
		Stdout:  stdout,
		Stderr:  stderr,
		Args:    []string{"root"},
	}

	err := Run(runParams)
	require.NoError(t, err, "stderr was: %s", stderr.String())
	expected := "root\n"
	assert.Equal(t, expected, stdout.String())
}

func TestStaveImportsNamedNS(t *testing.T) {
	t.Parallel()
	testDataDir := testdataStaveimportDir
	mu := mutexByDir(testDataDir)
	mu.Lock()
	defer mu.Unlock()

	ctx := t.Context()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	runParams := RunParams{
		BaseCtx: ctx,
		Dir:     testDataDir,
		Stdout:  stdout,
		Stderr:  stderr,
		Args:    []string{"zz:nS:deploy2"},
	}

	err := Run(runParams)
	require.NoError(t, err, "stderr was: %s", stderr.String())
	expected := "deploy2\n"
	assert.Equal(t, expected, stdout.String())
}

func TestStaveImportsNamedRoot(t *testing.T) {
	t.Parallel()
	testDataDir := testdataStaveimportDir
	mu := mutexByDir(testDataDir)
	mu.Lock()
	defer mu.Unlock()

	ctx := t.Context()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	runParams := RunParams{
		BaseCtx: ctx,
		Dir:     testDataDir,
		Stdout:  stdout,
		Stderr:  stderr,
		Args:    []string{"zz:buildSubdir2"},
	}

	err := Run(runParams)
	require.NoError(t, err, "stderr was: %s", stderr.String())
	expected := "buildsubdir2\n"
	assert.Equal(t, expected, stdout.String())
	assert.Empty(t, stderr.String())
}

func TestStaveImportsRootImportNS(t *testing.T) {
	t.Parallel()
	testDataDir := testdataStaveimportDir
	mu := mutexByDir(testDataDir)
	mu.Lock()
	defer mu.Unlock()

	ctx := t.Context()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	runParams := RunParams{
		BaseCtx: ctx,
		Dir:     testDataDir,
		Stdout:  stdout,
		Stderr:  stderr,
		Args:    []string{"nS:deploy"},
	}

	err := Run(runParams)
	require.NoError(t, err, "stderr was: %s", stderr.String())
	expected := "deploy\n"
	assert.Equal(t, expected, stdout.String())
}

func TestStaveImportsRootImport(t *testing.T) {
	t.Parallel()
	testDataDir := testdataStaveimportDir
	mu := mutexByDir(testDataDir)
	mu.Lock()
	defer mu.Unlock()

	ctx := t.Context()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	runParams := RunParams{
		BaseCtx: ctx,
		Dir:     testDataDir,
		Stdout:  stdout,
		Stderr:  stderr,
		Args:    []string{"buildSubdir"},
	}

	err := Run(runParams)
	require.NoError(t, err, "stderr was: %s", stderr.String())
	expected := "buildsubdir\n"
	assert.Equal(t, expected, stdout.String())
}

func TestStaveImportsAliasToNS(t *testing.T) {
	t.Parallel()
	testDataDir := testdataStaveimportDir
	mu := mutexByDir(testDataDir)
	mu.Lock()
	defer mu.Unlock()

	ctx := t.Context()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	runParams := RunParams{
		BaseCtx: ctx,
		Dir:     testDataDir,
		Stdout:  stdout,
		Stderr:  stderr,
		Args:    []string{"nsd2"},
	}

	err := Run(runParams)
	require.NoError(t, err, "stderr was: %s", stderr.String())
	expected := "deploy2\n"
	assert.Equal(t, expected, stdout.String())
}

func TestStaveImportsOneLine(t *testing.T) {
	t.Parallel()
	testDataDir := "./testdata/staveimport/oneline"
	mu := mutexByDir(testDataDir)
	mu.Lock()
	defer mu.Unlock()

	ctx := t.Context()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	runParams := RunParams{
		BaseCtx: ctx,
		Dir:     testDataDir,
		Stdout:  stdout,
		Stderr:  stderr,
		Args:    []string{"build"},
	}

	err := Run(runParams)
	require.NoError(t, err, "stderr was: %s", stderr.String())
	expected := "build\n"
	assert.Equal(t, expected, stdout.String())
}

func TestStaveImportsTrailing(t *testing.T) {
	t.Parallel()
	testDataDir := "./testdata/staveimport/trailing"
	mu := mutexByDir(testDataDir)
	mu.Lock()
	defer mu.Unlock()

	ctx := t.Context()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	runParams := RunParams{
		BaseCtx: ctx,
		Dir:     testDataDir,
		Stdout:  stdout,
		Stderr:  stderr,
		Args:    []string{"build"},
	}

	err := Run(runParams)
	require.NoError(t, err, "stderr was: %s", stderr.String())
	expected := "build\n"
	assert.Equal(t, expected, stdout.String())
}

func TestStaveImportsTaggedPackage(t *testing.T) {
	t.Parallel()
	testDataDir := "./testdata/staveimport/tagged"
	mu := mutexByDir(testDataDir)
	mu.Lock()
	defer mu.Unlock()

	ctx := t.Context()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	runParams := RunParams{
		BaseCtx: ctx,
		Dir:     testDataDir,
		Stdout:  stdout,
		Stderr:  stderr,
		List:    true,
	}

	err := Run(runParams)
	require.Error(t, err)

	actual := err.Error()
	// Match a shorter version of the error message, since the output from go list differs between versions
	expected := `
parsing stavefiles: error running "go list -f {{.Dir}}||{{.Name}} github.com/yaklabco/stave/pkg/stave/testdata/staveimport/tagged/pkg": exit status 1`[1:]
	actualShortened := lo.Substring(actual, 0, uint(len(expected)))

	assert.Contains(t, expected, actualShortened)
}

func TestStaveImportsSameNamespaceUniqueTargets(t *testing.T) {
	t.Parallel()
	testDataDir := "./testdata/staveimport/samenamespace/uniquetargets"
	mu := mutexByDir(testDataDir)
	mu.Lock()
	defer mu.Unlock()

	ctx := t.Context()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	runParams := RunParams{
		BaseCtx: ctx,
		Dir:     testDataDir,
		Stdout:  stdout,
		Stderr:  stderr,
		List:    true,
	}

	err := Run(runParams)
	require.NoError(t, err, "stderr was: %s", stderr.String())
	expected := `
Targets:
  samenamespace:build1    
  samenamespace:build2    
`[1:]

	assert.Equal(t, expected, stdout.String())
}

func TestStaveImportsSameNamespaceDupTargets(t *testing.T) {
	t.Parallel()
	testDataDir := "./testdata/staveimport/samenamespace/duptargets"
	mu := mutexByDir(testDataDir)
	mu.Lock()
	defer mu.Unlock()

	ctx := t.Context()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	runParams := RunParams{
		BaseCtx: ctx,
		Dir:     testDataDir,
		Stdout:  stdout,
		Stderr:  stderr,
		List:    true,
	}

	err := Run(runParams)
	require.Error(t, err)

	expected := `
parsing stavefiles: "samenamespace:build" target has multiple definitions: github.com/yaklabco/stave/pkg/stave/testdata/staveimport/samenamespace/duptargets/package1.Build, github.com/yaklabco/stave/pkg/stave/testdata/staveimport/samenamespace/duptargets/package2.Build
`[1:]

	assert.Equal(t, expected, err.Error())
}
