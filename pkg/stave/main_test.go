package stave

import (
	"bytes"
	"crypto/sha256"
	"debug/macho"
	"debug/pe"
	"encoding/hex"
	"errors"
	"fmt"
	"go/build"
	"go/parser"
	"go/token"
	"io"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/stave/internal"
	"github.com/yaklabco/stave/pkg/st"
)

const (
	testExeEnv = "STAVE_TEST_STRING"

	hiExclam           = "hi!"
	hiExclamAndNewline = hiExclam + "\n"

	dotExe = ".exe"

	testdataCompiled = "./testdata/compiled"

	targetsBuild = "Targets:\n  build    \n"

	windows = "windows"
	amd64   = "amd64"
)

func TestMain(m *testing.M) {
	if s := os.Getenv(testExeEnv); s != "" {
		_, _ = fmt.Fprint(os.Stdout, s)
		os.Exit(0)
	}
	os.Exit(actualTestMain(m))
}

func actualTestMain(m *testing.M) int {
	// ensure we write our temporary binaries to a directory that we'll delete
	// after running tests.
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		slog.Error(err.Error())
		return 1
	}
	defer func() {
		removeErr := os.RemoveAll(dir)
		if removeErr != nil {
			slog.Error("error removing temp dir: ", slog.Any("error", removeErr))
		}
	}()

	if err := os.Setenv(st.CacheEnv, dir); err != nil {
		slog.Error(err.Error())
		return 1
	}
	if err := os.Unsetenv(st.VerboseEnv); err != nil {
		slog.Error(err.Error())
		return 1
	}
	if err := os.Unsetenv(st.DebugEnv); err != nil {
		slog.Error(err.Error())
		return 1
	}
	if err := os.Unsetenv(st.IgnoreDefaultEnv); err != nil {
		slog.Error(err.Error())
		return 1
	}
	if err := os.Unsetenv(st.EnableColorEnv); err != nil {
		slog.Error(err.Error())
		return 1
	}
	if err := os.Unsetenv(st.TargetColorEnv); err != nil {
		slog.Error(err.Error())
		return 1
	}
	if err := resetTerm(); err != nil {
		slog.Error(err.Error())
		return 1
	}

	return m.Run()
}

func resetTerm() error {
	if term, exists := os.LookupEnv("TERM"); exists {
		log.Printf("Current terminal: %s", term)
		// unset TERM env var in order to disable color output to make the tests simpler
		// there is a specific test for colorized output, so all the other tests can use non-colorized one
		if err := os.Unsetenv("TERM"); err != nil {
			return err
		}
	}

	return os.Setenv(st.EnableColorEnv, "false")
}

func TestTransitiveDepCache(t *testing.T) {
	ctx := t.Context()

	cache, err := internal.OutputDebug(ctx, "go", "env", "GOCACHE")
	require.NoError(t, err)
	if cache == "" {
		t.Skip("skipping gocache tests on go version without cache")
	}
	// Test that if we change a transitive dep, that we recompile
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	runParams := RunParams{
		BaseCtx: ctx,
		Stderr:  stderr,
		Stdout:  stdout,
		Dir:     "testdata/transitiveDeps",
		Args:    []string{"Run"},
	}

	err = Run(runParams)
	require.NoError(t, err, "stderr was: %s", stderr.String())

	expected := "woof\n"
	assert.Equal(t, expected, stdout.String())

	// ok, so baseline, the generated and cached binary should do "woof"
	// now change out the transitive dependency that does the output
	// so that it produces different output.
	require.NoError(t, os.Rename("testdata/transitiveDeps/dep/dog.go", "testdata/transitiveDeps/dep/dog.notgo"))
	defer func() {
		assert.NoError(t, os.Rename("testdata/transitiveDeps/dep/dog.notgo", "testdata/transitiveDeps/dep/dog.go"))
	}()

	require.NoError(t, os.Rename("testdata/transitiveDeps/dep/cat.notgo", "testdata/transitiveDeps/dep/cat.go"))
	defer func() {
		assert.NoError(t, os.Rename("testdata/transitiveDeps/dep/cat.go", "testdata/transitiveDeps/dep/cat.notgo"))
	}()

	stderr.Reset()
	stdout.Reset()

	err = Run(runParams)
	require.NoError(t, err, "stderr was: %s", stderr.String())

	expected = "meow\n"
	assert.Equal(t, expected, stdout.String())
}

func TestTransitiveHashFast(t *testing.T) {
	ctx := t.Context()

	cache, err := internal.OutputDebug(ctx, "go", "env", "GOCACHE")
	require.NoError(t, err)
	if cache == "" {
		t.Skip("skipping hashfast tests on go version without cache")
	}

	// Test that if we change a transitive dep, that we don't recompile.
	// We intentionally run the first time without hashfast to ensure that
	// we recompile the binary with the current code.
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	runParams := RunParams{
		BaseCtx: ctx,
		Stderr:  stderr,
		Stdout:  stdout,
		Dir:     "testdata/transitiveDeps",
		Args:    []string{"Run"},
	}

	err = Run(runParams)
	require.NoError(t, err, "stderr was: %s", stderr.String())

	expected := "woof\n"
	assert.Equal(t, expected, stdout.String())

	// ok, so baseline, the generated and cached binary should do "woof"
	// now change out the transitive dependency that does the output
	// so that it produces different output.
	require.NoError(t, os.Rename("testdata/transitiveDeps/dep/dog.go", "testdata/transitiveDeps/dep/dog.notgo"))
	defer func() {
		assert.NoError(t, os.Rename("testdata/transitiveDeps/dep/dog.notgo", "testdata/transitiveDeps/dep/dog.go"))
	}()

	require.NoError(t, os.Rename("testdata/transitiveDeps/dep/cat.notgo", "testdata/transitiveDeps/dep/cat.go"))
	defer func() {
		assert.NoError(t, os.Rename("testdata/transitiveDeps/dep/cat.go", "testdata/transitiveDeps/dep/cat.notgo"))
	}()

	stderr.Reset()
	stdout.Reset()

	runParams.HashFast = true
	err = Run(runParams)
	require.NoError(t, err, "stderr was: %s", stderr.String())

	// we should still get woof, even though the dependency was changed to
	// return "meow", because we're only hashing the top level stavefiles, not
	// dependencies.
	assert.Equal(t, expected, stdout.String())
}

