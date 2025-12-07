package stave

import (
	"cmp"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"go/build"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"syscall"
	"time"

	cblog "github.com/charmbracelet/log"
	"github.com/samber/lo"
	"github.com/yaklabco/stave/cmd/stave/version"
	"github.com/yaklabco/stave/internal"
	"github.com/yaklabco/stave/internal/dryrun"
	"github.com/yaklabco/stave/internal/env"
	"github.com/yaklabco/stave/internal/log"
	"github.com/yaklabco/stave/internal/parallelism"
	"github.com/yaklabco/stave/internal/parse"
	"github.com/yaklabco/stave/pkg/sh"
	"github.com/yaklabco/stave/pkg/st"
	"github.com/yaklabco/stave/pkg/stave/prettylog"
)

const (
	// StavefilesDirName is the name of the default folder to look for if no directory was specified,
	// if this folder exists it will be assumed stave package lives inside it.
	StavefilesDirName = "stavefiles"

	curDir = "."

	longAgoShift = -time.Hour * 24 * 365 * 10
)

// RunParams contains the args for invoking a run of Stave.
type RunParams struct {
	BaseCtx context.Context // BaseCtx is the base context for the run, often used for cancellation.

	Stdin  io.Reader // reader to read stdin from
	Stdout io.Writer // writer to write stdout messages to
	Stderr io.Writer // writer to write stderr messages to

	WriterForLogger io.Writer // writer for logger to write to

	List   bool // tells the stavefile to print out a list of targets
	Init   bool // create an initial stavefile from template
	Clean  bool // clean out old generated binaries from cache dir
	Exec   bool // tells the stavefile to treat the rest of the command-line as a command to execute
	Hooks  bool // triggers hooks management mode
	Config bool // triggers config management mode

	Debug      bool          // turn on debug messages
	Dir        string        // directory to read stavefiles from
	WorkDir    string        // directory where stavefiles will run
	Force      bool          // forces recreation of the compiled binary
	Verbose    bool          // tells the stavefile to print out log statements
	Info       bool          // tells the stavefile to print out docstring for a specific target
	Keep       bool          // tells stave to keep the generated main file after compiling
	DryRun     bool          // tells stave that all sh.Run* commands should print, but not execute
	Timeout    time.Duration // tells stave to set a timeout to running the targets
	CompileOut string        // tells stave to compile a static binary to this path, but not execute
	GOOS       string        // sets the GOOS when producing a binary with -compileout
	GOARCH     string        // sets the GOARCH when producing a binary with -compileout
	Ldflags    string        // sets the ldflags when producing a binary with -compileout
	Args       []string      // args to pass to the compiled binary
	GoCmd      string        // the go binary command to run
	CacheDir   string        // the directory where we should store compiled binaries
	HashFast   bool          // don't rely on GOCACHE, just hash the stavefiles

	HooksAreRunning bool // indicates whether hooks are currently being executed
}

// UsesStavefiles returns true if we are getting our stave files from a stavefiles directory.
func (i RunParams) UsesStavefiles() bool {
	return filepath.Base(i.Dir) == StavefilesDirName
}

// Run is the entrypoint for running stave.  It exists external to stave's main
// function to allow it to be used from other programs, specifically so you can
// go run a simple file that run's stave's Run.
func Run(params RunParams) error {
	if params.WriterForLogger == nil {
		params.WriterForLogger = params.Stderr
	}
	logHandler := prettylog.SetupPrettyLogger(params.WriterForLogger)

	if params.Debug {
		logHandler.SetLevel(cblog.DebugLevel)
	}
	slog.Debug("logger initialized")

	preprocessRunParams(&params)

	ctx := params.BaseCtx
	err := applyBasicRunParams(params)
	if err != nil {
		return err
	}

	if howManyThingsToDo(params) > 1 {
		return errors.New("only one of --init, --clean, --list, --hooks, --config, or explicit targets may be specified")
	}

	if params.Init {
		if err := generateInit(params.Dir); err != nil {
			return err
		}
		slog.Info("created initial stavefile", slog.String(log.Filename, initFile))

		return nil
	}

	if params.Clean {
		if err := removeContents(params.CacheDir); err != nil {
			return err
		}
		slog.Info("cleaned cache dir", slog.String(log.Path, params.CacheDir))

		return nil
	}

	if params.Exec {
		return execInStave(ctx, params)
	}

	if params.Hooks {
		return runHooksMode(ctx, params)
	}

	if params.Config {
		return runConfigMode(ctx, params)
	}

	return stave(ctx, params)
}

