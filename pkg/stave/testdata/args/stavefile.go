//go:build stave
// +build stave

package main

import (
	"context"
	"fmt"
	"time"

	"github.com/yaklabco/stave/pkg/st"
)

var Aliases = map[string]interface{}{
	"speak": Say,
}

// Prints status.
func Status() {
	fmt.Println("status")
}

// Say says something. It's pretty cool.
// I think you should try it.
func Say(ctx context.Context, msg, name string) {
	fmt.Println("saying", msg, name)
}

func Count(i int) error {
	for x := 0; x < i; x++ {
		fmt.Print(x)
	}
	fmt.Println()
	return nil
}

func Wait(d time.Duration) {
	fmt.Println("waiting", d)
}

func Cough(ctx context.Context, b bool) error {
	if b {
		fmt.Println("coughing")
	} else {
		fmt.Println("not coughing")
	}
	return nil
}

func HasDep() {
	st.Deps(st.F(Say, "hi", "Susan"))
}

func DoubleIt(f float64) {
	fmt.Printf("%.1f * 2 = %.1f\n", f, f*2)
}
