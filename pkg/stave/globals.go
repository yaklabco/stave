//nolint:gochecknoglobals // These are all intended as constants (and are private).
package stave

import (
	"math"
	"strings"
	"text/template"

	"github.com/samber/lo"
)

func lowerFirstWord(str string) string {
	words := lo.Words(str)
	if len(words) == 0 {
		return str
	}

	firstWord := words[0]
	firstWord = strings.ToLower(firstWord)

	newStr := firstWord + lo.Substring(str, len(firstWord), math.MaxUint)

	return newStr
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
	mainFile = "stave_output_file.go"
	initFile = "stavefile.go"
)
