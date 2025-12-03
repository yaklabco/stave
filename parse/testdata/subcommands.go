//go:build stave
// +build stave

package main

import "github.com/yaklabco/stave/pkg/st"

type Build st.Namespace

func (Build) Foobar() error {
	// do your foobar build
	return nil
}

func (Build) Baz() {
	// do your baz build
}

type Init st.Namespace

func (Init) Foobar() error {
	// do your foobar defined in init namespace
	return nil
}
