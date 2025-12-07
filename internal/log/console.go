package log

import (
	"log"
	"os"

	"charm.land/lipgloss/v2"
	"github.com/yaklabco/stave/pkg/ui"
)

// SimpleConsoleLogger is an unstructured logger designed for emitting simple
// messages to the console in `-v`/`--verbose` mode.
//
//nolint:gochecknoglobals // This is unchanged in the course of the process lifecycle.
var SimpleConsoleLogger = log.New(os.Stderr, lipgloss.NewStyle().Foreground(ui.GetFangScheme().Flag).Render("[STAVE] "), 0)
