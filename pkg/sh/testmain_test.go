package sh

import (
	"flag"
	"fmt"
	"os"
	"testing"
)

var (
	helperCmd    bool
	printArgs    bool
	stderr       string
	stdout       string
	exitCode     int
	printVar     string
	dryRunOutput bool
)

func init() {
	flag.BoolVar(&helperCmd, "helper", false, "")
	flag.BoolVar(&printArgs, "printArgs", false, "")
	flag.StringVar(&stderr, "stderr", "", "")
	flag.StringVar(&stdout, "stdout", "", "")
	flag.IntVar(&exitCode, "exit", 0, "")
	flag.StringVar(&printVar, "printVar", "", "")
	flag.BoolVar(&dryRunOutput, "dryRunOutput", false, "")
}

func TestMain(m *testing.M) {
	flag.Parse()

	if printArgs {
		_, _ = fmt.Fprintln(os.Stdout, flag.Args())
		return
	}
	if printVar != "" {
		_, _ = fmt.Fprintln(os.Stdout, os.Getenv(printVar))
		return
	}

	if dryRunOutput {
		// Simulate dry-run mode and print the output of a command that would have been run.
		// We use a non-echo command to make the "DRYRUN: " prefix deterministic.
		_ = os.Setenv("STAVEFILE_DRYRUN_POSSIBLE", "1")
		_ = os.Setenv("STAVEFILE_DRYRUN", "1")
		s, err := Output("somecmd", "arg1", "arg two")
		if err != nil {
			_, _ = fmt.Fprintln(os.Stdout, "ERR:", err)
			return
		}
		_, _ = fmt.Fprintln(os.Stdout, s)
		return
	}

	if helperCmd {
		_, _ = fmt.Fprintln(os.Stderr, stderr)
		_, _ = fmt.Fprintln(os.Stdout, stdout)
		os.Exit(exitCode)
	}
	os.Exit(m.Run())
}
