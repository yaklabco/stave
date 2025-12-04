package stave

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArgs(t *testing.T) {
	ctx := t.Context()

	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}

	runParams := RunParams{
		BaseCtx: ctx,
		Dir:     "./testdata/args",
		Stderr:  stderr,
		Stdout:  stdout,
		Args:    []string{"status", "say", "hi", "bob", "count", "5", "status", "wait", "5ms", "cough", "false", "doubleIt", "3.1"},
	}

	err := Run(runParams)
	require.NoError(t, err, "stderr was: %s", stderr.String())
	expected := `status
saying hi bob
01234
status
waiting 5ms
not coughing
3.1 * 2 = 6.2
`

	assert.Equal(t, expected, stdout.String())
}

func TestBadIntArg(t *testing.T) {
	ctx := t.Context()

	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}

	runParams := RunParams{
		BaseCtx: ctx,
		Dir:     "./testdata/args",
		Stderr:  stderr,
		Stdout:  stdout,
		Args:    []string{"count", "abc123"},
	}

	err := Run(runParams)
	require.Error(t, err)

	expected := "can't convert argument \"abc123\" to int\n"
	assert.Equal(t, expected, stderr.String())
}

func TestBadBoolArg(t *testing.T) {
	ctx := t.Context()

	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}

	runParams := RunParams{
		BaseCtx: ctx,
		Dir:     "./testdata/args",
		Stderr:  stderr,
		Stdout:  stdout,
		Args:    []string{"cough", "abc123"},
	}

	err := Run(runParams)
	require.Error(t, err)

	expected := "can't convert argument \"abc123\" to bool\n"
	assert.Equal(t, expected, stderr.String())
}

func TestBadDurationArg(t *testing.T) {
	ctx := t.Context()

	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}

	runParams := RunParams{
		BaseCtx: ctx,
		Dir:     "./testdata/args",
		Stderr:  stderr,
		Stdout:  stdout,
		Args:    []string{"wait", "abc123"},
	}

	err := Run(runParams)
	require.Error(t, err)

	expected := "can't convert argument \"abc123\" to time.Duration\n"
	assert.Equal(t, expected, stderr.String())
}

func TestBadFloat64Arg(t *testing.T) {
	ctx := t.Context()

	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}

	runParams := RunParams{
		BaseCtx: ctx,
		Dir:     "./testdata/args",
		Stderr:  stderr,
		Stdout:  stdout,
		Args:    []string{"doubleIt", "abc123"},
	}

	err := Run(runParams)
	require.Error(t, err)

	expected := "can't convert argument \"abc123\" to float64\n"
	assert.Equal(t, expected, stderr.String())
}

func TestMissingArgs(t *testing.T) {
	ctx := t.Context()

	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}

	runParams := RunParams{
		BaseCtx: ctx,
		Dir:     "./testdata/args",
		Stderr:  stderr,
		Stdout:  stdout,
		Args:    []string{"say", "hi"},
	}

	err := Run(runParams)
	require.Error(t, err)

	expected := "not enough arguments for target \"Say\", expected 2, got 1\n"
	assert.Equal(t, expected, stderr.String())
}

func TestDocs(t *testing.T) {
	ctx := t.Context()

	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}

	runParams := RunParams{
		BaseCtx: ctx,
		Dir:     "./testdata/args",
		Stderr:  stderr,
		Stdout:  stdout,
		Info:    true,
		Args:    []string{"say"},
	}

	err := Run(runParams)
	require.NoError(t, err, "stderr was: %s", stderr.String())
	expected := `Say says something. It's pretty cool. I think you should try it.

Usage:

	stave say <msg> <name>

Aliases: speak

`

	assert.Equal(t, expected, stdout.String())
}

func TestMgF(t *testing.T) {
	ctx := t.Context()

	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}

	runParams := RunParams{
		BaseCtx: ctx,
		Dir:     "./testdata/args",
		Stderr:  stderr,
		Stdout:  stdout,
		Args:    []string{"HasDep"},
	}

	err := Run(runParams)
	require.NoError(t, err, "stderr was: %s", stderr.String())
	expected := "saying hi Susan\n"
	assert.Equal(t, expected, stdout.String())
}
