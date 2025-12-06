//go:build stave

package main

import (
	"fmt"
	"os"
)

// HookTest writes a marker file to prove execution.
// Set HOOK_TEST_MARKER env var to the path where the marker should be written.
func HookTest() {
	marker := os.Getenv("HOOK_TEST_MARKER")
	if marker != "" {
		//#nosec G306 -- test file, permissions not critical
		if err := os.WriteFile(marker, []byte("executed"), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "failed to write marker: %v\n", err)
			os.Exit(1)
		}
	}
	fmt.Println("hook target executed")
}

// HookFail is a target that always fails with exit code 1.
func HookFail() error {
	return fmt.Errorf("intentional failure for testing")
}
