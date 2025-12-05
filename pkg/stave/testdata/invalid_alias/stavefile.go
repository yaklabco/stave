//go:build stave

package main

import "fmt"

var Aliases = map[string]interface{}{
	"co": checkout,
}

func checkout() {
	fmt.Println("done!")
}
