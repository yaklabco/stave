//go:build stave
// +build stave

// This is a comment on the package which should get turned into output with the
// list of targets.
package main

import (
	"fmt"
	"log/slog"

	"github.com/yaklabco/stave/pkg/st"
)

var Default = SomePig

// this should not be a target because it returns a string
func ReturnsString() string {
	fmt.Println("more stuff")
	return ""
}

func TestVerbose() {
	slog.Info("hi!")
}

// This is the synopsis for SomePig.  There's more data that won't show up.
func SomePig() {
	st.Deps(f)
}

func f() {}
