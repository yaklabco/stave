//go:build stave
// +build stave

package main

// important things to note:
// * these two packages have the same package name, so they'll conflict
// when imported.
// * one is imported with underscore and one is imported normally.
//
// they should still work normally as staveimports

import (
	"fmt"

	//stave:import
	_ "github.com/yaklabco/stave/stave/testdata/staveimport/subdir1"
	//stave:import zz
	"github.com/yaklabco/stave/stave/testdata/staveimport/subdir2"
)

var Aliases = map[string]interface{}{
	"nsd2": staff.NS.Deploy2,
}

var Default = staff.NS.Deploy2

func Root() {
	fmt.Println("root")
}

