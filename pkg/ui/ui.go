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

// UI layout constants.
const (
	defaultMargin  = 2
	defaultPadding = 2
)

// GetBlockStyles generates reusable styles for titles and code block elements.
// Returns two lipgloss.Style objects: one for titles and one for blocks.
func GetBlockStyles() (lipgloss.Style, lipgloss.Style) {
	colorScheme := GetFangScheme()

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(colorScheme.QuotedString).
		Transform(strings.ToUpper).
		Padding(1, 0).
		Margin(0, defaultMargin)

	blockStyle := lipgloss.NewStyle().
		Background(colorScheme.Codeblock).
		Foreground(colorScheme.Base).
		MarginLeft(defaultMargin).
		Padding(1, defaultPadding)
	return titleStyle, blockStyle
}