func execInStave(ctx context.Context, params RunParams) error {
	if len(params.Args) < 1 {
		return errors.New("--exec requires a command (and optionally, arguments) to run")
	}

	dryrun.SetPossible(true)

	theCmd := dryrun.Wrap(ctx, params.Args[0], params.Args[1:]...)
	theCmd.Stderr = params.Stderr
	theCmd.Stdout = params.Stdout
	theCmd.Stdin = params.Stdin
	theCmd.Dir = params.Dir
	if params.WorkDir != params.Dir {
		theCmd.Dir = params.WorkDir
	}

	envMap, err := setupEnv(params)
	if err != nil {
		return fmt.Errorf("setting up environment for stavefile: %w", err)
	}
	theCmd.Env = env.ToAssignments(envMap)

	return theCmd.Run()
}

func runHooksMode(ctx context.Context, params RunParams) error {
	exitCode := RunHooksCommand(ctx, params)
	if exitCode != 0 {
		return st.Fatal(exitCode, "hooks command failed")
	}

	return nil
}

func runConfigMode(ctx context.Context, params RunParams) error {
	exitCode := RunConfigCommandContext(ctx, params.Stdout, params.Stderr, params.Args)
	if exitCode != 0 {
		return st.Fatal(exitCode, "config command failed")
	}
	return nil
}

func stave(ctx context.Context, params RunParams) error {
	files, err := Stavefiles(params.Dir, params.GOOS, params.GOARCH, params.UsesStavefiles())
	if err != nil {
		return fmt.Errorf("determining list of stavefiles: %w", err)
	}

	if len(files) == 0 {
		return errors.New("no .go files marked with the stave build tag in this directory")
	}
	slog.Debug("found stavefiles", slog.Any("files", files))
	exePath := params.CompileOut
	if params.CompileOut == "" {
		exePath, err = ExeName(ctx, params.GoCmd, params.CacheDir, files)
		if err != nil {
			return fmt.Errorf("getting exe name: %w", err)
		}
	}
	slog.Debug("executable path determined", slog.String("exePath", exePath))

	useCache := false
	if params.HashFast {
		slog.Debug("user has set STAVEFILE_HASHFAST, so we'll ignore GOCACHE")
	} else {
		theGoCache, err := internal.OutputDebug(ctx, params.GoCmd, "env", "GOCACHE")
		if err != nil {
			return fmt.Errorf("failed to run %s env GOCACHE: %w", params.GoCmd, err)
		}

		// if GOCACHE exists, always rebuild, so we catch transitive
		// dependencies that have changed.
		if theGoCache != "" {
			slog.Debug("go build cache exists, will ignore any compiled binary")
			useCache = true
		}
	}

	if !useCache {
		_, err = os.Stat(exePath)
		switch {
		case err == nil:
			if !params.Force {
				slog.Debug("Running existing executable")
				return RunCompiled(ctx, params, exePath)
			}
			slog.Debug("ignoring existing executable")
		case os.IsNotExist(err):
			slog.Debug("no existing executable, creating new")
		default:
			slog.Debug(
				"error reading existing executable",
				slog.String(log.Path, exePath),
				slog.Any(log.Error, err),
			)
			slog.Debug("creating new executable")
		}
	}

	// parse wants dir + filenames... arg
	fnames := make([]string, 0, len(files))
	for i := range files {
		fnames = append(fnames, filepath.Base(files[i]))
	}

	slog.Debug("parsing stavefiles")
	info, err := parse.PrimaryPackage(ctx, params.GoCmd, params.Dir, fnames)
	if err != nil {
		return fmt.Errorf("parsing stavefiles: %w", err)
	}

	// reproducible output for deterministic builds
	sort.Sort(info.Funcs)
	sort.Sort(info.Imports)

	main := filepath.Join(params.Dir, mainFile)
	binaryName := "stave"
	if params.CompileOut != "" {
		binaryName = filepath.Base(params.CompileOut)
	}

	err = GenerateMainFile(binaryName, main, info)
	if err != nil {
		return err
	}
	if !params.Keep {
		defer func() { _ = os.RemoveAll(main) }()
	}

	files = append(files, main)
	if err := Compile(ctx, CompileParams{
		Goos:      params.GOOS,
		Goarch:    params.GOARCH,
		Ldflags:   params.Ldflags,
		StavePath: params.Dir,
		GoCmd:     params.GoCmd,
		CompileTo: exePath,
		Gofiles:   files,
		Debug:     params.Debug,
		Stderr:    params.Stderr,
		Stdout:    params.Stdout,
	}); err != nil {
		return err
	}
	if !params.Keep {
		// move aside this file before we run the compiled version, in case the
		// compiled file screws things up.  Yes this doubles up with the above
		// defer, that's ok.
		_ = os.RemoveAll(main)
	} else {
		slog.Debug("keeping mainfile")
	}

	if params.CompileOut != "" {
		return nil
	}

	return RunCompiled(ctx, params, exePath)
}

