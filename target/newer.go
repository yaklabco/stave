package target

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const futureShift = time.Hour * 24 * 365 * 250

var (
	// errNewer is an ugly sentinel error to cause filepath.Walk to abort
	// as soon as a newer file is encountered.
	errNewer = errors.New("newer item encountered")
)

// DirNewer reports whether any item in sources is newer than the target time.
// Sources are searched recursively and searching stops as soon as any entry
// is newer than the target.
func DirNewer(target time.Time, sources ...string) (bool, error) {
	walkFn := func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.ModTime().After(target) {
			return errNewer
		}
		return nil
	}
	for _, source := range sources {
		source = os.ExpandEnv(source)
		err := filepath.Walk(source, walkFn)
		if err == nil {
			continue
		}
		if errors.Is(err, errNewer) {
			return true, nil
		}
		return false, err
	}
	return false, nil
}

// GlobNewer performs glob expansion on each source and passes the results to
// PathNewer for inspection. It returns the first time PathNewer encounters a
// newer file.
func GlobNewer(target time.Time, sources ...string) (bool, error) {
	for _, globPattern := range sources {
		files, err := filepath.Glob(globPattern)
		if err != nil {
			return false, err
		}
		if len(files) == 0 {
			return false, fmt.Errorf("glob didn't match any files: %s", globPattern)
		}
		newer, err := PathNewer(target, files...)
		if err != nil {
			return false, err
		}
		if newer {
			return true, nil
		}
	}
	return false, nil
}

// PathNewer checks whether any of the sources are newer than the target time.
// It stops at the first newer file it encounters. Each source path is passed
// through os.ExpandEnv.
func PathNewer(target time.Time, sources ...string) (bool, error) {
	for _, source := range sources {
		source = os.ExpandEnv(source)
		stat, err := os.Stat(source)
		if err != nil {
			return false, err
		}
		if stat.ModTime().After(target) {
			return true, nil
		}
	}
	return false, nil
}

// OldestModTime recurses a list of target filesystem objects and finds the
// oldest ModTime among them.
func OldestModTime(targets ...string) (time.Time, error) {
	oldestTime := time.Now().Add(futureShift)
	for _, target := range targets {
		walkFn := func(_ string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			mTime := info.ModTime()
			if mTime.Before(oldestTime) {
				oldestTime = mTime
			}
			return nil
		}
		if err := filepath.Walk(target, walkFn); err != nil {
			return oldestTime, err
		}
	}
	return oldestTime, nil
}

// NewestModTime recurses a list of target filesystem objects and finds the
// newest ModTime among them.
func NewestModTime(targets ...string) (time.Time, error) {
	newestTime := time.Time{}
	for _, target := range targets {
		walkFn := func(_ string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			mTime := info.ModTime()
			if mTime.After(newestTime) {
				newestTime = mTime
			}
			return nil
		}
		if err := filepath.Walk(target, walkFn); err != nil {
			return newestTime, err
		}
	}
	return newestTime, nil
}
