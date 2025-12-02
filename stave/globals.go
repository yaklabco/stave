//nolint:gochecknoglobals // These are all intended as constants (and are private).
package stave

import (
	"io"
	"log"
	"strings"
	"text/template"
)

var mainfileTemplate = template.Must(template.New("").Funcs(map[string]interface{}{
	"lower": strings.ToLower,
	"lowerFirst": func(s string) string {
		parts := strings.Split(s, ":")
		for i, t := range parts {
			parts[i] = lowerFirstWord(t)
		}
		return strings.Join(parts, ":")
	},
}).Parse(staveMainfileTplString))

var initOutput = template.Must(template.New("").Parse(staveTpl))

const (
	mainfile = "stave_output_file.go"
	initFile = "stavefile.go"
)

var debug = log.New(io.Discard, "DEBUG: ", log.Ltime|log.Lmicroseconds)

// set by ldflags when you "stave build".
var (
	commitHash = "<not set>"
	timestamp  = "<not set>"
	gitTag     = "<not set>"
)
