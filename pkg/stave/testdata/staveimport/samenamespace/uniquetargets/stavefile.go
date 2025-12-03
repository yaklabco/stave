//go:build stave
// +build stave

package main

import (
	// stave:import samenamespace
	_ "github.com/yaklabco/stave/pkg/stave/testdata/staveimport/samenamespace/uniquetargets/package1"
	// stave:import samenamespace
	_ "github.com/yaklabco/stave/pkg/stave/testdata/staveimport/samenamespace/uniquetargets/package2"
)