func TestListStavefilesMain(t *testing.T) {
	buf := &bytes.Buffer{}
	files, err := Stavefiles("testdata/mixed_main_files", "", "", false)
	require.NoError(t, err, buf.String())

	expected := []string{
		filepath.FromSlash("testdata/mixed_main_files/stave_helpers.go"),
		filepath.FromSlash("testdata/mixed_main_files/stavefile.go"),
	}

	assert.Equal(t, expected, files)
}

func TestListStavefilesIgnoresGOOS(t *testing.T) {
	buf := &bytes.Buffer{}
	if runtime.GOOS == windows {
		t.Setenv("GOOS", "linux")
	} else {
		t.Setenv("GOOS", windows)
	}

	files, err := Stavefiles("testdata/goos_stavefiles", "", "", false)
	require.NoError(t, err, buf.String())

	var expected []string
	if runtime.GOOS == windows {
		expected = []string{filepath.FromSlash("testdata/goos_stavefiles/stavefile_windows.go")}
	} else {
		expected = []string{filepath.FromSlash("testdata/goos_stavefiles/stavefile_nonwindows.go")}
	}

	assert.Equal(t, expected, files)
}

func TestListStavefilesIgnoresRespectsGOOSArg(t *testing.T) {
	buf := &bytes.Buffer{}
	var goos string
	if runtime.GOOS == windows {
		goos = "linux"
	} else {
		goos = windows
	}

	// Set GOARCH as amd64 because windows is not on all non-x86 architectures.
	files, err := Stavefiles("testdata/goos_stavefiles", goos, amd64, false)
	require.NoError(t, err, buf.String())

	var expected []string
	if goos == windows {
		expected = []string{filepath.FromSlash("testdata/goos_stavefiles/stavefile_windows.go")}
	} else {
		expected = []string{filepath.FromSlash("testdata/goos_stavefiles/stavefile_nonwindows.go")}
	}

	assert.Equal(t, expected, files)
}

func TestCompileDiffGoosGoarch(t *testing.T) {
	ctx := t.Context()

	target, err := os.MkdirTemp("./testdata", "")
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, os.RemoveAll(target))
	}()

	// intentionally choose an arch and os to build that are not our current one.

	goos := windows
	if runtime.GOOS == windows {
		goos = "darwin"
	}
	goarch := amd64
	if runtime.GOARCH == amd64 {
		goarch = "386"
	}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	runParams := RunParams{
		BaseCtx: ctx,
		Stderr:  stderr,
		Stdout:  stdout,
		Debug:   true,
		Dir:     "testdata",
		// this is relative to the Dir above
		CompileOut: filepath.Join(".", filepath.Base(target), "output"),
		GOOS:       goos,
		GOARCH:     goarch,
	}

	err = Run(runParams)
	require.NoError(t, err, "stderr was: %s", stderr.String())

	theOS, theArch, err := fileData(filepath.Join(target, "output"))
	require.NoError(t, err, "stderr was: %s", stderr.String())
	if goos == windows {
		assert.Equal(t, winExe, theOS)
	} else {
		assert.Equal(t, macExe, theOS)
	}
	if goarch == amd64 {
		assert.Equal(t, arch64, theArch)
	} else {
		assert.Equal(t, arch32, theArch)
	}
}

func TestListStavefilesLib(t *testing.T) {
	buf := &bytes.Buffer{}
	files, err := Stavefiles("testdata/mixed_lib_files", "", "", false)
	require.NoError(t, err, buf.String())

	expected := []string{
		filepath.FromSlash("testdata/mixed_lib_files/stave_helpers.go"),
		filepath.FromSlash("testdata/mixed_lib_files/stavefile.go"),
	}
	assert.Equal(t, expected, files)
}

func TestMixedStaveImports(t *testing.T) {
	ctx := t.Context()

	require.NoError(t, resetTerm())

	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}

	runParams := RunParams{
		BaseCtx: ctx,
		Dir:     "./testdata/mixed_lib_files",
		Stdout:  stdout,
		Stderr:  stderr,
		List:    true,
	}

	err := Run(runParams)
	require.NoError(t, err, "stderr was: %s", stderr.String())
	expected := targetsBuild
	assert.Equal(t, expected, stdout.String())
}

func TestStavefilesFolder(t *testing.T) {
	ctx := t.Context()

	require.NoError(t, resetTerm())

	wd, err := os.Getwd()
	t.Log(wd)
	require.NoError(t, err)

	require.NoError(t, os.Chdir("testdata/with_stavefiles_folder"))
	// restore previous state
	defer func() {
		assert.NoError(t, os.Chdir(wd))
	}()

	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}

	runParams := RunParams{
		BaseCtx: ctx,
		Dir:     "",
		Stdout:  stdout,
		Stderr:  stderr,
		List:    true,
	}

	err = Run(runParams)
	require.NoError(t, err, "stderr was: %s", stderr.String())

	expected := targetsBuild
	assert.Equal(t, expected, stdout.String())
}

func TestStavefilesFolderMixedWithStavefiles(t *testing.T) {
	ctx := t.Context()

	require.NoError(t, resetTerm())
	wd, err := os.Getwd()
	t.Log(wd)
	require.NoError(t, err)

	require.NoError(t, os.Chdir("testdata/with_stavefiles_folder_and_stave_files_in_dot"))
	// restore previous state
	defer func() {
		assert.NoError(t, os.Chdir(wd))
	}()

	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}

	runParams := RunParams{
		BaseCtx: ctx,
		Dir:     "",
		Stdout:  stdout,
		Stderr:  stderr,
		List:    true,
	}

	err = Run(runParams)
	require.NoError(t, err, "stderr was: %s", stderr.String())

	expected := targetsBuild
	assert.Equal(t, expected, stdout.String())

	expectedErrStr := "[WARNING] You have both a stavefiles directory and stave files in the current directory, in future versions the files will be ignored in favor of the directory\n" //nolint:lll // Long string-literal.
	assert.Equal(t, expectedErrStr, stderr.String())
}