func howManyThingsToDo(params RunParams) int {
	nThingsToDo := 0
	switch {
	case params.List:
		nThingsToDo++
	case params.Init:
		nThingsToDo++
	case params.Clean:
		nThingsToDo++
	case params.Exec:
		nThingsToDo++
	case params.Hooks:
		nThingsToDo++
	case params.Config:
		nThingsToDo++
	case len(params.Args) > 0:
		nThingsToDo++
	}
	return nThingsToDo
}

func preprocessRunParams(params *RunParams) {
	params.BaseCtx = cmp.Or(params.BaseCtx, context.Background())

	params.Stdin = cmp.Or(params.Stdin, io.Reader(os.Stdin))
	params.Stdout = cmp.Or(params.Stdout, io.Writer(os.Stdout))
	params.Stderr = cmp.Or(params.Stderr, io.Writer(os.Stderr))

	params.HashFast = cmp.Or(params.HashFast, st.HashFast())

	params.GoCmd = cmp.Or(params.GoCmd, st.GoCmd())

	params.Dir = cmp.Or(params.Dir, curDir)

	params.WorkDir = cmp.Or(params.WorkDir, params.Dir)

	params.CacheDir = cmp.Or(params.CacheDir, st.CacheDir())

	// . will be default unless we find a stave folder.
	stavefilesDir := filepath.Join(params.Dir, StavefilesDirName)

	stavefilesDirStat, err := os.Stat(stavefilesDir)
	if err != nil || !stavefilesDirStat.IsDir() {
		return
	}

	originalDir := params.Dir
	params.Dir = stavefilesDir // preemptive assignment

	// TODO: Remove this fallback when the bw compatibility is removed.
	files, err := Stavefiles(originalDir, params.GOOS, params.GOARCH, false)
	if err != nil || len(files) == 0 {
		return
	}

	slog.Warn(
		"You have both a stavefiles directory and stave files in the " +
			"current directory, in future versions the files will be ignored in favor of the directory",
	)
	params.Dir = originalDir
}

func applyBasicRunParams(params RunParams) error {
	if params.DryRun {
		dryrun.SetRequested(true)
	}

	if lo.IsEmpty(params.CompileOut) && (params.GOARCH != "" || params.GOOS != "") {
		return errors.New("-goos and -goarch only apply when running with -compile")
	}

	if params.Info && len(params.Args) != 1 {
		return errors.New("-d requires exactly one target, for which to show details")
	}

	return nil
}

