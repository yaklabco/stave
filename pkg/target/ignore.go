package target

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	gi "github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

var (
	ignorePatterns []gi.Pattern //nolint:gochecknoglobals // Part of a mutexed pattern.
	ignoreMatcher  gi.Matcher   //nolint:gochecknoglobals // Part of a mutexed pattern.
	ignoreStrings  []string     //nolint:gochecknoglobals // Part of a mutexed pattern.
	ignoreMu       sync.RWMutex //nolint:gochecknoglobals // Part of a mutexed pattern.
)

// AddIgnorePattern adds a pattern to the global ignorelist.
// Patterns use the same syntax as .gitignore.
func AddIgnorePattern(pattern string) error {
	return addIgnorePatternWithDomain(pattern, nil)
}

func addIgnorePatternWithDomain(pattern string, domain []string) error {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" || strings.HasPrefix(pattern, "#") {
		return nil
	}

	ignoreMu.Lock()
	defer ignoreMu.Unlock()

	ignoreStrings = append(ignoreStrings, pattern)
	// gitignore.ParsePattern takes the pattern and a domain (for anchoring).
	ignorePatterns = append(ignorePatterns, gi.ParsePattern(pattern, domain))
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

	return LoadIgnoreReader(fileDesc)
}

// LoadIgnoreReader populates the global ignorelist from an io.Reader.
// The content is expected to be in the .gitignore format.
func LoadIgnoreReader(r io.Reader) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if err := AddIgnorePattern(scanner.Text()); err != nil {
			return err
		}
	}
	return scanner.Err()
}

// LoadGitIgnore loads the .gitignore state from the current directory and its
// parents up to the nearest repository root.
func LoadGitIgnore() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// Find the repository root by looking for a .git directory.
	root := ""
	curr := cwd
	for {
		if stat, err := os.Stat(filepath.Join(curr, ".git")); err == nil && stat.IsDir() {
			root = curr
			break
		}
		parent := filepath.Dir(curr)
		if parent == curr {
			break
		}
		curr = parent
	}

	// If no .git root found, we just process from the current directory.
	// However, usually we stop at the root. If root is empty, curr is the root of the filesystem.
	start := root
	if start == "" {
		start = "/" // Fallback to filesystem root if no .git found?
		// Actually, if no .git found, maybe we should just use the current directory or nothing.
		// Git only respects .gitignore files up to the repository root.
		// If we are not in a git repo, maybe we only load from the current directory.
		start = cwd
	}

	// Collect all .gitignore files from 'start' down to 'cwd'.
	// Higher level patterns should be added first so they can be overridden by lower ones.
	var gitignoreFiles []string
	curr = cwd
	for {
		ignorePath := filepath.Join(curr, ".gitignore")
		if _, err := os.Stat(ignorePath); err == nil {
			gitignoreFiles = append(gitignoreFiles, ignorePath)
		}
		if curr == start {
			break
		}
		curr = filepath.Dir(curr)
	}

	// Reverse the list so we add patterns from top to bottom.
	for i := len(gitignoreFiles) - 1; i >= 0; i-- {
		path := gitignoreFiles[i]
		// Determine the domain (relative path from root to the directory of this .gitignore)
		dir := filepath.Dir(path)
		rel, err := filepath.Rel(start, dir)
		if err != nil {
			return err
		}
		var domain []string
		if rel != "." {
			domain = strings.Split(filepath.ToSlash(rel), "/")
		}

		if err := loadIgnoreFileWithDomain(path, domain); err != nil {
			return err
		}
	}

	return nil
}

func loadIgnoreFileWithDomain(path string, domain []string) error {
	fileDesc, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fileDesc.Close()

	scanner := bufio.NewScanner(fileDesc)
	for scanner.Scan() {
		if err := addIgnorePatternWithDomain(scanner.Text(), domain); err != nil {
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
	ignoreStrings = nil
	ignoreMatcher = nil
}

// IgnoreList returns the current state of the global ignorelist.
func IgnoreList() []string {
	ignoreMu.RLock()
	defer ignoreMu.RUnlock()
	if len(ignoreStrings) == 0 {
		return nil
	}
	return append([]string(nil), ignoreStrings...)
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