func TestUntaggedStavefilesFolder(t *testing.T) {
	ctx := t.Context()

	require.NoError(t, resetTerm())

	wd, err := os.Getwd()
	t.Log(wd)
	require.NoError(t, err)

	require.NoError(t, os.Chdir("testdata/with_untagged_stavefiles_folder"))
	// restore previous state
	defer func() {
		assert.NoError(t, os.Chdir(wd))
	}()

	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}

	runParams := RunParams{
		BaseCtx: ctx,
		Dir:     "",
		Stdout:  stdout,
		Stderr:  stderr,
		List:    true,
	}

	err = Run(runParams)
	require.NoError(t, err, "stderr was: %s", stderr.String())

	expected := targetsBuild
	assert.Equal(t, expected, stdout.String())
}

func TestMixedTaggingStavefilesFolder(t *testing.T) {
	ctx := t.Context()

	require.NoError(t, resetTerm())

	wd, err := os.Getwd()
	t.Log(wd)
	require.NoError(t, err)

	require.NoError(t, os.Chdir("testdata/with_mixtagged_stavefiles_folder"))
	// restore previous state
	defer func() {
		assert.NoError(t, os.Chdir(wd))
	}()

	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}

	runParams := RunParams{
		BaseCtx: ctx,
		Dir:     "",
		Stdout:  stdout,
		Stderr:  stderr,
		List:    true,
	}

	err = Run(runParams)
	require.NoError(t, err, "stderr was: %s", stderr.String())

	expected := "Targets:\n  build            \n  untaggedBuild    \n"
	assert.Equal(t, expected, stdout.String())
}

func TestSetDirWithStavefilesFolder(t *testing.T) {
	ctx := t.Context()

	require.NoError(t, resetTerm())

	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}

	runParams := RunParams{
		BaseCtx: ctx,
		Dir:     "testdata/setdir_with_stavefiles_folder",
		Stdout:  stdout,
		Stderr:  stderr,
		List:    true,
	}

	err := Run(runParams)
	require.NoError(t, err, "stderr was: %s", stderr.String())
	expected := targetsBuild
	assert.Equal(t, expected, stdout.String())
}

func TestGoRun(t *testing.T) {
	c := exec.Command("go", "run", "main.go")
	c.Dir = "./testdata"
	c.Env = os.Environ()
	b, err := c.CombinedOutput()
	require.NoError(t, err, "stderr was: %s", string(b))

	expected := "stuff\n"
	assert.Equal(t, expected, string(b))
}

func TestVerbose(t *testing.T) {
	ctx := t.Context()

	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}

	runParams := RunParams{
		BaseCtx: ctx,
		Dir:     "./testdata",
		Stdout:  stdout,
		Stderr:  stderr,
		Args:    []string{"testverbose"},
	}

	err := Run(runParams)
	require.NoError(t, err, "stderr was: %s", stderr.String())
	expected := ""
	assert.Equal(t, expected, stdout.String())

	stderr.Reset()
	stdout.Reset()
	runParams.Verbose = true
	err = Run(runParams)
	require.NoError(t, err, "stderr was: %s", stderr.String())

	expected = "Running target: TestVerbose\nhi!\n"
	assert.Equal(t, expected, stderr.String())
}

func TestList(t *testing.T) {
	ctx := t.Context()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	runParams := RunParams{
		BaseCtx: ctx,
		Dir:     "./testdata/list",
		Stdout:  stdout,
		Stderr:  stderr,
		List:    true,
	}

	err := Run(runParams)
	require.NoError(t, err, "stderr was: %s", stderr.String())
	expected := `
This is a comment on the package which should get turned into output with the list of targets.

Targets:
  somePig*       This is the synopsis for SomePig.
  testVerbose    

* default target
`[1:]

	assert.Equal(t, expected, stdout.String())
}

var terminals = []struct {
	code          string
	supportsColor bool
}{
	{"", true},
	{"vt100", false},
	{"cygwin", false},
	{"xterm-mono", false},
	{"xterm", true},
	{"xterm-vt220", true},
	{"xterm-16color", true},
	{"xterm-256color", true},
	{"screen-256color", true},
}

func TestListWithColor(t *testing.T) {
	t.Setenv(st.EnableColorEnv, "true")
	t.Setenv(st.TargetColorEnv, st.Cyan.String())

	expectedPlainText := `
This is a comment on the package which should get turned into output with the list of targets.

Targets:
  somePig*       This is the synopsis for SomePig.
  testVerbose    

* default target
`[1:]

	// NOTE: using the literal string would be complicated because I would need to break it
	// in the middle and join with a normal string for the target names,
	// otherwise the single backslash would be taken literally and encoded as \\
	expectedColorizedText := "" +
		"This is a comment on the package which should get turned into output with the list of targets.\n" +
		"\n" +
		"Targets:\n" +
		"  \x1b[36msomePig*\x1b[0m       This is the synopsis for SomePig.\n" +
		"  \x1b[36mtestVerbose\x1b[0m    \n" +
		"\n" +
		"* default target\n"

	for _, terminal := range terminals {
		t.Run(terminal.code, func(t *testing.T) {
			ctx := t.Context()

			t.Setenv("TERM", terminal.code)

			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			runParams := RunParams{
				BaseCtx: ctx,
				Dir:     "./testdata/list",
				Stdout:  stdout,
				Stderr:  stderr,
				List:    true,
			}

			err := Run(runParams)
			require.NoError(t, err, "stderr was: %s", stderr.String())
			var expected string
			if terminal.supportsColor {
				expected = expectedColorizedText
			} else {
				expected = expectedPlainText
			}

			assert.Equal(t, expected, stdout.String())
		})
	}
}

func TestNoArgNoDefaultList(t *testing.T) {
	ctx := t.Context()

	require.NoError(t, resetTerm())
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	runParams := RunParams{
		BaseCtx: ctx,
		Dir:     "testdata/no_default",
		Stdout:  stdout,
		Stderr:  stderr,
	}

	err := Run(runParams)
	require.NoError(t, err, "stderr was: %s", stderr.String())
	assert.Empty(t, stderr.String())

	expected := `
Targets:
  bazBuz    Prints out 'BazBuz'.
  fooBar    Prints out 'FooBar'.
`[1:]

	assert.Equal(t, expected, stdout.String())
}

