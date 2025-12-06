//nolint:gochecknoglobals // These are all intended as constants (and are private).
package st

import (
	"strings"

	"github.com/samber/lo"
)

// Color is ANSI color type.
type Color int

//go:generate go tool golang.org/x/tools/cmd/stringer -type=Color
const (
	Black Color = iota
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White
	BrightBlack
	BrightRed
	BrightGreen
	BrightYellow
	BrightBlue
	BrightMagenta
	BrightCyan
	BrightWhite
)

// AnsiColor are ANSI color codes for supported terminal colors.
var ansiColor = map[Color]string{
	Black:         "\u001b[30m",
	Red:           "\u001b[31m",
	Green:         "\u001b[32m",
	Yellow:        "\u001b[33m",
	Blue:          "\u001b[34m",
	Magenta:       "\u001b[35m",
	Cyan:          "\u001b[36m",
	White:         "\u001b[37m",
	BrightBlack:   "\u001b[30;1m",
	BrightRed:     "\u001b[31;1m",
	BrightGreen:   "\u001b[32;1m",
	BrightYellow:  "\u001b[33;1m",
	BrightBlue:    "\u001b[34;1m",
	BrightMagenta: "\u001b[35;1m",
	BrightCyan:    "\u001b[36;1m",
	BrightWhite:   "\u001b[37;1m",
}

var ansiColorByLowerString = lo.MapKeys(ansiColor, func(_ string, key Color) string {
	return strings.ToLower(key.String())
})

// AnsiColorReset is an ANSI color code to reset the terminal color.
const AnsiColorReset = "\033[0m"

// DefaultTargetAnsiColor is a default ANSI color for colorizing targets.
// It is set to Cyan as an arbitrary color, because it has a neutral meaning.
var DefaultTargetAnsiColor = ansiColor[Cyan]

func getAnsiColor(color string) (string, bool) {
	colorLower := strings.ToLower(color)

	value, ok := ansiColorByLowerString[colorLower]

	return value, ok
}
