//go:build stave

package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/charmbracelet/log"
	"github.com/yaklabco/stave/pkg/st" // st contains helpful utility functions, like Deps
	"github.com/yaklabco/stave/pkg/stave/prettylog"
)

func init() {
	logHandler := prettylog.SetupPrettyLogger(os.Stdout)
	if st.Debug() {
		logHandler.SetLevel(log.DebugLevel)
	}
}

// Default target to run when none is specified
// If not set, running stave will list available targets
// var Default = Build

// Build compiles the project and creates an executable named "MyApp"
func Build() error {
	st.Deps(InstallDeps)
	_, _ = fmt.Fprintln(os.Stdout, "Building...")
	cmd := exec.Command("go", "build", "-o", "MyApp", ".")
	return cmd.Run()
}

// Install installs the built application to the target system directory
func Install() error {
	st.Deps(Build)
	_, _ = fmt.Fprintln(os.Stdout, "Installing...")
	return os.Rename("./MyApp", "/usr/bin/MyApp")
}

// InstallDeps installs the necessary dependencies for the project
func InstallDeps() error {
	_, _ = fmt.Fprintln(os.Stdout, "Installing Deps...")
	cmd := exec.Command("go", "get", "github.com/stretchr/piglatin")
	return cmd.Run()
}

// Clean up after yourself.
func Clean() error {
	_, _ = fmt.Fprintln(os.Stdout, "Cleaning...")

	return os.RemoveAll("./MyApp")
}
