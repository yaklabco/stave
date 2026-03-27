package target

import (
	"log/slog"
	"os"
)

// Path first expands environment variables like $FOO or ${FOO}, and then
// reports if any of the sources have been modified more recently than the
// destination. Path does not descend into directories, it literally just checks
// the modtime of each thing you pass to it. If the destination file doesn't
// exist, it always returns true and nil. It's an error if any of the sources
// don't exist.
func Path(dst string, sources ...string) (bool, error) {
	dst = os.ExpandEnv(dst)
	stat, err := os.Stat(dst)
	if os.IsNotExist(err) {
		slog.Debug("target.Path", "dst", dst, "verdict", true, "reason", "destination does not exist")
		return true, nil
	}
	if err != nil {
		return false, err
	}
	destTime := stat.ModTime()
	newer, srcPath, srcTime, err := pathNewer(destTime, sources...)
	if err != nil {
		return false, err
	}
	slog.Debug("target.Path",
		"dst", dst,
		"dst_time", destTime,
		"src", srcPath,
		"src_time", srcTime,
		"verdict", newer,
	)
	return newer, nil
}

// Glob expands each of the globs (file patterns) into individual sources and
// then calls Path on the result, reporting if any of the resulting sources have
// been modified more recently than the destination. Syntax for Glob patterns is
// the same as stdlib's filepath.Glob. Note that Glob does not expand
// environment variables before globbing -- env var expansion happens during
// the call to Path. It is an error for any glob to return an empty result.
func Glob(dst string, globs ...string) (bool, error) {
	dst = os.ExpandEnv(dst)
	stat, err := os.Stat(dst)
	if os.IsNotExist(err) {
		slog.Debug("target.Glob", "dst", dst, "verdict", true, "reason", "destination does not exist")
		return true, nil
	}
	if err != nil {
		return false, err
	}
	destTime := stat.ModTime()
	newer, srcPath, srcTime, err := globNewer(destTime, globs...)
	if err != nil {
		return false, err
	}
	slog.Debug("target.Glob",
		"dst", dst,
		"dst_time", destTime,
		"src", srcPath,
		"src_time", srcTime,
		"verdict", newer,
	)
	return newer, nil
}

// Dir reports whether any of the sources have been modified more recently
// than the destination. If a source or destination is a directory, this
// function returns true if a source has any file (excluding the directory
// itself) that has been modified more recently than the most recently modified
// file in dst (also excluding the directory itself). If the destination
// file doesn't exist, it always returns true and nil. It's an error if any
// of the sources don't exist.
//
// Dir respects the global ignorelist populated by AddIgnorePattern and
// LoadIgnoreFile.
func Dir(dst string, sources ...string) (bool, error) {
	dst = os.ExpandEnv(dst)
	stat, err := os.Stat(dst)
	if os.IsNotExist(err) {
		slog.Debug("target.Dir", "dst", dst, "verdict", true, "reason", "destination does not exist")
		return true, nil
	}
	if err != nil {
		return false, err
	}
	destTime := stat.ModTime()
	if stat.IsDir() {
		var err error
		destTime, err = NewestModTime(dst)
		if err != nil {
			return false, err
		}
	}
	newer, srcPath, srcTime, err := dirNewer(destTime, sources...)
	if err != nil {
		return false, err
	}
	slog.Debug("target.Dir",
		"dst", dst,
		"dst_time", destTime,
		"src", srcPath,
		"src_time", srcTime,
		"verdict", newer,
	)
	return newer, nil
}