type mainfileTemplateData struct {
	Description string
	Funcs       []*parse.Function
	DefaultFunc parse.Function
	Aliases     map[string]*parse.Function
	Imports     []*parse.Import
	BinaryName  string
}

// listGoFiles returns a list of all .go files in a given directory,
// matching the provided tag.
func listGoFiles(stavePath, tag string, envMap map[string]string) ([]string, error) {
	origStavePath := stavePath
	if !filepath.IsAbs(stavePath) {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("can't get current working directory: %w", err)
		}
		stavePath = filepath.Join(cwd, stavePath)
	}

	bctx := build.Default
	bctx.BuildTags = []string{tag}

	if _, ok := envMap["GOOS"]; ok {
		bctx.GOOS = envMap["GOOS"]
	}

	if _, ok := envMap["GOARCH"]; ok {
		bctx.GOARCH = envMap["GOARCH"]
	}

	pkg, err := bctx.Import(".", stavePath, 0)
	if err != nil {
		var noGoError *build.NoGoError
		if errors.As(err, &noGoError) {
			return []string{}, nil
		}

		// Allow multiple packages in the same directory
		var multiplePackageError *build.MultiplePackageError
		if !errors.As(err, &multiplePackageError) {
			return nil, fmt.Errorf("failed to parse go source files: %w", err)
		}
	}

	if pkg == nil {
		return []string{}, errors.New("unexpected nil return-value from bctx.Import")
	}

	goFiles := make([]string, len(pkg.GoFiles))
	for i := range pkg.GoFiles {
		goFiles[i] = filepath.Join(origStavePath, pkg.GoFiles[i])
	}

	slog.Debug(
		"found go files",
		slog.Int("num_files", len(goFiles)),
		slog.String("tag", tag),
		slog.Any("files", goFiles),
	)

	return goFiles, nil
}

// Stavefiles returns the list of stavefiles in dir.
func Stavefiles(stavePath, goos, goarch string, isStavefilesDirectory bool) ([]string, error) {
	start := time.Now()
	defer func() {
		slog.Debug("finished scanning for Stavefiles", slog.Duration(log.Duration, time.Since(start)))
	}()

	envMap := internal.EnvWithGOOS(goos, goarch)

	slog.Debug("getting all files including those with stave tag", slog.String(log.Path, stavePath))
	staveFiles, err := listGoFiles(stavePath, "stave", envMap)
	if err != nil {
		return nil, fmt.Errorf("listing stave files: %w", err)
	}

	if isStavefilesDirectory {
		// For the stavefiles directory, we always use all go files, both with
		// and without the stave tag, as per normal go build tag rules.
		slog.Debug("using all go files in stavefiles directory", slog.String(log.Path, stavePath))
		return staveFiles, nil
	}

	// For folders other than the stavefiles directory, we only consider files
	// that have the stave build tag and ignore those that don't.

	slog.Debug("getting all files without stave tag", slog.String(log.Path, stavePath))
	nonStaveFiles, err := listGoFiles(stavePath, "", envMap)
	if err != nil {
		return nil, fmt.Errorf("listing non-stave files: %w", err)
	}

	// convert non-Stave list to a map of files to exclude.
	exclude := map[string]bool{}
	for _, f := range nonStaveFiles {
		if f != "" {
			slog.Debug("marked file as non-stave", slog.String(log.Path, f))
			exclude[f] = true
		}
	}

	// filter out the non-stave files from the stave files.
	var files []string
	for _, f := range staveFiles {
		if f != "" && !exclude[f] {
			files = append(files, f)
		}
	}

	return files, nil
}

// CompileParams groups parameters for Compile.
type CompileParams struct {
	Goos      string
	Goarch    string
	Ldflags   string
	StavePath string
	GoCmd     string
	CompileTo string
	Gofiles   []string
	Debug     bool
	Stderr    io.Writer
	Stdout    io.Writer
}

