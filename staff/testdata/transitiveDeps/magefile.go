//go:build mage
// +build mage

package main

import "github.com/yaklabco/staff/staff/testdata/transitiveDeps/dep"

func Run() {
	dep.Speak()
}