func TestIgnoreDefault(t *testing.T) {
	ctx := t.Context()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	runParams := RunParams{
		BaseCtx: ctx,
		Dir:     "./testdata/list",
		Stdout:  stdout,
		Stderr:  stderr,
	}
	t.Setenv(st.IgnoreDefaultEnv, "1")
	require.NoError(t, resetTerm())

	err := Run(runParams)
	require.NoError(t, err, "stderr was: %s", stderr.String())
	expected := `
This is a comment on the package which should get turned into output with the list of targets.

Targets:
  somePig*       This is the synopsis for SomePig.
  testVerbose    

* default target
`[1:]

	assert.Equal(t, expected, stdout.String())
}

func TestTargetError(t *testing.T) {
	ctx := t.Context()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	runParams := RunParams{
		BaseCtx: ctx,
		Dir:     "./testdata",
		Stdout:  stdout,
		Stderr:  stderr,
		Args:    []string{"returnsnonnilerror"},
	}

	err := Run(runParams)
	require.Error(t, err)

	expected := "Error: bang!\n"
	assert.Equal(t, expected, stderr.String())
}

func TestStdinCopy(t *testing.T) {
	ctx := t.Context()

	stdin := strings.NewReader(hiExclam)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	runParams := RunParams{
		BaseCtx: ctx,
		Dir:     "./testdata",
		Stdin:   stdin,
		Stdout:  stdout,
		Stderr:  stderr,
		Args:    []string{"CopyStdin"},
	}

	err := Run(runParams)
	require.NoError(t, err, "stderr was: %s", stderr.String())
	expected := hiExclam
	assert.Equal(t, expected, stdout.String())
}

func TestTargetPanics(t *testing.T) {
	ctx := t.Context()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	runParams := RunParams{
		BaseCtx: ctx,
		Dir:     "./testdata",
		Stdout:  stdout,
		Stderr:  stderr,
		Args:    []string{"panics"},
	}

	err := Run(runParams)
	require.Error(t, err)

	expected := "Error: boom!\n"
	assert.Equal(t, expected, stderr.String())
}

func TestPanicsErr(t *testing.T) {
	ctx := t.Context()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	runParams := RunParams{
		BaseCtx: ctx,
		Dir:     "./testdata",
		Stdout:  stdout,
		Stderr:  stderr,
		Args:    []string{"panicserr"},
	}

	err := Run(runParams)
	require.Error(t, err)

	expected := "Error: kaboom!\n"
	assert.Equal(t, expected, stderr.String())
}

// ensure we include the hash of the mainfile template in determining the
// executable name to run, so we automatically create a new exe if the template
// changes.
func TestHashTemplate(t *testing.T) {
	ctx := t.Context()

	templ := staveMainfileTplString
	defer func() { staveMainfileTplString = templ }()
	name, err := ExeName(ctx, "go", st.CacheDir(), []string{"testdata/func.go", "testdata/command.go"})
	require.NoError(t, err)

	staveMainfileTplString = "some other template"
	changed, err := ExeName(ctx, "go", st.CacheDir(), []string{"testdata/func.go", "testdata/command.go"})
	require.NoError(t, err)

	assert.NotEqual(t, name, changed)
}

// Test if the -keep flag does keep the mainfile around after running.
func TestKeepFlag(t *testing.T) {
	ctx := t.Context()

	buildFile := "./testdata/keep_flag/" + mainfile
	_ = os.Remove(buildFile)
	defer func() {
		assert.NoError(t, os.Remove(buildFile))
	}()

	logWriter := tLogWriter{t}

	runParams := RunParams{
		BaseCtx: ctx,
		Dir:     "./testdata/keep_flag",
		Stdout:  logWriter,
		Stderr:  logWriter,
		List:    true,
		Keep:    true,
		Force:   true, // need force so we always regenerate
	}

	err := Run(runParams)
	require.NoError(t, err)
	_, err = os.Stat(buildFile)
	require.NoError(t, err)
}

type tLogWriter struct {
	*testing.T
}

func (t tLogWriter) Write(b []byte) (int, error) {
	t.Log(string(b))
	return len(b), nil
}

// Test if generated mainfile references anything other than the stdlib.
func TestOnlyStdLib(t *testing.T) {
	ctx := t.Context()

	buildFile := "./testdata/onlyStdLib/" + mainfile
	_ = os.Remove(buildFile)
	defer func() {
		assert.NoError(t, os.Remove(buildFile))
	}()

	logWriter := tLogWriter{t}

	runParams := RunParams{
		BaseCtx: ctx,
		Dir:     "./testdata/onlyStdLib",
		Stdout:  logWriter,
		Stderr:  logWriter,
		List:    true,
		Keep:    true,
		Force:   true, // need force so we always regenerate
		Verbose: true,
	}

	err := Run(runParams)
	require.NoError(t, err)
	_, err = os.Stat(buildFile)
	require.NoError(t, err)

	fset := &token.FileSet{}
	// Parse src but stop after processing the imports.
	fd, err := parser.ParseFile(fset, buildFile, nil, parser.ImportsOnly)
	require.NoError(t, err)

	// Print the imports from the file's AST.
	for _, importSpec := range fd.Imports {
		// the path value comes in as a quoted string, i.e. literally \"context\"
		path := strings.Trim(importSpec.Path.Value, "\"")
		pkg, err := build.Default.Import(path, "./testdata/keep_flag", build.FindOnly)
		require.NoError(t, err)

		// Check if pkg.Dir is under GOROOT using filepath.Rel instead of deprecated filepath.HasPrefix
		rel, err := filepath.Rel(build.Default.GOROOT, pkg.Dir)
		require.NoError(t, err)
		assert.False(t, strings.HasPrefix(rel, ".."))
	}
}

func TestMultipleTargets(t *testing.T) {
	ctx := t.Context()

	var stderr, stdout bytes.Buffer
	runParams := RunParams{
		BaseCtx: ctx,
		Dir:     "./testdata",
		Stdout:  &stdout,
		Stderr:  &stderr,
		Args:    []string{"TestVerbose", "ReturnsNilError"},
		Verbose: true,
	}

	err := Run(runParams)
	require.NoError(t, err, "stderr was: %s", stderr.String())
	expectedErrStr := "Running target: TestVerbose\nhi!\nRunning target: ReturnsNilError\n"
	assert.Equal(t, expectedErrStr, stderr.String())

	expectedOutStr := "stuff\n"
	assert.Equal(t, expectedOutStr, stdout.String())
}

