package dryrun

import (
	"flag"
	"fmt"
	"os"
	"testing"
)

var (
	printIsDryRunRequested bool
	printIsDryRunPossible  bool
	printIsDryRun          bool
)

func init() {
	flag.BoolVar(&printIsDryRunRequested, "printIsDryRunRequested", false, "")
	flag.BoolVar(&printIsDryRunPossible, "printIsDryRunPossible", false, "")
	flag.BoolVar(&printIsDryRun, "printIsDryRun", false, "")
}

func TestMain(m *testing.M) {
	flag.Parse()
	if printIsDryRunRequested {
		_, _ = fmt.Fprintln(os.Stdout, IsRequested())
		return
	}
	if printIsDryRunPossible {
		_, _ = fmt.Fprintln(os.Stdout, IsPossible())
		return
	}
	if printIsDryRun {
		_, _ = fmt.Fprintln(os.Stdout, IsDryRun())
		return
	}
	os.Exit(m.Run())
}
