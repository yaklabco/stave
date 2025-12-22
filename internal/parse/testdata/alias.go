//go:build stave

package main

var Aliases = map[string]any{
	"void": ReturnsVoid,
	"baz":  Build.Baz,
}
