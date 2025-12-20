package sh

import (
	"github.com/yaklabco/stave/internal/ish"
)

// Rm removes the given file or directory even if non-empty. It will not return
// an error if the target doesn't exist, only if the target cannot be removed.
func Rm(path string) error {
	return ish.Rm(path)
}

// Copy robustly copies the source file to the destination, overwriting the destination if necessary.
func Copy(dst string, src string) error {
	return ish.Copy(dst, src)
}