// Compile uses the go tool to compile the files into an executable at path.
func Compile(ctx context.Context, params CompileParams) error {
	slog.Debug(
		"compiling",
		slog.String(log.Path, params.CompileTo),
		slog.String("go_cmd", params.GoCmd),
	)
	if params.Debug {
		if err := internal.RunDebug(ctx, params.GoCmd, "version"); err != nil {
			return err
		}
		if err := internal.RunDebug(ctx, params.GoCmd, "env"); err != nil {
			return err
		}
	}

	envMap := internal.EnvWithGOOS(params.Goos, params.Goarch)

	// strip off the path since we're setting the path in the build command
	for i := range params.Gofiles {
		params.Gofiles[i] = filepath.Base(params.Gofiles[i])
	}

	buildArgs := []string{"build", "-o", params.CompileTo}
	if params.Ldflags != "" {
		buildArgs = append(buildArgs, "-ldflags", params.Ldflags)
	}

	args := make([]string, len(buildArgs), len(buildArgs)+len(params.Gofiles))
	copy(args, buildArgs)
	args = append(args, params.Gofiles...)

	slog.Debug("running go", slog.String(log.Cmd, params.GoCmd), slog.Any(log.Args, args))
	theCmd := dryrun.Wrap(ctx, params.GoCmd, args...)
	theCmd.Env = env.ToAssignments(envMap)
	theCmd.Stderr = params.Stderr
	theCmd.Stdout = params.Stdout
	theCmd.Dir = params.StavePath

	start := time.Now()
	err := theCmd.Run()
	slog.Debug("finished compiling", slog.Duration(log.Duration, time.Since(start)))
	if err != nil {
		return errors.New("error compiling stavefiles")
	}

	return nil
}

// GenerateMainFile generates the stave mainfile at path.
func GenerateMainFile(binaryName, path string, info *parse.PkgInfo) error {
	slog.Debug("generating mainfile", slog.String(log.Path, path))

	outputFile, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("error creating generated mainfile: %w", err)
	}
	defer func() { _ = outputFile.Close() }()
	data := mainfileTemplateData{
		Description: info.Description,
		Funcs:       info.Funcs,
		Aliases:     info.Aliases,
		Imports:     info.Imports,
		BinaryName:  binaryName,
	}

	if info.DefaultFunc != nil {
		data.DefaultFunc = *info.DefaultFunc
	}

	slog.Debug("writing new file", slog.String(log.Path, path))
	if err := mainfileTemplate.Execute(outputFile, data); err != nil {
		return fmt.Errorf("can't execute mainfile template: %w", err)
	}
	if err := outputFile.Close(); err != nil {
		return fmt.Errorf("error closing generated mainfile: %w", err)
	}
	// we set an old modtime on the generated mainfile so that the go tool
	// won't think it has changed more recently than the compiled binary.
	longAgo := time.Now().Add(longAgoShift)
	if err := os.Chtimes(path, longAgo, longAgo); err != nil {
		return fmt.Errorf("error setting old modtime on generated mainfile: %w", err)
	}
	return nil
}

// ExeName reports the executable filename that this version of Stave would
// create for the given stavefiles.
func ExeName(ctx context.Context, goCmd, cacheDir string, files []string) (string, error) {
	hashes := make([]string, 0, len(files)+1)
	for _, s := range files {
		h, err := hashFile(s)
		if err != nil {
			return "", err
		}
		hashes = append(hashes, h)
	}
	// hash the mainfile template to ensure if it gets updated, we make a new
	// binary.
	hashes = append(hashes, fmt.Sprintf("%x", sha256.Sum256([]byte(staveMainfileTplString))))
	sort.Strings(hashes)
	ver, err := internal.OutputDebug(ctx, goCmd, "version")
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256([]byte(strings.Join(hashes, "") + version.OverallVersionString(ctx) + ver))
	filename := hex.EncodeToString(hash[:])

	out := filepath.Join(cacheDir, filename)
	if runtime.GOOS == "windows" {
		out += ".exe"
	}
	return out, nil
}

