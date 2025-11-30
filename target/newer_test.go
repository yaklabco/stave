package target

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewestModTime(t *testing.T) {
	t.Parallel()
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("error creating temp dir: %s", err.Error())
	}
	defer os.RemoveAll(dir)
	for _, name := range []string{"a", "b", "c", "d"} {
		out := filepath.Join(dir, name)
		if err := ioutil.WriteFile(out, []byte("hi!"), 0644); err != nil {
			t.Fatalf("error writing file: %s", err.Error())
		}
	}
	time.Sleep(10 * time.Millisecond)
	outName := filepath.Join(dir, "c")
	outfh, err := os.OpenFile(outName, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("error opening file to append: %s", err.Error())
	}
	if _, err := outfh.WriteString("\nbye!\n"); err != nil {
		t.Fatalf("error appending to file: %s", err.Error())
	}
	if err := outfh.Close(); err != nil {
		t.Fatalf("error closing file: %s", err.Error())
	}

	afi, err := os.Stat(filepath.Join(dir, "a"))
	if err != nil {
		t.Fatalf("error stating unmodified file: %s", err.Error())
	}

	cfi, err := os.Stat(outName)
	if err != nil {
		t.Fatalf("error stating modified file: %s", err.Error())
	}
	if afi.ModTime().Equal(cfi.ModTime()) {
		t.Fatal("modified and unmodified file mtimes equal")
	}

	newest, err := NewestModTime(dir)
	if err != nil {
		t.Fatalf("error finding newest mod time: %s", err.Error())
	}
	if !newest.Equal(cfi.ModTime()) {
		t.Fatal("expected newest mod time to match c")
	}
}

func TestOldestModTime(t *testing.T) {
	t.Parallel()
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("error creating temp dir: %s", err.Error())
	}
	defer os.RemoveAll(dir)
	for _, name := range []string{"a", "b", "c", "d"} {
		out := filepath.Join(dir, name)
		if err := ioutil.WriteFile(out, []byte("hi!"), 0644); err != nil {
			t.Fatalf("error writing file: %s", err.Error())
		}
	}
	time.Sleep(10 * time.Millisecond)
	for _, name := range []string{"a", "b", "d"} {
		outName := filepath.Join(dir, name)
		outfh, err := os.OpenFile(outName, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			t.Fatalf("error opening file to append: %s", err.Error())
		}
		if _, err := outfh.WriteString("\nbye!\n"); err != nil {
			t.Fatalf("error appending to file: %s", err.Error())
		}
		if err := outfh.Close(); err != nil {
			t.Fatalf("error closing file: %s", err.Error())
		}
	}

	afi, err := os.Stat(filepath.Join(dir, "a"))
	if err != nil {
		t.Fatalf("error stating unmodified file: %s", err.Error())
	}

	outName := filepath.Join(dir, "c")
	cfi, err := os.Stat(outName)
	if err != nil {
		t.Fatalf("error stating modified file: %s", err.Error())
	}
	if afi.ModTime().Equal(cfi.ModTime()) {
		t.Fatal("modified and unmodified file mtimes equal")
	}

	newest, err := OldestModTime(dir)
	if err != nil {
		t.Fatalf("error finding oldest mod time: %s", err.Error())
	}
	if !newest.Equal(cfi.ModTime()) {
		t.Fatal("expected newest mod time to match c")
	}
}

func TestDirNewerDirect(t *testing.T) {
	t.Parallel()

	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// Create an old file
	oldFile := filepath.Join(dir, "old.txt")
	if err := ioutil.WriteFile(oldFile, []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}

	oldStat, err := os.Stat(oldFile)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(10 * time.Millisecond)

	// Create a new file
	newFile := filepath.Join(dir, "new.txt")
	if err := ioutil.WriteFile(newFile, []byte("new"), 0644); err != nil {
		t.Fatal(err)
	}

	newStat, err := os.Stat(newFile)
	if err != nil {
		t.Fatal(err)
	}

	// Test source newer
	newer, err := DirNewer(oldStat.ModTime(), newFile)
	if err != nil {
		t.Fatalf("DirNewer() error = %v", err)
	}
	if !newer {
		t.Error("DirNewer() should return true when source is newer")
	}

	// Test source older
	newer, err = DirNewer(newStat.ModTime(), oldFile)
	if err != nil {
		t.Fatalf("DirNewer() error = %v", err)
	}
	if newer {
		t.Error("DirNewer() should return false when source is older")
	}

	// Test walk error
	_, err = DirNewer(time.Now(), filepath.Join(dir, "nonexistent"))
	if err == nil {
		t.Error("DirNewer() should return error for non-existent path")
	}
}

