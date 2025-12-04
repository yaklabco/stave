//nolint:gochecknoglobals // These are all intended as constants (and are private).
package stave

import (
	"io"
	"log"
	"regexp"
	"strings"
	"text/template"
)

// (Aaaa)(Bbbb) -> aaaaBbbb.
var firstWordRx = regexp.MustCompile(`^([[:upper:]][^[:upper:]]+)([[:upper:]].*)$`)

// (AAAA)(Bbbb) -> aaaaBbbb.
var firstAbbrevRx = regexp.MustCompile(`^([[:upper:]]+)([[:upper:]][^[:upper:]].*)$`)

func lowerFirstWord(str string) string {
	if match := firstWordRx.FindStringSubmatch(str); match != nil {
		return strings.ToLower(match[1]) + match[2]
	}
	if match := firstAbbrevRx.FindStringSubmatch(str); match != nil {
		return strings.ToLower(match[1]) + match[2]
	}
	return strings.ToLower(str)
}

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
