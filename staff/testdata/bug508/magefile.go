//go:build mage
// +build mage

package main

import (
	//mage:import
	"github.com/yaklabco/staff/staff/testdata/bug508/deps"
)

var Default = deps.Test
