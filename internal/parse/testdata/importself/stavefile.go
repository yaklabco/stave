//go:build stave

package main

import (
	"fmt"

	//stave:import
	_ "github.com/yaklabco/stave/internal/parse/testdata/importself"
)

func Build() {
	fmt.Println("built")
}
