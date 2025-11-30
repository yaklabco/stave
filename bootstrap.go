//go:build ignore
// +build ignore

package main

import (
	"os"

	"github.com/yaklabco/staff/staff"
)

// This is a bootstrap builder, to build staff when you don't already *have* staff.
// Run it like
// go run bootstrap.go
// and it will install staff with all the right flags created for you.

func main() {
	os.Args = []string{os.Args[0], "-v", "install"}
	os.Exit(staff.Main())
}
