//nolint:gochecknoglobals // These are all intended as constants (and are private).
package parse

import (
	"io"
	"log"
)

var debug = log.New(io.Discard, "DEBUG: ", log.Ltime|log.Lmicroseconds)

var argTypes = map[string]string{
	"string":           "string",
	"int":              "int",
	"float64":          "float64",
	"&{time Duration}": "time.Duration",
	"bool":             "bool",
}
