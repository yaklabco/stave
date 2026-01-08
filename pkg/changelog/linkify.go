package changelog

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/samber/lo"
)

const (
	unreleasedMarkerString = "unreleased"
)

var repoURLPattern = regexp.MustCompile(`(https://github\.com/[^/]+(/[^/]+)+)/compare/(([^/]+/)*)v[0-9]+\.[0-9]+\.[0-9]+`)

// Linkify reads the content of the changelog file, runs LinkifyContent on it,
// and saves the result back to the file.
func Linkify(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	newContent, err := LinkifyContent(string(content))
	if err != nil {
		return err
	}

	return os.WriteFile(path, []byte(newContent), info.Mode())
}

// LinkifyContent adds missing link reference definitions to the bottom of the changelog.
func LinkifyContent(content string) (string, error) {
	cl, err := Parse(content)
	if err != nil {
		return "", err
	}

	// 1. Identify headings that need links (missing or update required)
	toLinkify := make(map[string]struct{})
	for _, h := range cl.Headings {
		if !cl.HasLinkForVersion(h.Name) {
			toLinkify[h.Name] = struct{}{}
		}
	}

	// If the newly-linkified heading is the topmost one under "Unreleased",
	// then the "Unreleased" Link Reference Definition should be updated.
	if len(cl.Headings) > 1 && strings.ToLower(cl.Headings[0].Name) == unreleasedMarkerString {
		if lo.HasKey(toLinkify, cl.Headings[1].Name) {
			toLinkify[cl.Headings[0].Name] = struct{}{}
		}
	}

	if len(toLinkify) == 0 {
		return content, nil
	}

	// 2. Determine base URL from existing links
	baseURL := ""
	tagPrefix := ""
	for _, l := range cl.Links {
		if matches := repoURLPattern.FindStringSubmatch(l.URL); len(matches) > 3 {
			baseURL = matches[1]
			tagPrefix = matches[3]
			break
		}
	}

	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		return "", errors.New("could not determine base URL from existing links")
	}

	// 3. Generate links
	newLinks := make(map[string]string)
	for iHeading, theHeading := range cl.Headings {
		if !lo.HasKey(toLinkify, theHeading.Name) {
			continue
		}

		var link string
		if strings.ToLower(theHeading.Name) == unreleasedMarkerString {
			// Find next heading for comparison
			if iHeading+1 >= len(cl.Headings) {
				continue
			}
			next := cl.Headings[iHeading+1].Name
			link = fmt.Sprintf("%s/compare/%sv%s...HEAD", baseURL, tagPrefix, next)
		} else {
			// It's a version. Compare to the next version underneath.
			if iHeading+1 < len(cl.Headings) {
				next := cl.Headings[iHeading+1].Name
				link = fmt.Sprintf("%s/compare/%sv%s...%sv%s", baseURL, tagPrefix, next, tagPrefix, theHeading.Name)
			} else {
				// Last version. link to the tag.
				link = fmt.Sprintf("%s/releases/tag/%sv%s", baseURL, tagPrefix, theHeading.Name)
			}
		}
		newLinks[theHeading.Name] = link
	}

	// 4. Insert links in order
	lines := strings.Split(strings.TrimSuffix(content, "\n"), "\n")

	// Find where links start
	firstLinkLine := -1
	lastLinkLine := -1
	for _, l := range cl.Links {
		if firstLinkLine == -1 || l.Line < firstLinkLine {
			firstLinkLine = l.Line
		}
		if l.Line > lastLinkLine {
			lastLinkLine = l.Line
		}
	}

	// If no links, add to the end
	if firstLinkLine == -1 {
		lastNonEmpty := len(lines) - 1
		for lastNonEmpty >= 0 && strings.TrimSpace(lines[lastNonEmpty]) == "" {
			lastNonEmpty--
		}

		var result []string
		result = append(result, lines[:lastNonEmpty+1]...)
		result = append(result, "") // spacing
		for _, h := range cl.Headings {
			if link, ok := newLinks[h.Name]; ok {
				name := h.Name
				if strings.ToLower(name) == unreleasedMarkerString {
					name = unreleasedMarkerString
				}
				result = append(result, fmt.Sprintf("[%s]: %s", name, link))
			}
		}
		return strings.Join(result, "\n") + "\n", nil
	}

	// Insert new links among existing ones
	var finalLinks []string
	headingToLink := make(map[string]string)
	for _, l := range cl.Links {
		headingToLink[strings.ToLower(l.Name)] = l.URL
	}
	for name, url := range newLinks {
		headingToLink[strings.ToLower(name)] = url
	}

	for _, h := range cl.Headings {
		name := strings.ToLower(h.Name)
		if url, ok := headingToLink[name]; ok {
			finalLinks = append(finalLinks, fmt.Sprintf("[%s]: %s", name, url))
		}
	}

	// Replace the link block
	var result []string
	result = append(result, lines[:firstLinkLine-1]...)
	result = append(result, finalLinks...)

	if lastLinkLine < len(lines) {
		result = append(result, lines[lastLinkLine:]...)
	}

	return strings.Join(result, "\n") + "\n", nil
}
