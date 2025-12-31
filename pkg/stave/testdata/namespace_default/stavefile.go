//go:build stave

package main

import (
	"fmt"

	"github.com/yaklabco/stave/pkg/st"
)

type Build st.Namespace

func (Build) Default() {
	fmt.Println("Build:Default")
}

func (Build) Test() {
	fmt.Println("Build:Test")
}

type Deploy st.Namespace

func (Deploy) Production() {
	fmt.Println("Deploy:Production")
}
