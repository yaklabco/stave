//go:build stave

package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/yaklabco/stave/pkg/st" // st contains helpful utility functions, like Deps
)

// Default target to run when none is specified
// If not set, running stave will list available targets
// var Default = Build

// A build step that requires additional params, or platform specific steps for example
func Build() error {
	st.Deps(InstallDeps)
	fmt.Println("Building...")
	cmd := exec.Command("go", "build", "-o", "MyApp", ".")
	return cmd.Run()
}

// A custom install step if you need your bin someplace other than go/bin
func Install() error {
	st.Deps(Build)
	fmt.Println("Installing...")
	return os.Rename("./MyApp", "/usr/bin/MyApp")
}

// Manage your deps, or running package managers.
func InstallDeps() error {
	fmt.Println("Installing Deps...")
	cmd := exec.Command("go", "get", "github.com/stretchr/piglatin")
	return cmd.Run()
}

// Clean up after yourself
func Clean() {
	fmt.Println("Cleaning...")
	os.RemoveAll("MyApp")
}
