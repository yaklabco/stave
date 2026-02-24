//nolint:gochecknoglobals // These are all intended as constants (and are private).
package stave

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	"github.com/samber/lo"
)

const (
	hashLengthLimit = 16
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

var mainfileTemplate = template.Must(template.New("").Funcs(map[string]any{
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
	// mainFileBase is the base prefix used for generated mainfile names.
	mainFileBase = "stave_output_file"
	initFile     = "stavefile.go"
)

// mainFilePathFromExePath derives a generated main filename from the
// computed executable path. It uses a short hash prefix from the exe name,
// and the current process ID.
func mainFilePathFromExePath(dir, exePath string) string {
	base := filepath.Base(exePath)
	if runtime.GOOS == "windows" && strings.HasSuffix(base, ".exe") {
		base = strings.TrimSuffix(base, ".exe")
	}
	// keep it reasonably short while still unique enough
	hash := base
	if len(hash) > hashLengthLimit {
		hash = hash[:hashLengthLimit]
	}
	return filepath.Join(dir, fmt.Sprintf("%s_%s_%d.go", mainFileBase, hash, os.Getpid()))
}
