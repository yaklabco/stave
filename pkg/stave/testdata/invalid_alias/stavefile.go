//go:build stave

package main

import "fmt"

var Aliases = map[string]any{
	"co": checkout,
}

func checkout() {
	fmt.Println("done!")
}
