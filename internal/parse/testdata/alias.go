//go:build stave

package main

var Aliases = map[string]interface{}{
	"void": ReturnsVoid,
	"baz":  Build.Baz,
}
