package target

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestIgnore(t *testing.T) {
	// Don't use t.Parallel() because it uses global state

	dir, err := os.MkdirTemp("", "stave-ignore-test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	// Create a file to ignore
	ignoredFile := filepath.Join(dir, "ignored.txt")
	require.NoError(t, os.WriteFile(ignoredFile, []byte("ignore me"), 0644))

	// Create a file to keep
	keptFile := filepath.Join(dir, "kept.txt")
	require.NoError(t, os.WriteFile(keptFile, []byte("keep me"), 0644))

	// Create a subdirectory to ignore
	ignoredDir := filepath.Join(dir, "ignored_dir")
	require.NoError(t, os.Mkdir(ignoredDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(ignoredDir, "file.txt"), []byte("ignore me too"), 0644))

	// Set a reference time
	refTime := time.Now().Add(-10 * time.Minute)

	// Initially, nothing is ignored
	newer, err := DirNewer(refTime, dir)
	require.NoError(t, err)
	require.True(t, newer, "Initially should detect new files")

	// Add ignore patterns
	ClearIgnoreList()
	require.NoError(t, AddIgnorePattern("ignored.txt"))
	require.NoError(t, AddIgnorePattern("ignored_dir"))

	// Check if ignored.txt is ignored by PathNewer
	newer, err = PathNewer(refTime, ignoredFile)
	require.NoError(t, err)
	require.False(t, newer, "ignored.txt should be ignored")

	// Check if kept.txt is NOT ignored
	newer, err = PathNewer(refTime, keptFile)
	require.NoError(t, err)
	require.True(t, newer, "kept.txt should NOT be ignored")

	// Check DirNewer with only ignored files
	// Make sure kept.txt and the base dir are OLD
	oldTime := time.Now().Add(-20 * time.Minute)
	require.NoError(t, os.Chtimes(keptFile, oldTime, oldTime))
	require.NoError(t, os.Chtimes(dir, oldTime, oldTime))

	newer, err = DirNewer(refTime, dir)
	require.NoError(t, err)
	require.False(t, newer, "DirNewer should ignore files according to patterns")

	ClearIgnoreList()
}

func TestIgnoreList(t *testing.T) {
	ClearIgnoreList()
	defer ClearIgnoreList()

	require.Nil(t, IgnoreList(), "Empty list should return nil")

	patterns := []string{"*.log", "temp/", "!important.log"}
	for _, p := range patterns {
		require.NoError(t, AddIgnorePattern(p))
	}

	require.Equal(t, patterns, IgnoreList(), "Should return all added patterns")

	ClearIgnoreList()
	require.Nil(t, IgnoreList(), "Should be nil after clear")
}

func TestGitignoreSyntax(t *testing.T) {
	tests := []struct {
		pattern  string
		path     string
		isDir    bool
		expected bool
	}{
		// Basic file match
		{"hello.txt", "hello.txt", false, true},
		{"hello.txt", "foo/hello.txt", false, true},
		{"hello.txt", "foo/bar/hello.txt", false, true},
		{"hello.txt", "hello.txt.bak", false, false},

		// Directory match
		{"temp/", "temp/foo.txt", false, true},
		{"temp/", "a/temp/foo.txt", false, true},
		{"temp/", "temptemp/foo.txt", false, false},
		{"temp/", "temp", true, true},

		// Wildcards
		{"*.log", "error.log", false, true},
		{"*.log", "logs/error.log", false, true},
		{"*.log", "error.log.txt", false, false},

		// Negation
		{"*.log", "important.log", false, true},
		{"!important.log", "important.log", false, false},

		// Absolute path (starts with /)
		{"/root.txt", "root.txt", false, true},
		{"/root.txt", "subdir/root.txt", false, false},

		// Double asterisk
		{"foo/**/bar", "foo/bar", false, true},
		{"foo/**/bar", "foo/a/bar", false, true},
		{"foo/**/bar", "foo/a/b/bar", false, true},
		{"foo/**/bar", "other/foo/bar", false, false},

		// Nested negation
		{"dir/*", "dir/a", false, true},
		{"!dir/a", "dir/a", false, false},

		// Escaping
		{"\\#foo", "#foo", false, true},
		{"\\!bar", "!bar", false, true},

		// Spaces (Trimmed by stave before passing to go-git)
		{"  spaces.txt  ", "spaces.txt", false, true},

		// Directory match with slash in middle
		{"a/b/", "a/b/c", false, true},
		{"a/b/", "x/a/b/c", false, false},
	}

	for _, tc := range tests {
		t.Run(tc.pattern+"_match_"+tc.path, func(t *testing.T) {
			ClearIgnoreList()
			err := AddIgnorePattern(tc.pattern)
			require.NoError(t, err)

			actual := isIgnored(tc.path, tc.isDir)
			require.Equal(t, tc.expected, actual, "Pattern %q should match %q (isDir=%v): %v", tc.pattern, tc.path, tc.isDir, tc.expected)
		})
	}
	ClearIgnoreList()
}

func TestLoadIgnoreFile(t *testing.T) {
	dir, err := os.MkdirTemp("", "stave-ignore-file-test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	ignoreFilePath := filepath.Join(dir, ".staveignore")
	content := `
# a comment
*.log
temp/
!important.log
`
	require.NoError(t, os.WriteFile(ignoreFilePath, []byte(content), 0644))

	ClearIgnoreList()
	err = LoadIgnoreFile(ignoreFilePath)
	require.NoError(t, err)

	require.True(t, isIgnored("test.log", false))
	require.True(t, isIgnored("temp/foo.txt", false))
	require.True(t, isIgnored("temp", true))
	require.False(t, isIgnored("test.txt", false))
	require.False(t, isIgnored("important.log", false))

	ClearIgnoreList()
}

func TestLoadGitIgnore(t *testing.T) {
	// Setup a nested structure:
	// root/
	//   .git/ (dummy)
	//   .gitignore (Pattern: *.log)
	//   subdir/
	//     .gitignore (Pattern: !important.log)
	//     important.log
	//     other.log

	tmpDir, err := os.MkdirTemp("", "stave-loadgitignore-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	root := filepath.Join(tmpDir, "root")
	require.NoError(t, os.Mkdir(root, 0755))
	require.NoError(t, os.Mkdir(filepath.Join(root, ".git"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(root, ".gitignore"), []byte("*.log\n"), 0644))

	subdir := filepath.Join(root, "subdir")
	require.NoError(t, os.Mkdir(subdir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(subdir, ".gitignore"), []byte("!important.log\n"), 0644))

	// Change working directory to subdir
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(subdir))
	defer func() { _ = os.Chdir(oldCwd) }()

	ClearIgnoreList()
	err = LoadGitIgnore()
	require.NoError(t, err)

	// In subdir/
	// other.log should be ignored (from root/.gitignore)
	// important.log should NOT be ignored (overridden by subdir/.gitignore)
	require.True(t, isIgnored("other.log", false), "other.log should be ignored")
	require.False(t, isIgnored("subdir/important.log", false), "important.log should NOT be ignored")

	// Verify patterns accumulated:
	patterns := IgnoreList()
	require.Contains(t, patterns, "*.log")
	require.Contains(t, patterns, "!important.log")
	// Order should be parent first, then subdir.
	require.Equal(t, []string{"*.log", "!important.log"}, patterns)

	// Test with path from root
	require.True(t, isIgnored("subdir/other.log", false), "subdir/other.log should be ignored")
	require.False(t, isIgnored("subdir/important.log", false), "subdir/important.log should NOT be ignored")

	ClearIgnoreList()
}
