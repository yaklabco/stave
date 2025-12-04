//go:build ignore

package main

import (
	"context"
	"os"

	"github.com/yaklabco/stave/pkg/stave"
)

// This is a bootstrap builder, to build stave when you don't already *have* stave.
// Run it like
// go run bootstrap.go
// and it will install stave with all the right flags created for you.

func main() {
	os.Args = []string{os.Args[0], "-v", "install"}
	os.Exit(stave.Run(context.Background()))
}
