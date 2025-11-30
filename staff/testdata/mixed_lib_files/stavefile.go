//go:build stave
// +build stave

package main

import "github.com/yaklabco/staff/staff/testdata/mixed_lib_files/subdir"

func Build() {
	subdir.Build()
}