func TestFirstTargetFails(t *testing.T) {
	ctx := t.Context()

	var stderr, stdout bytes.Buffer
	runParams := RunParams{
		BaseCtx: ctx,
		Dir:     "./testdata",
		Stdout:  &stdout,
		Stderr:  &stderr,
		Args:    []string{"ReturnsNonNilError", "ReturnsNilError"},
		Verbose: true,
	}

	err := Run(runParams)
	require.Error(t, err)

	expectedErrStr := "Running target: ReturnsNonNilError\nError: bang!\n"
	assert.Equal(t, expectedErrStr, stderr.String())
	assert.Empty(t, stdout.String())
}

func TestBadSecondTargets(t *testing.T) {
	ctx := t.Context()

	var stderr, stdout bytes.Buffer
	runParams := RunParams{
		BaseCtx: ctx,
		Dir:     "./testdata",
		Stdout:  &stdout,
		Stderr:  &stderr,
		Args:    []string{"TestVerbose", "NotGonnaWork"},
	}

	err := Run(runParams)
	require.Error(t, err)

	expectedErrStr := "Unknown target specified: \"NotGonnaWork\"\n"
	assert.Equal(t, expectedErrStr, stderr.String())
	assert.Empty(t, stdout.String())
}

func TestSetDir(t *testing.T) {
	ctx := t.Context()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	err := Run(RunParams{
		BaseCtx: ctx,
		Dir:     "testdata/setdir",
		Stdout:  stdout,
		Stderr:  stderr,
		Args:    []string{"TestCurrentDir"},
	})
	require.NoError(t, err, "stderr was: %s", stderr.String())

	expected := "setdir.go\n"
	assert.Equal(t, expected, stdout.String())
}

func TestSetWorkingDir(t *testing.T) {
	ctx := t.Context()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	err := Run(RunParams{
		BaseCtx: ctx,
		Dir:     "testdata/setworkdir",
		WorkDir: "testdata/setworkdir/data",
		Stdout:  stdout,
		Stderr:  stderr,
		Args:    []string{"TestWorkingDir"},
	})
	require.NoError(t, err, "stderr was: %s", stderr.String())

	expected := "file1.txt, file2.txt\n"
	assert.Equal(t, expected, stdout.String())
}

// Test the timeout option.
func TestTimeout(t *testing.T) {
	ctx := t.Context()

	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}

	runParams := RunParams{
		BaseCtx: ctx,
		Dir:     "testdata/context",
		Stdout:  stdout,
		Stderr:  stderr,
		Args:    []string{"timeout"},
		Timeout: 100 * time.Millisecond,
	}

	err := Run(runParams)
	require.Error(t, err)

	expected := "Error: context deadline exceeded\n"
	assert.Equal(t, expected, stderr.String())
}

func TestInfoTarget(t *testing.T) {
	ctx := t.Context()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	runParams := RunParams{
		BaseCtx: ctx,
		Dir:     "./testdata",
		Stdout:  stdout,
		Stderr:  stderr,
		Args:    []string{"panics"},
		Info:    true,
	}

	err := Run(runParams)
	require.NoError(t, err, "stderr was: %s", stderr.String())
	expected := "Function that panics.\n\nUsage:\n\n\tstave panics\n\n"
	assert.Equal(t, expected, stdout.String())
}

func TestInfoAlias(t *testing.T) {
	ctx := t.Context()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	runParams := RunParams{
		BaseCtx: ctx,
		Dir:     "./testdata/alias",
		Stdout:  stdout,
		Stderr:  stderr,
		Args:    []string{"status"},
		Info:    true,
	}

	err := Run(runParams)
	require.NoError(t, err, "stderr was: %s", stderr.String())
	actual := stdout.String()
	expected := "Prints status.\n\nUsage:\n\n\tstave status\n\nAliases: st, stat\n\n"
	assert.Equal(t, expected, actual)

	runParams = RunParams{
		Dir:    "./testdata/alias",
		Stdout: stdout,
		Stderr: stderr,
		Args:   []string{"checkout"},
		Info:   true,
	}

	stdout.Reset()
	stderr.Reset()
	err = Run(runParams)
	require.NoError(t, err, "stderr was: %s", stderr.String())

	actual = stdout.String()
	expected = "Usage:\n\n\tstave checkout\n\nAliases: co\n\n"
	assert.Equal(t, expected, actual)
}

func TestAlias(t *testing.T) {
	ctx := t.Context()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	debug.SetOutput(stderr)

	runParams := RunParams{
		BaseCtx: ctx,
		Dir:     "testdata/alias",
		Stdout:  stdout,
		Stderr:  stderr,
		Args:    []string{"status"},
		Debug:   true,
	}

	err := Run(runParams)
	require.NoError(t, err, "stderr was: %s", stderr.String())
	expected := "alias!\n"
	assert.Equal(t, expected, stdout.String())

	stdout.Reset()
	stderr.Reset()
	runParams.Args = []string{"st"}
	err = Run(runParams)
	require.NoError(t, err, "stderr was: %s", stderr.String())

	assert.Equal(t, expected, stdout.String())
}

func TestInvalidAlias(t *testing.T) {
	ctx := t.Context()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	log.SetOutput(io.Discard)

	runParams := RunParams{
		BaseCtx: ctx,
		Dir:     "./testdata/invalid_alias",
		Stdout:  stdout,
		Stderr:  stderr,
		Args:    []string{"co"},
	}

	err := Run(runParams)
	require.Error(t, err)

	expected := "Unknown target specified: \"co\"\n"
	assert.Equal(t, expected, stderr.String())
}

func TestRunCompiledPrintsError(t *testing.T) {
	ctx := t.Context()

	stderr := &bytes.Buffer{}
	logger := log.New(stderr, "", 0)
	err := RunCompiled(ctx, RunParams{}, "thiswon'texist", logger)
	require.Error(t, err)
}

