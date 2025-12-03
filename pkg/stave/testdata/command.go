//go:build stave
// +build stave

package main

import (
	"fmt"
	"log"

	"github.com/yaklabco/stave/pkg/st"
)

// This should work as a default - even if it's in a different file
var Default = ReturnsNilError

// this should not be a target because it returns a string
func ReturnsString() string {
	fmt.Println("more stuff")
	return ""
}

func TestVerbose() {
	log.Println("hi!")
}

func ReturnsVoid() {
	st.Deps(f)
}

func f() {}
