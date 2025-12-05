//go:build stave

package main

import "github.com/yaklabco/stave/pkg/stave/testdata/mixed_lib_files/subdir"

func Build() {
	subdir.Build()
}
