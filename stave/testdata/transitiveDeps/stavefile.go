//go:build stave
// +build stave

package main

import "github.com/yaklabco/stave/stave/testdata/transitiveDeps/dep"

func Run() {
	dep.Speak()
}

