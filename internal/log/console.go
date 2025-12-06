package log

import (
	"log"
	"os"
)

// SimpleConsoleLogger is an unstructured logger designed for emitting simple
// messages to the console in `-v`/`--verbose` mode.
//
//nolint:gochecknoglobals // This is unchanged in the course of the process lifecycle.
var SimpleConsoleLogger = log.New(os.Stderr, "[STAVE] ", 0)
