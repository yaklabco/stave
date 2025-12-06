package sh_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/yaklabco/stave/pkg/sh"
)

// compareFiles checks that two files are identical for testing purposes. That means they have the same length,
// the same contents, and the same permissions. It does NOT mean they have the same timestamp, as that is expected
// to change in normal Stave sh.Copy operation.
func compareFiles(file1 string, file2 string) error {
	stat1, err := os.Stat(file1)
	if err != nil {
		return fmt.Errorf("can't stat %s: %w", file1, err)
	}
	stat2, err := os.Stat(file2)
	if err != nil {
		return fmt.Errorf("can't stat %s: %w", file2, err)
	}
	if stat1.Size() != stat2.Size() {
		return fmt.Errorf("files %s and %s have different sizes: %d vs %d", file1, file2, stat1.Size(), stat2.Size())
	}
	if stat1.Mode() != stat2.Mode() {
		return fmt.Errorf(
			"files %s and %s have different permissions: %#4o vs %#4o",
			file1, file2, stat1.Mode(), stat2.Mode(),
		)
	}
	f1bytes, err := os.ReadFile(file1)
	if err != nil {
		return fmt.Errorf("can't read %s: %w", file1, err)
	}
	f2bytes, err := os.ReadFile(file2)
	if err != nil {
		return fmt.Errorf("can't read %s: %w", file2, err)
	}
	if !bytes.Equal(f1bytes, f2bytes) {
		return fmt.Errorf("files %s and %s have different contents", file1, file2)
	}
	return nil
}

func TestHelpers(t *testing.T) {
	mytmpdir, err := os.MkdirTemp("", "stave")
	if err != nil {
		t.Fatalf("can't create test directory: %v", err)
	}
	defer func() {
		derr := os.RemoveAll(mytmpdir)
		if derr != nil {
			t.Errorf("error cleaning up after TestHelpers: %v", derr)
		}
	}()
	srcname := filepath.Join(mytmpdir, "test1.txt")
	//#nosec G306 -- test file does not require restricted permissions.
	err = os.WriteFile(srcname, []byte("All work and no play makes Jack a dull boy."), 0o644)
	if err != nil {
		t.Fatalf("can't create test file %s: %v", srcname, err)
	}
	destname := filepath.Join(mytmpdir, "test2.txt")

	t.Run("sh/copy", func(t *testing.T) {
		cerr := sh.Copy(destname, srcname)
		if cerr != nil {
			t.Errorf("test file copy from %s to %s failed: %v", srcname, destname, cerr)
		}
		cerr = compareFiles(srcname, destname)
		if cerr != nil {
			t.Errorf("test file copy verification failed: %v", cerr)
		}
	})

	// While we've got a temporary directory, test how forgiving sh.Rm is
	t.Run("sh/rm/ne", func(t *testing.T) {
		nef := filepath.Join(mytmpdir, "file_not_exist.txt")
		rerr := sh.Rm(nef)
		if rerr != nil {
			t.Errorf("sh.Rm complained when removing nonexistent file %s: %v", nef, rerr)
		}
	})

	t.Run("sh/copy/ne", func(t *testing.T) {
		nef := filepath.Join(mytmpdir, "file_not_exist.txt")
		nedf := filepath.Join(mytmpdir, "file_not_exist2.txt")
		cerr := sh.Copy(nedf, nef)
		if cerr == nil {
			t.Errorf("sh.Copy succeeded copying nonexistent file %s", nef)
		}
	})

	// We test sh.Rm by clearing up our own test files and directories
	t.Run("sh/rm", func(t *testing.T) {
		rerr := sh.Rm(destname)
		if rerr != nil {
			t.Errorf("failed to remove file %s: %v", destname, rerr)
		}
		rerr = sh.Rm(srcname)
		if rerr != nil {
			t.Errorf("failed to remove file %s: %v", srcname, rerr)
		}
		rerr = sh.Rm(mytmpdir)
		if rerr != nil {
			t.Errorf("failed to remove dir %s: %v", mytmpdir, rerr)
		}
		_, rerr = os.Stat(mytmpdir)
		if rerr == nil {
			t.Errorf("removed dir %s but it's still there?", mytmpdir)
		}
	})

	t.Run("sh/rm/nedir", func(t *testing.T) {
		rerr := sh.Rm(mytmpdir)
		if rerr != nil {
			t.Errorf("sh.Rm complained removing nonexistent dir %s", mytmpdir)
		}
	})
}
