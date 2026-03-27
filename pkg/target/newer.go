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
// is newer than the target. The modification time of the root directories in
// sources are ignored.
//
// DirNewer respects the global ignorelist populated by AddIgnorePattern and
// LoadIgnoreFile.
func DirNewer(target time.Time, sources ...string) (bool, error) {
	newer, _, _, err := dirNewer(target, sources...)
	return newer, err
}

func dirNewer(target time.Time, sources ...string) (bool, string, time.Time, error) {
	var newestSoFar time.Time
	var newestPath string

	for _, source := range sources {
		source = os.ExpandEnv(source)

		// Get absolute path of source to compare properly in walkFn
		absSource, err := filepath.Abs(source)
		if err != nil {
			absSource = source // fallback
		}

		walkFn := func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if isIgnored(path, info.IsDir()) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			// Skip the entry for the directory itself if it's one of our root sources.
			if info.IsDir() {
				absPath, err := filepath.Abs(path)
				if err == nil && absPath == absSource {
					return nil
				}
			}

			modTime := info.ModTime()
			if modTime.After(target) {
				newestSoFar = modTime
				newestPath = path
				return errNewer
			}
			if modTime.After(newestSoFar) {
				newestSoFar = modTime
				newestPath = path
			}
			return nil
		}
		err = filepath.Walk(source, walkFn)
		if err == nil {
			continue
		}
		if errors.Is(err, errNewer) {
			return true, newestPath, newestSoFar, nil
		}
		return false, "", time.Time{}, err
	}
	return false, newestPath, newestSoFar, nil
}

// GlobNewer performs glob expansion on each source and passes the results to
// PathNewer for inspection. It returns the first time PathNewer encounters a
// newer file.
func GlobNewer(target time.Time, sources ...string) (bool, error) {
	newer, _, _, err := globNewer(target, sources...)
	return newer, err
}

func globNewer(target time.Time, sources ...string) (bool, string, time.Time, error) {
	var newestSoFar time.Time
	var newestPath string

	for _, globPattern := range sources {
		files, err := filepath.Glob(globPattern)
		if err != nil {
			return false, "", time.Time{}, err
		}
		if len(files) == 0 {
			return false, "", time.Time{}, fmt.Errorf("glob didn't match any files: %s", globPattern)
		}
		newer, path, modTime, err := pathNewer(target, files...)
		if err != nil {
			return false, "", time.Time{}, err
		}
		if newer {
			return true, path, modTime, nil
		}
		if modTime.After(newestSoFar) {
			newestSoFar = modTime
			newestPath = path
		}
	}
	return false, newestPath, newestSoFar, nil
}

// PathNewer checks whether any of the sources are newer than the target time.
// It stops at the first newer file it encounters. Each source path is passed
// through os.ExpandEnv.
//
// PathNewer respects the global ignorelist populated by AddIgnorePattern and
// LoadIgnoreFile.
func PathNewer(target time.Time, sources ...string) (bool, error) {
	newer, _, _, err := pathNewer(target, sources...)
	return newer, err
}

func pathNewer(target time.Time, sources ...string) (bool, string, time.Time, error) {
	var newestSoFar time.Time
	var newestPath string

	for _, source := range sources {
		source = os.ExpandEnv(source)
		stat, err := os.Stat(source)
		if err != nil {
			return false, "", time.Time{}, err
		}
		if isIgnored(source, stat.IsDir()) {
			continue
		}
		modTime := stat.ModTime()
		if modTime.After(target) {
			return true, source, modTime, nil
		}
		if modTime.After(newestSoFar) {
			newestSoFar = modTime
			newestPath = source
		}
	}
	return false, newestPath, newestSoFar, nil
}

// OldestModTime recurses a list of target filesystem objects and finds the
// oldest ModTime among them. The modification time of the root directories in
// targets are ignored.
//
// OldestModTime respects the global ignorelist populated by AddIgnorePattern and
// LoadIgnoreFile.
func OldestModTime(targets ...string) (time.Time, error) {
	oldestTime := time.Now().Add(futureShift)
	for _, target := range targets {
		// Get absolute path of target to compare properly in walkFn
		absTarget, err := filepath.Abs(target)
		if err != nil {
			absTarget = target // fallback
		}

		walkFn := func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if isIgnored(path, info.IsDir()) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			// Skip the entry for the directory itself if it's one of our root targets.
			if info.IsDir() {
				absPath, err := filepath.Abs(path)
				if err == nil && absPath == absTarget {
					return nil
				}
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
// newest ModTime among them. The modification time of the root directories in
// targets are ignored.
//
// NewestModTime respects the global ignorelist populated by AddIgnorePattern and
// LoadIgnoreFile.
func NewestModTime(targets ...string) (time.Time, error) {
	newestTime := time.Time{}
	for _, target := range targets {
		// Get absolute path of target to compare properly in walkFn
		absTarget, err := filepath.Abs(target)
		if err != nil {
			absTarget = target // fallback
		}

		walkFn := func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if isIgnored(path, info.IsDir()) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			// Skip the entry for the directory itself if it's one of our root targets.
			if info.IsDir() {
				absPath, err := filepath.Abs(path)
				if err == nil && absPath == absTarget {
					return nil
				}
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
