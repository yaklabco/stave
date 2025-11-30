//go:build stave
// +build stave

package main

import (
	"fmt"

	//stave:import
	_ "github.com/yaklabco/staff/parse/testdata/importself"
)

func Build() {
	fmt.Println("built")
}