func TestCompiledFlags(t *testing.T) {
	ctx := t.Context()

	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	dir := testdataCompiled
	compileDir, err := os.MkdirTemp(dir, "")
	require.NoError(t, err, "stderr was: %s", stderr.String())
	name := filepath.Join(compileDir, "stave_test_out")
	if runtime.GOOS == windows {
		name += dotExe
	}

	// The CompileOut directory is relative to the
	// invocation directory, so chop off the invocation dir.
	outName := "./" + name[len(dir)-1:]
	defer func() {
		assert.NoError(t, os.RemoveAll(compileDir))
	}()

	runParams := RunParams{
		BaseCtx:    ctx,
		Dir:        dir,
		Stdout:     stdout,
		Stderr:     stderr,
		CompileOut: outName,
	}

	err = Run(runParams)
	require.NoError(t, err, "stderr was: %s", stderr.String())

	run := func(stdout, stderr *bytes.Buffer, filename string, args ...string) error {
		stderr.Reset()
		stdout.Reset()
		cmd := exec.Command(filename, args...)
		cmd.Env = os.Environ()
		cmd.Stderr = stderr
		cmd.Stdout = stdout
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("running '%s %s' failed with: %w\nstdout: %s\nstderr: %s",
				filename, strings.Join(args, " "), err, stdout, stderr)
		}
		return nil
	}

	// get info for target with flag -i target
	err = run(stdout, stderr, name, "-i", "deploy")
	require.NoError(t, err, "stderr was: %s", stderr.String())
	want := "This is the synopsis for Deploy. This part shouldn't show up.\n\nUsage:\n\n\t" + filepath.Base(name) + " deploy"
	assert.Equal(t, want, strings.TrimSpace(stdout.String()))

	// run target with verbose flag -v
	err = run(stdout, stderr, name, "-v", "testverbose")
	require.NoError(t, err, "stderr was: %s", stderr.String())
	want = hiExclam
	assert.Contains(t, stderr.String(), want)

	// pass list flag -l
	err = run(stdout, stderr, name, "-l")
	require.NoError(t, err, "stderr was: %s", stderr.String())
	want = "This is the synopsis for Deploy"
	assert.Contains(t, stdout.String(), want)
	want = "This is very verbose"
	assert.Contains(t, stdout.String(), want)

	// pass flag -t 1ms
	err = run(stdout, stderr, name, "-t", "1ms", "sleep")
	require.Error(t, err)
	want = "context deadline exceeded"
	assert.Contains(t, err.Error(), want)
}

func TestCompiledEnvironmentVars(t *testing.T) {
	ctx := t.Context()

	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}

	dir := testdataCompiled
	compileDir, err := os.MkdirTemp(dir, "")
	require.NoError(t, err, "stderr was: %s", stderr.String())

	name := filepath.Join(compileDir, "stave_test_out")
	if runtime.GOOS == windows {
		name += dotExe
	}

	// The CompileOut directory is relative to the
	// invocation directory, so chop off the invocation dir.
	outName := "./" + name[len(dir)-1:]
	defer func() {
		assert.NoError(t, os.RemoveAll(compileDir))
	}()

	runParams := RunParams{
		BaseCtx:    ctx,
		Dir:        dir,
		Stdout:     stdout,
		Stderr:     stderr,
		CompileOut: outName,
	}

	err = Run(runParams)
	require.NoError(t, err, "stderr was: %s", stderr.String())

	run := func(stdout, stderr *bytes.Buffer, filename string, envval string, args ...string) error {
		stderr.Reset()
		stdout.Reset()
		cmd := exec.Command(filename, args...)
		cmd.Env = []string{envval}
		cmd.Stderr = stderr
		cmd.Stdout = stdout
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("running '%s %s' failed with: %w\nstdout: %s\nstderr: %s",
				filename, strings.Join(args, " "), err, stdout, stderr)
		}
		return nil
	}

	err = run(stdout, stderr, name, "STAVEFILE_INFO=1", "deploy")
	require.NoError(t, err, "stderr was: %s", stderr.String())
	want := "This is the synopsis for Deploy. This part shouldn't show up.\n\nUsage:\n\n\t" + filepath.Base(name) + " deploy\n\n"
	assert.Equal(t, want, stdout.String())

	err = run(stdout, stderr, name, st.VerboseEnv+"=1", "testverbose")
	require.NoError(t, err, "stderr was: %s", stderr.String())
	want = hiExclam
	assert.Contains(t, stderr.String(), want)

	err = run(stdout, stderr, name, "STAVEFILE_LIST=1")
	require.NoError(t, err, "stderr was: %s", stderr.String())
	want = "This is the synopsis for Deploy"
	assert.Contains(t, stdout.String(), want)
	want = "This is very verbose"
	assert.Contains(t, stdout.String(), want)

	err = run(stdout, stderr, name, st.IgnoreDefaultEnv+"=1")
	require.NoError(t, err, "stderr was: %s", stderr.String())
	want = "Compiled package description."
	assert.Contains(t, stdout.String(), want)

	err = run(stdout, stderr, name, "STAVEFILE_TIMEOUT=1ms", "sleep")
	require.Error(t, err)
	want = "context deadline exceeded"
	assert.Contains(t, stderr.String(), want)
}

func TestCompiledVerboseFlag(t *testing.T) {
	ctx := t.Context()

	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	dir := testdataCompiled
	compileDir, err := os.MkdirTemp(dir, "")
	require.NoError(t, err, "stderr was: %s", stderr.String())
	filename := filepath.Join(compileDir, "stave_test_out")
	if runtime.GOOS == windows {
		filename += dotExe
	}

	// The CompileOut directory is relative to the
	// invocation directory, so chop off the invocation dir.
	outName := "./" + filename[len(dir)-1:]
	defer func() {
		assert.NoError(t, os.RemoveAll(compileDir))
	}()

	runParams := RunParams{
		BaseCtx:    ctx,
		Dir:        dir,
		Stdout:     stdout,
		Stderr:     stderr,
		CompileOut: outName,
	}

	err = Run(runParams)
	require.NoError(t, err, "stderr was: %s", stderr.String())

	run := func(verboseEnv string, args ...string) string {
		var stdout, stderr bytes.Buffer
		args = append(args, "printverboseflag")
		cmd := exec.Command(filename, args...)
		cmd.Env = []string{verboseEnv}
		cmd.Stderr = &stderr
		cmd.Stdout = &stdout
		err := cmd.Run()
		require.NoError(t, err, "running '%s %s' failed with: %v\nstdout: %s\nstderr: %s", filename, strings.Join(args, " "), err, stdout.String(), stderr.String())

		return strings.TrimSpace(stdout.String())
	}

	got := run("STAVEFILE_VERBOSE=false")
	want := "st.Verbose()==false"
	assert.Equal(t, want, got)

	got = run("STAVEFILE_VERBOSE=false", "-v")
	want = "st.Verbose()==true"
	assert.Equal(t, want, got)

	got = run("STAVEFILE_VERBOSE=true")
	want = "st.Verbose()==true"
	assert.Equal(t, want, got)

	got = run("STAVEFILE_VERBOSE=true", "-v=false")
	want = "st.Verbose()==false"
	assert.Equal(t, want, got)
}

