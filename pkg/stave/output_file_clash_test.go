package stave

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConcurrentRuns(t *testing.T) {
	t.Parallel()

	// We want to run two stave runs in the same directory at the same time.
	// We'll use a temporary directory to avoid messing with the project.

	tmpDir, err := os.MkdirTemp("", "stave-concurrent-test")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Copy a simple stavefile to the tmpDir
	stavefileContent := `//go:build stave

package main

import (
	"fmt"
	"time"
)

func Default(runID int) error {
	fmt.Println("hello")
	time.Sleep(time.Duration(0.5+float64(runID))*time.Second)
	return nil
}
`
	err = os.WriteFile(filepath.Join(tmpDir, "stavefile.go"), []byte(stavefileContent), 0644)
	require.NoError(t, err)

	// We also need a go.mod in that directory so 'go build' works
	goModContent := `module testconcurrent
go 1.24
`
	err = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644)
	require.NoError(t, err)

	ctx := t.Context()

	const numRuns = 4
	var wg sync.WaitGroup
	wg.Add(numRuns)

	runStave := func(id int) {
		defer wg.Done()

		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		params := RunParams{
			BaseCtx: ctx,
			Dir:     tmpDir,
			Stdout:  stdout,
			Stderr:  stderr,
			Args:    []string{"Default", strconv.Itoa(id)},
			Force:   true, // Force recompile to ensure GenerateMainFile is called
		}

		err := Run(params)
		if err != nil {
			t.Errorf("[DEBUG_LOG] Run %d failed: %v\nStderr: %s\n", id, err, stderr.String())
		}
	}

	for id := range numRuns {
		go runStave(id)
	}

	wg.Wait()
}
