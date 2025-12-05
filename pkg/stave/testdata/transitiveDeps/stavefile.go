//go:build stave

package main

import "github.com/yaklabco/stave/pkg/stave/testdata/transitiveDeps/dep"

func Run() {
	dep.Speak()
}