func TestSignals(t *testing.T) {
	ctx := t.Context()

	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	dir := "./testdata/signals"
	compileDir, err := os.MkdirTemp(dir, "")
	require.NoError(t, err, "stderr was: %s", stderr.String())
	name := filepath.Join(compileDir, "stave_test_out")

	// The CompileOut directory is relative to the
	// invocation directory, so chop off the invocation dir.
	outName := "./" + name[len(dir)-1:]
	defer func() {
		assert.NoError(t, os.RemoveAll(compileDir))
	}()

	runParams := RunParams{
		BaseCtx:    ctx,
		Dir:        dir,
		Stdout:     stdout,
		Stderr:     stderr,
		CompileOut: outName,
	}

	err = Run(runParams)
	require.NoError(t, err, "stderr was: %s", stderr.String())

	run := func(stdout, stderr *bytes.Buffer, filename string, target string, signals ...syscall.Signal) error {
		stderr.Reset()
		stdout.Reset()
		cmd := exec.Command(filename, target)
		cmd.Stderr = stderr
		cmd.Stdout = stdout
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("running '%s %s' failed with: %w\nstdout: %s\nstderr: %s",
				filename, target, err, stdout, stderr)
		}

		pid := cmd.Process.Pid
		go func() {
			time.Sleep(time.Millisecond * 500)
			for _, s := range signals {
				killErr := syscall.Kill(pid, s)
				if killErr != nil {
					t.Errorf("failed to kill process %d with signal %s: %v", pid, s, killErr)
				}
				time.Sleep(time.Millisecond * 50)
			}
		}()

		if err := cmd.Wait(); err != nil {
			return fmt.Errorf("running '%s %s' failed with: %w\nstdout: %s\nstderr: %s",
				filename, target, err, stdout, stderr)
		}

		return nil
	}

	err = run(stdout, stderr, name, "exitsAfterSighup", syscall.SIGHUP)
	require.NoError(t, err, "stderr was: %s", stderr.String())
	want := "received sighup\n"
	assert.Contains(t, stdout.String(), want)

	err = run(stdout, stderr, name, "exitsAfterSigint", syscall.SIGINT)
	require.NoError(t, err, "stderr was: %s", stderr.String())
	want = "exiting...done\n"
	assert.Contains(t, stdout.String(), want)
	want = "cancelling stave targets, waiting up to 5 seconds for cleanup...\n"
	assert.Contains(t, stderr.String(), want)

	err = run(stdout, stderr, name, "exitsAfterCancel", syscall.SIGINT)
	require.NoError(t, err, "stderr was: %s", stderr.String())
	want = "exiting...done\ndeferred cleanup\n"
	assert.Contains(t, stdout.String(), want)
	want = "cancelling stave targets, waiting up to 5 seconds for cleanup...\n"
	assert.Contains(t, stderr.String(), want)

	err = run(stdout, stderr, name, "ignoresSignals", syscall.SIGINT, syscall.SIGINT)
	require.Error(t, err)
	want = "cancelling stave targets, waiting up to 5 seconds for cleanup...\nexiting stave\nError: exit forced\n"
	assert.Contains(t, stderr.String(), want)

	err = run(stdout, stderr, name, "ignoresSignals", syscall.SIGINT)
	require.Error(t, err)
	want = "cancelling stave targets, waiting up to 5 seconds for cleanup...\nError: cleanup timeout exceeded\n"
	assert.Contains(t, stderr.String(), want)
}

func TestCompiledDeterministic(t *testing.T) {
	dir := testdataCompiled
	compileDir, err := os.MkdirTemp(dir, "")
	require.NoError(t, err)

	var exp string
	outFile := filepath.Join(dir, mainfile)

	// compile a couple times to be sure
	for iRun, runLabel := range []string{"one", "two", "three", "four"} {
		t.Run(runLabel, func(t *testing.T) {
			// probably don't run this parallel
			filename := filepath.Join(compileDir, "stave_test_out")
			if runtime.GOOS == windows {
				filename += dotExe
			}

			// The CompileOut directory is relative to the
			// invocation directory, so chop off the invocation dir.
			outName := "./" + filename[len(dir)-1:]
			defer func() {
				assert.NoError(t, os.RemoveAll(compileDir))
			}()
			defer func() {
				assert.NoError(t, os.Remove(outFile))
			}()

			runParams := RunParams{
				Stderr:     os.Stderr,
				Stdout:     os.Stdout,
				Verbose:    true,
				Keep:       true,
				Dir:        dir,
				CompileOut: outName,
			}

			err := Run(runParams)
			require.NoError(t, err)
			fd, err := os.Open(outFile)
			require.NoError(t, err)
			defer func() {
				assert.NoError(t, fd.Close())
			}()

			hasher := sha256.New()
			_, err = io.Copy(hasher, fd)
			require.NoError(t, err)

			got := hex.EncodeToString(hasher.Sum(nil))
			// set exp on first iteration, subsequent iterations prove the compiled file is identical
			if iRun == 0 {
				exp = got
			}

			if iRun > 0 {
				assert.Equal(t, exp, got)
			}
		})
	}
}

