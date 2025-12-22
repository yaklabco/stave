//go:build stave

package main

import "fmt"

var Aliases = map[string]any{
	"st":   Status,
	"stat": Status,
	"co":   Checkout,
}

// Prints status.
func Status() {
	fmt.Println("alias!")
}

func Checkout() {}