func TestGlobNewerDirect(t *testing.T) {
	t.Parallel()

	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// Create test files
	for i := 0; i < 3; i++ {
		filename := filepath.Join(dir, fmt.Sprintf("test%d.txt", i))
		if err := ioutil.WriteFile(filename, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
		time.Sleep(5 * time.Millisecond)
	}

	// Get the time of the second file
	file1 := filepath.Join(dir, "test1.txt")
	stat, err := os.Stat(file1)
	if err != nil {
		t.Fatal(err)
	}

	// Test matching glob
	glob := filepath.Join(dir, "test*.txt")
	newer, err := GlobNewer(stat.ModTime(), glob)
	if err != nil {
		t.Fatalf("GlobNewer() error = %v", err)
	}
	if !newer {
		t.Error("GlobNewer() should return true when glob matches newer files")
	}

	// Test empty glob error
	emptyGlob := filepath.Join(dir, "nonexistent*.txt")
	_, err = GlobNewer(time.Now(), emptyGlob)
	if err == nil {
		t.Error("GlobNewer() should return error for empty glob")
	}

	// Test multiple globs
	glob1 := filepath.Join(dir, "test0.txt")
	glob2 := filepath.Join(dir, "test2.txt")
	newer, err = GlobNewer(stat.ModTime(), glob1, glob2)
	if err != nil {
		t.Fatalf("GlobNewer() with multiple globs error = %v", err)
	}
	if !newer {
		t.Error("GlobNewer() should return true when any glob matches newer file")
	}
}

func TestPathNewerDirect(t *testing.T) {
	t.Parallel()

	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// Create an old file
	oldFile := filepath.Join(dir, "old.txt")
	if err := ioutil.WriteFile(oldFile, []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}

	oldStat, err := os.Stat(oldFile)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(10 * time.Millisecond)

	// Create a new file
	newFile := filepath.Join(dir, "new.txt")
	if err := ioutil.WriteFile(newFile, []byte("new"), 0644); err != nil {
		t.Fatal(err)
	}

	newStat, err := os.Stat(newFile)
	if err != nil {
		t.Fatal(err)
	}

	// Test source newer
	newer, err := PathNewer(oldStat.ModTime(), newFile)
	if err != nil {
		t.Fatalf("PathNewer() error = %v", err)
	}
	if !newer {
		t.Error("PathNewer() should return true when source is newer")
	}

	// Test source older
	newer, err = PathNewer(newStat.ModTime(), oldFile)
	if err != nil {
		t.Fatalf("PathNewer() error = %v", err)
	}
	if newer {
		t.Error("PathNewer() should return false when source is older")
	}

	// Test missing source error
	_, err = PathNewer(time.Now(), filepath.Join(dir, "nonexistent"))
	if err == nil {
		t.Error("PathNewer() should return error for non-existent source")
	}
}

func TestOldestModTimeError(t *testing.T) {
	t.Parallel()

	// Test non-existent path
	_, err := OldestModTime("/this/path/does/not/exist/xyz123")
	if err == nil {
		t.Error("OldestModTime() should return error for non-existent path")
	}
}

func TestNewestModTimeError(t *testing.T) {
	t.Parallel()

	// Test non-existent path
	_, err := NewestModTime("/this/path/does/not/exist/xyz123")
	if err == nil {
		t.Error("NewestModTime() should return error for non-existent path")
	}
}