func TestGoCmd(t *testing.T) {
	ctx := t.Context()

	textOutput := "TestGoCmd"
	t.Setenv(testExeEnv, textOutput)

	// fake out the compiled file, since the code checks for it.
	fd, err := os.CreateTemp("", "")
	require.NoError(t, err)
	name := fd.Name()
	dir := filepath.Dir(name)
	defer func() {
		assert.NoError(t, os.Remove(name))
	}()
	_ = fd.Close()

	buf := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	if err := Compile(ctx, CompileParams{
		Goos:      "",
		Goarch:    "",
		Ldflags:   "",
		StavePath: dir,
		GoCmd:     os.Args[0],
		CompileTo: name,
		Gofiles:   []string{},
		Debug:     false,
		Stderr:    stderr,
		Stdout:    buf,
	}); err != nil {
		t.Log("stderr: ", stderr.String())
		t.Fatal(err)
	}
	if buf.String() != textOutput {
		t.Fatalf("We didn't run the custom go cmd. Expected output %q, but got %q", textOutput, buf)
	}
}

func TestGoModules(t *testing.T) {
	ctx := t.Context()

	require.NoError(t, resetTerm())
	dir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, os.RemoveAll(dir))
	}()

	// beware, stave builds in go versions older than 1.17 so both build tag formats need to be present
	err = os.WriteFile(filepath.Join(dir, "stavefile.go"), []byte(`//go:build stave
// +build stave

package main

func Test() {
	print("nothing is imported here for >1.17 compatibility")
}
`), 0600)
	require.NoError(t, err)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := exec.CommandContext(ctx, "go", "mod", "init", "app")
	cmd.Dir = dir
	cmd.Env = os.Environ()
	cmd.Stderr = stderr
	cmd.Stdout = stdout
	require.NoError(t, cmd.Run(), "failed to run 'go mod init', stderr was: %s", stderr.String())

	stderr.Reset()
	stdout.Reset()

	// we need to run go mod tidy, since go build will no longer auto-add dependencies.
	cmd = exec.Command("go", "mod", "tidy")
	cmd.Dir = dir
	cmd.Env = os.Environ()
	cmd.Stderr = stderr
	cmd.Stdout = stdout
	require.NoError(t, cmd.Run(), "failed to run 'go mod init', stderr was: %s", stderr.String())

	stderr.Reset()
	stdout.Reset()
	err = Run(RunParams{
		Dir:    dir,
		Stderr: stderr,
		Stdout: stdout,
	})
	require.NoError(t, err, "stderr was: %s", stderr.String())

	expected := `
Targets:
  test    
`[1:]

	assert.Equal(t, expected, stdout.String())
}

func TestNamespaceDep(t *testing.T) {
	ctx := t.Context()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	runParams := RunParams{
		BaseCtx: ctx,
		Dir:     "./testdata/namespaces",
		Stderr:  stderr,
		Stdout:  stdout,
		Args:    []string{"TestNamespaceDep"},
	}

	err := Run(runParams)
	require.NoError(t, err, "stderr was: %s", stderr.String())

	expected := hiExclamAndNewline
	assert.Equal(t, expected, stdout.String())
}

func TestNamespace(t *testing.T) {
	ctx := t.Context()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	runParams := RunParams{
		BaseCtx: ctx,
		Dir:     "./testdata/namespaces",
		Stdout:  stdout,
		Stderr:  stderr,
		Args:    []string{"ns:error"},
	}

	err := Run(runParams)
	require.NoError(t, err, "stderr was: %s", stderr.String())

	expected := hiExclamAndNewline
	assert.Equal(t, expected, stdout.String())
}

func TestNamespaceDefault(t *testing.T) {
	ctx := t.Context()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	runParams := RunParams{
		BaseCtx: ctx,
		Dir:     "./testdata/namespaces",
		Stdout:  stdout,
		Stderr:  stderr,
	}

	err := Run(runParams)
	require.NoError(t, err, "stderr was: %s", stderr.String())

	expected := hiExclamAndNewline
	assert.Equal(t, expected, stdout.String())
}

func TestAliasToImport(_ *testing.T) {
}

func TestWrongDependency(t *testing.T) {
	ctx := t.Context()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	runParams := RunParams{
		BaseCtx: ctx,
		Dir:     "./testdata/wrong_dep",
		Stdout:  stdout,
		Stderr:  stderr,
	}

	err := Run(runParams)
	require.Error(t, err)

	expected := "Error: argument 0 (complex128), is not a supported argument type\n"
	assert.Equal(t, expected, stderr.String())
}

// Regression tests, add tests to ensure we do not regress on known issues.

// TestBug508 is a regression test for: Bug: using Default with imports selects first matching func by name.
func TestBug508(t *testing.T) {
	ctx := t.Context()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	runParams := RunParams{
		BaseCtx: ctx,
		Dir:     "./testdata/bug508",
		Stderr:  stderr,
		Stdout:  stdout,
	}

	err := Run(runParams)
	require.NoError(t, err, "stderr was: %s", stderr.String())

	expected := "test\n"
	assert.Equal(t, expected, stdout.String())
}

// / This code liberally borrowed from https://github.com/rsc/goversion/blob/master/version/exe.go

type (
	exeType  int
	archSize int
)

const (
	winExe exeType = iota
	macExe

	arch32 archSize = iota
	arch64
)

// fileData tells us if the given file is mac or windows and if they're 32bit or
// 64 bit.  Other exe versions are not supported.
func fileData(file string) (exeType, archSize, error) {
	fd, err := os.Open(file)
	if err != nil {
		return -1, -1, err
	}
	defer func() { _ = fd.Close() }()
	data := make([]byte, 16)
	if _, err := io.ReadFull(fd, data); err != nil {
		return -1, -1, err
	}
	if bytes.HasPrefix(data, []byte("MZ")) {
		// hello windows exe!
		e, err := pe.NewFile(fd)
		if err != nil {
			return -1, -1, err
		}
		if e.Machine == pe.IMAGE_FILE_MACHINE_AMD64 {
			return winExe, arch64, nil
		}
		return winExe, arch32, nil
	}

	if bytes.HasPrefix(data, []byte("\xFE\xED\xFA")) || bytes.HasPrefix(data[1:], []byte("\xFA\xED\xFE")) {
		// hello mac exe!
		fe, err := macho.NewFile(fd)
		if err != nil {
			return -1, -1, err
		}
		if fe.Cpu&0x01000000 != 0 {
			return macExe, arch64, nil
		}
		return macExe, arch32, nil
	}
	return -1, -1, errors.New("unrecognized executable format")
}
