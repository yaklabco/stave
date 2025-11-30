package stave

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLowerFirstWord(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		s    string
		want string
	}{
		{
			name: "FooBar to fooBar",
			s:    "FooBar",
			want: "fooBar",
		},
		{
			name: "HTTPServer to httpServer",
			s:    "HTTPServer",
			want: "httpServer",
		},
		{
			name: "FOO to foo",
			s:    "FOO",
			want: "foo",
		},
		{
			name: "foo to foo",
			s:    "foo",
			want: "foo",
		},
		{
			name: "empty string",
			s:    "",
			want: "",
		},
		{
			name: "A to a",
			s:    "A",
			want: "a",
		},
		{
			name: "ABC to abc",
			s:    "ABC",
			want: "abc",
		},
		{
			name: "ABCDef to abcDef",
			s:    "ABCDef",
			want: "abcDef",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := lowerFirstWord(tt.s)
			if got != tt.want {
				t.Errorf("lowerFirstWord(%q) = %q, want %q", tt.s, got, tt.want)
			}
		})
	}
}

func TestFilter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		list   []string
		prefix string
		want   []string
	}{
		{
			name:   "matching prefix",
			list:   []string{"STAVEFILE_VERBOSE=1", "STAVEFILE_DEBUG=1", "PATH=/usr/bin"},
			prefix: "STAVEFILE",
			want:   []string{"STAVEFILE_VERBOSE=1", "STAVEFILE_DEBUG=1"},
		},
		{
			name:   "no matches",
			list:   []string{"PATH=/usr/bin", "HOME=/home/user"},
			prefix: "STAVEFILE",
			want:   []string{},
		},
		{
			name:   "empty list",
			list:   []string{},
			prefix: "STAVEFILE",
			want:   []string{},
		},
		{
			name:   "all match",
			list:   []string{"STAVEFILE_A=1", "STAVEFILE_B=2"},
			prefix: "STAVEFILE",
			want:   []string{"STAVEFILE_A=1", "STAVEFILE_B=2"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := filter(tt.list, tt.prefix)
			if len(got) != len(tt.want) {
				t.Errorf("filter() returned %d items, want %d", len(got), len(tt.want))
			}
			for i := range got {
				if i >= len(tt.want) {
					break
				}
				if got[i] != tt.want[i] {
					t.Errorf("filter()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestHashFile(t *testing.T) {
	t.Parallel()

	setupFile := func(t *testing.T, content string) string {
		t.Helper()
		tmpfile, err := os.CreateTemp("", "test-*.txt")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := tmpfile.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
		if err := tmpfile.Close(); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { os.Remove(tmpfile.Name()) })
		return tmpfile.Name()
	}

	t.Run("valid file", func(t *testing.T) {
		t.Parallel()

		filename := setupFile(t, "test content")
		hash, err := hashFile(filename)
		if err != nil {
			t.Errorf("hashFile() error = %v", err)
		}
		if hash == "" {
			t.Error("hashFile() returned empty hash")
		}
	})

	t.Run("non-existent file error", func(t *testing.T) {
		t.Parallel()

		_, err := hashFile("/nonexistent/file/xyz123.txt")
		if err == nil {
			t.Error("hashFile() should return error for non-existent file")
		}
	})

	t.Run("empty file", func(t *testing.T) {
		t.Parallel()

		filename := setupFile(t, "")
		hash, err := hashFile(filename)
		if err != nil {
			t.Errorf("hashFile() error = %v", err)
		}
		if hash == "" {
			t.Error("hashFile() returned empty hash for empty file")
		}
	})

	t.Run("same content same hash", func(t *testing.T) {
		t.Parallel()

		file1 := setupFile(t, "same content")
		file2 := setupFile(t, "same content")

		hash1, err := hashFile(file1)
		if err != nil {
			t.Fatal(err)
		}

		hash2, err := hashFile(file2)
		if err != nil {
			t.Fatal(err)
		}

		if hash1 != hash2 {
			t.Errorf("hashFile() produced different hashes for same content: %q != %q", hash1, hash2)
		}
	})

	t.Run("different content different hash", func(t *testing.T) {
		t.Parallel()

		file1 := setupFile(t, "content one")
		file2 := setupFile(t, "content two")

		hash1, err := hashFile(file1)
		if err != nil {
			t.Fatal(err)
		}

		hash2, err := hashFile(file2)
		if err != nil {
			t.Fatal(err)
		}

		if hash1 == hash2 {
			t.Error("hashFile() produced same hash for different content")
		}
	})
}

func TestRemoveContents(t *testing.T) {
	t.Parallel()

	setupTempDir := func(t *testing.T) string {
		t.Helper()
		dir, err := os.MkdirTemp("", "stave-test-*")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { os.RemoveAll(dir) })
		return dir
	}

	t.Run("removes files preserves subdirs", func(t *testing.T) {
		t.Parallel()

		dir := setupTempDir(t)

		// Create some files
		for i := 0; i < 3; i++ {
			filename := filepath.Join(dir, "file"+string(rune('0'+i))+".txt")
			if err := os.WriteFile(filename, []byte("test"), 0644); err != nil {
				t.Fatal(err)
			}
		}

		// Create a subdirectory with a file
		subdir := filepath.Join(dir, "subdir")
		if err := os.Mkdir(subdir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(subdir, "subfile.txt"), []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}

		err := removeContents(dir)
		if err != nil {
			t.Fatalf("removeContents() error = %v", err)
		}

		// Check that files are removed
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatal(err)
		}

		fileCount := 0
		dirCount := 0
		for _, entry := range entries {
			if entry.IsDir() {
				dirCount++
			} else {
				fileCount++
			}
		}

		if fileCount != 0 {
			t.Errorf("removeContents() left %d files, want 0", fileCount)
		}

		if dirCount == 0 {
			t.Error("removeContents() removed subdirectories")
		}

		// Verify the subdirectory still exists
		if _, err := os.Stat(subdir); os.IsNotExist(err) {
			t.Error("removeContents() removed subdirectory")
		}
	})

	t.Run("non-existent dir no error", func(t *testing.T) {
		t.Parallel()

		err := removeContents("/nonexistent/directory/xyz123")
		if err != nil {
			t.Errorf("removeContents() on non-existent dir error = %v, want nil", err)
		}
	})
}

