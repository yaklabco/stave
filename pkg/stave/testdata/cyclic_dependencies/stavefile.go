//go:build stave
// +build stave

package main

import (
	"fmt"

	"github.com/yaklabco/stave/pkg/st"
)

func Step1() {
	st.Deps(Step2)
	fmt.Println("Step1")
}

func Step2() {
	st.Deps(Step3)
	fmt.Println("Step2")
}

func Step3() {
	st.Deps(Step1)
	fmt.Println("Step3")
}
