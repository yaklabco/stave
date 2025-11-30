//go:build stave
// +build stave

package main

import (
	"github.com/yaklabco/staff/mg"
)

var Default = FooBar

func WrongSignature(c complex128) {
}

func FooBar() {
	mg.Deps(mg.F(WrongSignature, 0))
}

