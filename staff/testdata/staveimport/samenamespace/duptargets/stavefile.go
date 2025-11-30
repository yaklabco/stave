//go:build stave
// +build stave

package sametarget

import (
	// stave:import samenamespace
	_ "github.com/yaklabco/staff/staff/testdata/staveimport/samenamespace/duptargets/package1"
	// stave:import samenamespace
	_ "github.com/yaklabco/staff/staff/testdata/staveimport/samenamespace/duptargets/package2"
)

