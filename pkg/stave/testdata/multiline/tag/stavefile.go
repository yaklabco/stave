//go:build stave

// This is a global comment for the stave output.
// It should retain line returns.
package main

//stave:multiline // Enable multiline description output.

import "github.com/yaklabco/stave/pkg/st"

// DoIt is a dummy function with a multiline comment.
// That should show up with multiple lines.
func DoIt() {
}

// Sub is a namespace.
// It also has a line return.
type Sub st.Namespace

// DoItToo is a dummy function with a multiline comment.
// Here's the second line.
func (Sub) DoItToo() {
}
