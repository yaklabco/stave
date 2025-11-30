//go:build mage
// +build mage

package sametarget

import (
	// mage:import samenamespace
	_ "github.com/yaklabco/staff/staff/testdata/mageimport/samenamespace/duptargets/package1"
	// mage:import samenamespace
	_ "github.com/yaklabco/staff/staff/testdata/mageimport/samenamespace/duptargets/package2"
)
