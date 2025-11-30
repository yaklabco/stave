//go:build stave
// +build stave

package main

import (
	// stave:import samenamespace
	_ "github.com/yaklabco/staff/staff/testdata/staveimport/samenamespace/uniquetargets/package1"
	// stave:import samenamespace
	_ "github.com/yaklabco/staff/staff/testdata/staveimport/samenamespace/uniquetargets/package2"
)

