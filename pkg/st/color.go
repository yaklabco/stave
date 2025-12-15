//nolint:gochecknoglobals // These are all intended as constants (and are private).
package st

import (
	"image/color"
	"maps"
	"slices"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
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

var ansiColorInv = lo.Invert(ansiColor)

// AnsiColorReset is an ANSI color code to reset the terminal color.
const AnsiColorReset = "\033[0m"

// DefaultTargetAnsiColor is a default ANSI color for colorizing targets.
// It is set to Cyan as an arbitrary color, because it has a neutral meaning.
var DefaultTargetAnsiColor = ansiColor[Cyan]

// noColorTERMs defines terminals that do not support ANSI color output.
// Keep this list small and conservative.
var noColorTERMs = lo.Keyify([]string{
	"dumb",
	"vt100",
	"cygwin",
	"xterm-mono",
})

func getAnsiColor(name string) (string, bool) {
	nameLower := strings.ToLower(name)
	value, ok := ansiColorByLowerString[nameLower]
	return value, ok
}

// NoColorTERMs returns the set of TERM values for which Stave should disable ANSI color output.
// This is exported so code generators can embed the same policy in stdlib-only code.
func NoColorTERMs() []string {
	terms := slices.Collect(maps.Keys(noColorTERMs))
	slices.Sort(terms)
	return terms
}

// TerminalSupportsColor returns true if the given TERM value is not in the
// known-no-color blacklist. An empty term is treated as supporting colors
// (letting Lipgloss handle further TTY detection).
func TerminalSupportsColor(term string) bool {
	if term == "" {
		return true
	}
	_, blacklisted := noColorTERMs[term]
	return !blacklisted
}

// TargetStyle returns a Lipgloss style configured with the user's target color.
// This respects STAVEFILE_TARGET_COLOR if set, otherwise uses the default cyan.
func TargetStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(targetLipglossColor())
}

// targetLipglossColor returns the Lipgloss color for targets based on env config.
func targetLipglossColor() color.Color {
	ansi := TargetColor()
	// TargetColor returns raw ANSI, extract the color code for Lipgloss.
	// Default cyan is ANSI 36.
	intVal, ok := ansiColorInv[ansi]
	if !ok {
		intVal = Cyan
	}

	return lipgloss.Color(strconv.Itoa(int(intVal)))
}