func hashFile(filename string) (string, error) {
	inputFile, err := os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("can't open input file for hashing: %w", err)
	}
	defer func() { _ = inputFile.Close() }()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, inputFile); err != nil {
		return "", fmt.Errorf("can't write data to hash: %w", err)
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func generateInit(dir string) error {
	slog.Debug("generating default stavefile", slog.String(log.Dir, dir))
	outputFile, err := os.Create(filepath.Join(dir, initFile))
	if err != nil {
		return fmt.Errorf("could not create stave template: %w", err)
	}
	defer func() { _ = outputFile.Close() }()

	if err := initOutput.Execute(outputFile, nil); err != nil {
		return fmt.Errorf("can't execute stavefile template: %w", err)
	}

	return nil
}

// RunCompiled runs an already-compiled stave command with the given args,.
func RunCompiled(ctx context.Context, params RunParams, exePath string) error {
	slog.Debug("running binary", slog.String(log.Path, exePath))
	theCmd := dryrun.Wrap(ctx, exePath, params.Args...)
	theCmd.Stderr = params.Stderr
	theCmd.Stdout = params.Stdout
	theCmd.Stdin = params.Stdin
	theCmd.Dir = params.Dir
	if params.WorkDir != params.Dir {
		theCmd.Dir = params.WorkDir
	}

	envMap, err := setupEnv(params)
	if err != nil {
		return fmt.Errorf("setting up environment for stavefile: %w", err)
	}
	theCmd.Env = env.ToAssignments(envMap)

	slog.Debug(
		"running stavefile with stave vars",
		slog.Any("env", lo.PickBy(envMap, func(key string, _ string) bool {
			return strings.HasPrefix(key, "STAVEFILE_")
		})),
	)
	// catch SIGINT to allow stavefile to handle them
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT)
	defer signal.Stop(sigCh)
	err = theCmd.Run()
	if !sh.CmdRan(err) {
		slog.Error("failed to run compiled stavefile", slog.Any(log.Error, err))
	}
	return err
}

func setupEnv(params RunParams) (map[string]string, error) {
	envMap := env.GetMap()

	// We don't want to actually allow dryrun in the outermost invocation of
	// stave, since that will inhibit the very compilation of the stavefile & the
	// use of the resulting binary.
	// But every situation that's within such an execution is one in which dryrun
	// is supported, so we set this environment variable which will be carried
	// over throughout all such situations.
	envMap["STAVEFILE_DRYRUN_POSSIBLE"] = "1"

	if params.Verbose {
		envMap["STAVEFILE_VERBOSE"] = "1"
	}
	if params.List {
		envMap["STAVEFILE_LIST"] = "1"
	}
	if params.Info {
		envMap["STAVEFILE_INFO"] = "1"
	}
	if params.Debug {
		envMap["STAVEFILE_DEBUG"] = "1"
	}
	if params.GoCmd != "" {
		envMap["STAVEFILE_GOCMD"] = params.GoCmd
	}
	if params.Timeout > 0 {
		envMap["STAVEFILE_TIMEOUT"] = params.Timeout.String()
	}
	if params.DryRun {
		envMap["STAVEFILE_DRYRUN"] = "1"
	}

	if params.HooksAreRunning {
		envMap[HooksAreRunningEnv] = "1"
	}

	if err := parallelism.Apply(envMap); err != nil {
		return nil, err
	}

	return envMap, nil
}

// removeContents removes all files but not any subdirectories in the given
// directory.
func removeContents(dir string) error {
	slog.Debug("removing all files in given directory", slog.String(log.Dir, dir))
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		err = os.Remove(filepath.Join(dir, entry.Name()))
		if err != nil {
			return err
		}
	}

	return nil
}
