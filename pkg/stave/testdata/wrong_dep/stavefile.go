//go:build stave
// +build stave

package main

import (
	"github.com/yaklabco/stave/pkg/st"
)

var Default = FooBar

func WrongSignature(c complex128) {
}

func FooBar() {
	st.Deps(st.F(WrongSignature, 0))
}
