//go:build stave

// Compiled package description.
package main

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/yaklabco/stave/pkg/st"
)

var Default = Deploy

// This is very verbose.
func TestVerbose() {
	slog.Info("hi!")
}

// PrintVerboseFlag prints the value of st.Verbose() to stdout.
func PrintVerboseFlag() {
	fmt.Printf("st.Verbose()==%v", st.Verbose())
}

// This is the synopsis for Deploy. This part shouldn't show up.
func Deploy() {
	st.Deps(f)
}

// Sleep sleeps 5 seconds.
func Sleep() {
	time.Sleep(5 * time.Second)
}

func f() {
	slog.Info("i am independent -- not")
}
