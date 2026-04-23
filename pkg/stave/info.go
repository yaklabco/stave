package stave

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"

	"github.com/yaklabco/stave/internal/parse"
)

// runInfoMode handles the -i/--info flag by parsing stavefiles and rendering
// the doc string directly, without compiling a temporary binary.
func runInfoMode(ctx context.Context, params RunParams) error {
	if len(params.Args) < 1 {
		return errors.New("no target specified for -i/--info flag")
	}

	files, err := Stavefiles(params.Dir, params.GOOS, params.GOARCH, params.UsesStavefiles())
	if err != nil {
		return fmt.Errorf("determining list of stavefiles: %w", err)
	}

	if len(files) == 0 {
		return errors.New("no .go files marked with the stave build tag in this directory")
	}

	fnames := make([]string, 0, len(files))
	for _, f := range files {
		fnames = append(fnames, filepath.Base(f))
	}

	info, err := parse.PrimaryPackage(ctx, params.GoCmd, params.Dir, fnames, params.Multiline)
	if err != nil {
		return fmt.Errorf("parsing stavefiles: %w", err)
	}

	sort.Sort(info.Funcs)
	sort.Sort(info.Imports)

	data := buildTemplateData(generateBinaryName(params), info)

	return renderTargetInfo(
		params.Stdout,
		params.Args[0],
		data,
	)
}

// renderTargetInfo renders the output of `stave -i`.
//
// It is implemented in the Stave binary (not in the generated mainfile) so it can
// use Charmbracelet styling without requiring additional dependencies in user projects.
func renderTargetInfo(writer io.Writer, targetName string, data *mainfileTemplateData) error {
	allFuncs := make([]*parse.Function, 0, len(data.Funcs))
	allFuncs = append(allFuncs, data.Funcs...)

	for _, imp := range data.Imports {
		allFuncs = append(allFuncs, imp.Info.Funcs...)
	}

	var theTargetFunction *parse.Function
	for _, theFunc := range allFuncs {
		if lowerFirstTargetName(theFunc.TargetName()) == lowerFirstTargetName(targetName) {
			theTargetFunction = theFunc
			break
		}
	}
	if theTargetFunction == nil {
		return fmt.Errorf("target %q not found in parsed functions", targetName)
	}

	var builder strings.Builder
	if theTargetFunction.Comment != "" {
		builder.WriteString(theTargetFunction.Comment)
		builder.WriteString("\n\n")
	}

	fmt.Fprintf(&builder, "Usage:\n\n\t%s %s", data.BinaryName, strings.ToLower(theTargetFunction.TargetName()))
	for _, reqArg := range theTargetFunction.Args {
		fmt.Fprintf(&builder, " <%s>", reqArg.Name)
	}
	builder.WriteString("\n\n")

	aliases := make([]string, 0, len(data.Aliases))
	for alias, target := range data.Aliases {
		if target.Name == theTargetFunction.Name && target.Receiver == theTargetFunction.Receiver {
			aliases = append(aliases, alias)
		}
	}

	if len(aliases) > 0 {
		sort.Strings(aliases)
		fmt.Fprintf(&builder, "Aliases: %s\n\n", strings.Join(aliases, ", "))
	}

	if theTargetFunction.IsWatch {
		builder.WriteString("This is a watch target, which means it will be re-run whenever any of its dependencies change.\n")
	}

	_, err := fmt.Fprint(writer, builder.String())
	if err != nil {
		return fmt.Errorf("writing target info to output: %w", err)
	}

	return nil
}
