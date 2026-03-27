package target

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"sync"

	gi "github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

var (
	ignorePatterns []gi.Pattern //nolint:gochecknoglobals // Part of a mutexed pattern.
	ignoreMatcher  gi.Matcher   //nolint:gochecknoglobals // Part of a mutexed pattern.
	ignoreMu       sync.RWMutex //nolint:gochecknoglobals // Part of a mutexed pattern.
)

// AddIgnorePattern adds a pattern to the global ignorelist.
// Patterns use the same syntax as .gitignore.
func AddIgnorePattern(pattern string) error {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" || strings.HasPrefix(pattern, "#") {
		return nil
	}

	ignoreMu.Lock()
	defer ignoreMu.Unlock()

	// gitignore.ParsePattern takes the pattern and a base path (for anchoring).
	// Since we're usually matching from the current context, we use "." as base.
	ignorePatterns = append(ignorePatterns, gi.ParsePattern(pattern, nil))
	ignoreMatcher = gi.NewMatcher(ignorePatterns)
	return nil
}

// LoadIgnoreFile populates the global ignorelist from a file.
// The file is expected to be in the .gitignore format.
func LoadIgnoreFile(path string) error {
	fileDesc, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fileDesc.Close()

	scanner := bufio.NewScanner(fileDesc)
	for scanner.Scan() {
		if err := AddIgnorePattern(scanner.Text()); err != nil {
			return err
		}
	}
	return scanner.Err()
}

// ClearIgnoreList clears the global ignorelist.
func ClearIgnoreList() {
	ignoreMu.Lock()
	defer ignoreMu.Unlock()
	ignorePatterns = nil
	ignoreMatcher = nil
}

// isIgnored reports whether the given path should be ignored.
// It requires an isDir hint for accurate .gitignore-style matching.
func isIgnored(path string, isDir bool) bool {
	ignoreMu.RLock()
	defer ignoreMu.RUnlock()

	if ignoreMatcher == nil {
		return false
	}

	// go-git's matcher expects path components.
	// We use filepath.ToSlash for portability and then split.
	// The matcher expects a slice of strings representing the path components.
	parts := strings.Split(filepath.ToSlash(path), "/")

	// Filter out empty parts (e.g., from leading/multiple/trailing slashes).
	cleanParts := parts[:0]
	for _, p := range parts {
		if p != "" {
			cleanParts = append(cleanParts, p)
		}
	}

	return ignoreMatcher.Match(cleanParts, isDir)
}
