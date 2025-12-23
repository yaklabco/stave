package changelog

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// ExtractSection extracts a section from the changelog file and writes it to outputFile.
// if section is empty, it extracts the most recent numbered (non-Unreleased) section.
func ExtractSection(inputFile, outputFile, section string) error {
	content, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	cl, err := Parse(string(content))
	if err != nil {
		return fmt.Errorf("failed to parse changelog: %w", err)
	}

	var targetHeading *Heading
	if section == "" {
		// Find most recent numbered section
		for _, h := range cl.Headings {
			if h.IsRelease {
				targetHeading = &h
				break
			}
		}
	} else {
		// Find specific section
		for _, h := range cl.Headings {
			if strings.EqualFold(h.Name, section) {
				targetHeading = &h
				break
			}
		}
	}

	if targetHeading == nil {
		if section == "" {
			return errors.New("no numbered section found in changelog")
		}
		return fmt.Errorf("section %q not found in changelog", section)
	}

	// Find the end line of the section
	// It's the line before the next heading or the line before the first link, or the end of the file.
	lines := strings.Split(string(content), "\n")
	startLine := targetHeading.Line - 1 // 0-indexed for slice
	endLine := len(lines)

	// Check for next heading
	for _, h := range cl.Headings {
		if h.Line > targetHeading.Line {
			if h.Line-1 < endLine {
				endLine = h.Line - 1
			}
			// Headings are usually in order, so we could probably break,
			// but just to be safe we check all.
		}
	}

	// Check for links
	for _, l := range cl.Links {
		if l.Line > targetHeading.Line {
			if l.Line-1 < endLine {
				endLine = l.Line - 1
			}
		}
	}

	// Extract lines
	sectionLines := lines[startLine:endLine]

	// Trim trailing empty lines from the extracted section
	for len(sectionLines) > 0 && strings.TrimSpace(sectionLines[len(sectionLines)-1]) == "" {
		sectionLines = sectionLines[:len(sectionLines)-1]
	}

	outputContent := strings.Join(sectionLines, "\n")
	if len(sectionLines) > 0 {
		outputContent += "\n"
	}

	err = os.WriteFile(outputFile, []byte(outputContent), 0o600)
	if err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}
