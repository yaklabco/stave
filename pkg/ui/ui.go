package ui

import (
	"os"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/fang"
)

// GetFangScheme returns the same light/dark-aware color scheme fang uses.
func GetFangScheme() fang.ColorScheme {
	// This mirrors fang.mustColorscheme(DefaultColorScheme)
	isDark := lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
	return fang.DefaultColorScheme(lipgloss.LightDark(isDark))
}

// GetBlockStyles generates reusable styles for titles and code block elements.
// Returns two lipgloss.Style objects: one for titles and one for blocks.
func GetBlockStyles() (lipgloss.Style, lipgloss.Style) {
	cs := GetFangScheme()

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(cs.QuotedString).
		Transform(strings.ToUpper).
		Padding(1, 0).
		Margin(0, 2)

	blockStyle := lipgloss.NewStyle().
		Background(cs.Codeblock).
		Foreground(cs.Base).
		MarginLeft(2).
		Padding(1, 2)
	return titleStyle, blockStyle
}
