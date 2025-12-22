package stave

import (
	"context"
	"path/filepath"

	"github.com/yaklabco/stave/internal/parse"
)

// TargetNames returns a list of all targets in the current directory or stavefiles/ directory.
func TargetNames(ctx context.Context, dir string) ([]string, error) {
	params := RunParams{
		Dir: dir,
	}
	preprocessRunParams(&params)

	files, err := Stavefiles(params.Dir, params.GOOS, params.GOARCH, params.UsesStavefiles())
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, nil
	}

	filenames := make([]string, 0, len(files))
	for i := range files {
		filenames = append(filenames, filepath.Base(files[i]))
	}

	info, err := parse.PrimaryPackage(ctx, params.GoCmd, params.Dir, filenames)
	if err != nil {
		return nil, err
	}

	targets := make([]string, 0, len(info.Funcs)+len(info.Aliases))
	for _, f := range info.Funcs {
		targets = append(targets, lowerFirstTargetName(f.TargetName()))
	}
	for alias := range info.Aliases {
		targets = append(targets, lowerFirstTargetName(alias))
	}

	return targets, nil
}
