//go:build mage
// +build mage

package main

import (
	"fmt"

	//mage:import
	_ "github.com/yaklabco/staff/parse/testdata/importself"
)

func Build() {
	fmt.Println("built")
}
