//go:build stave
// +build stave

package main

import "github.com/yaklabco/staff/staff/testdata/transitiveDeps/dep"

func Run() {
	dep.Speak()
}

