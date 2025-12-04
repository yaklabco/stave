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
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/samber/lo"
	"github.com/yaklabco/stave/cmd/stave/version"
	"github.com/yaklabco/stave/internal"
	"github.com/yaklabco/stave/internal/dryrun"
	"github.com/yaklabco/stave/internal/env"
	"github.com/yaklabco/stave/internal/parallelism"
	"github.com/yaklabco/stave/internal/parse"
	"github.com/yaklabco/stave/pkg/sh"
	"github.com/yaklabco/stave/pkg/st"
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

	Init  bool // create an initial stavefile from template
	Clean bool // clean out old generated binaries from cache dir

	Debug      bool          // turn on debug messages
	Dir        string        // directory to read stavefiles from
	WorkDir    string        // directory where stavefiles will run
	Force      bool          // forces recreation of the compiled binary
	Verbose    bool          // tells the stavefile to print out log statements
	List       bool          // tells the stavefile to print out a list of targets
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
}

// UsesStavefiles returns true if we are getting our stave files from a stavefiles directory.
func (i RunParams) UsesStavefiles() bool {
	return filepath.Base(i.Dir) == StavefilesDirName
}

// Run is the entrypoint for running stave.  It exists external to stave's main
// function to allow it to be used from other programs, specifically so you can
// go run a simple file that run's stave's Run.
func Run(params RunParams) error {
	preprocessRunParams(&params)

	ctx := params.BaseCtx
	err := applyBasicRunParams(params)
	if err != nil {
		return err
	}

	out := log.New(params.Stdout, "", 0)

	if howManyThingsToDo(params) > 1 {
		return errors.New("only one of -init, -clean, -list, or explicit targets may be specified")
	}

	if params.Init {
		if err := generateInit(params.Dir); err != nil {
			return err
		}
		out.Println(initFile, "created")

		return nil
	}

	if params.Clean {
		if err := removeContents(params.CacheDir); err != nil {
			return err
		}
		out.Println(params.CacheDir, "cleaned")

		return nil
	}

	return stave(ctx, params)
}

func stave(ctx context.Context, params RunParams) error {
	files, err := Stavefiles(params.Dir, params.GOOS, params.GOARCH, params.UsesStavefiles())
	if err != nil {
		return fmt.Errorf("determining list of stavefiles: %w", err)
	}

	if len(files) == 0 {
		return errors.New("no .go files marked with the stave build tag in this directory")
	}
	debug.Printf("found stavefiles: %s", strings.Join(files, ", "))
	exePath := params.CompileOut
	if params.CompileOut == "" {
		exePath, err = ExeName(ctx, params.GoCmd, params.CacheDir, files)
		if err != nil {
			return fmt.Errorf("getting exe name: %w", err)
		}
	}
	debug.Println("output exe is ", exePath)

	useCache := false
	if params.HashFast {
		debug.Println("user has set STAVEFILE_HASHFAST, so we'll ignore GOCACHE")
	} else {
		theGoCache, err := internal.OutputDebug(ctx, params.GoCmd, "env", "GOCACHE")
		if err != nil {
			return fmt.Errorf("failed to run %s env GOCACHE: %w", params.GoCmd, err)
		}

		// if GOCACHE exists, always rebuild, so we catch transitive
		// dependencies that have changed.
		if theGoCache != "" {
			debug.Println("go build cache exists, will ignore any compiled binary")
			useCache = true
		}
	}

	errlog := log.New(params.Stderr, "", 0)

	if !useCache {
		_, err = os.Stat(exePath)
		switch {
		case err == nil:
			if !params.Force {
				debug.Println("Running existing exe")
				return RunCompiled(ctx, params, exePath, errlog)
			}
			debug.Println("ignoring existing executable")
		case os.IsNotExist(err):
			debug.Println("no existing exe, creating new")
		default:
			debug.Printf("error reading existing exe at %v: %v", exePath, err)
			debug.Println("creating new exe")
		}
	}

	// parse wants dir + filenames... arg
	fnames := make([]string, 0, len(files))
	for i := range files {
		fnames = append(fnames, filepath.Base(files[i]))
	}
	if params.Debug {
		parse.EnableDebug()
	}
	debug.Println("parsing files")
	info, err := parse.PrimaryPackage(ctx, params.GoCmd, params.Dir, fnames)
	if err != nil {
		return fmt.Errorf("parsing stavefiles: %w", err)
	}

	// reproducible output for deterministic builds
	sort.Sort(info.Funcs)
	sort.Sort(info.Imports)

	main := filepath.Join(params.Dir, mainfile)
	binaryName := "stave"
	if params.CompileOut != "" {
		binaryName = filepath.Base(params.CompileOut)
	}

	err = GenerateMainfile(binaryName, main, info)
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
		debug.Print("keeping mainfile")
	}

	if params.CompileOut != "" {
		return nil
	}

	return RunCompiled(ctx, params, exePath, errlog)
}

func howManyThingsToDo(params RunParams) int {
	nThingsToDo := 0
	switch {
	case params.Init:
		nThingsToDo++
	case params.Clean:
		nThingsToDo++
	case params.List:
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

	errlog := log.New(params.Stderr, "", 0)

	stavefilesDirStat, err := os.Stat(stavefilesDir)
	if err == nil {
		if stavefilesDirStat.IsDir() {
			originalDir := params.Dir
			params.Dir = stavefilesDir // preemptive assignment
			// TODO: Remove this fallback and the above Stavefiles invocation when the bw compatibility is removed.
			files, err := Stavefiles(originalDir, params.GOOS, params.GOARCH, false)
			if err == nil {
				if len(files) != 0 {
					errlog.Println("[WARNING] You have both a stavefiles directory and stave files in the " +
						"current directory, in future versions the files will be ignored in favor of the directory")
					params.Dir = originalDir
				}
			}
		}
	}

	params.CacheDir = st.CacheDir()
}

func applyBasicRunParams(params RunParams) error {
	if params.Debug {
		debug.SetOutput(params.Stderr)
	}

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

	debug.Printf("found %d go files with build tag %s (files: %v)", len(goFiles), tag, goFiles)
	return goFiles, nil
}

// Stavefiles returns the list of stavefiles in dir.
func Stavefiles(stavePath, goos, goarch string, isStavefilesDirectory bool) ([]string, error) {
	start := time.Now()
	defer func() {
		debug.Println("time to scan for Stavefiles:", time.Since(start))
	}()

	envMap := internal.EnvWithGOOS(goos, goarch)

	debug.Println("getting all files including those with stave tag in", stavePath)
	staveFiles, err := listGoFiles(stavePath, "stave", envMap)
	if err != nil {
		return nil, fmt.Errorf("listing stave files: %w", err)
	}

	if isStavefilesDirectory {
		// For the stavefiles directory, we always use all go files, both with
		// and without the stave tag, as per normal go build tag rules.
		debug.Println("using all go files in stavefiles directory", stavePath)
		return staveFiles, nil
	}

	// For folders other than the stavefiles directory, we only consider files
	// that have the stave build tag and ignore those that don't.

	debug.Println("getting all files without stave tag in", stavePath)
	nonStaveFiles, err := listGoFiles(stavePath, "", envMap)
	if err != nil {
		return nil, fmt.Errorf("listing non-stave files: %w", err)
	}

	// convert non-Stave list to a map of files to exclude.
	exclude := map[string]bool{}
	for _, f := range nonStaveFiles {
		if f != "" {
			debug.Printf("marked file as non-stave: %q", f)
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
	debug.Println("compiling to", params.CompileTo)
	debug.Println("compiling using gocmd:", params.GoCmd)
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

	debug.Printf("running %s %s", params.GoCmd, strings.Join(args, " "))
	theCmd := dryrun.Wrap(ctx, params.GoCmd, args...)
	theCmd.Env = env.ToAssignments(envMap)
	theCmd.Stderr = params.Stderr
	theCmd.Stdout = params.Stdout
	theCmd.Dir = params.StavePath

	start := time.Now()
	err := theCmd.Run()
	debug.Println("time to compile Stavefile:", time.Since(start))
	if err != nil {
		return errors.New("error compiling stavefiles")
	}

	return nil
}

// GenerateMainfile generates the stave mainfile at path.
func GenerateMainfile(binaryName, path string, info *parse.PkgInfo) error {
	debug.Println("Creating mainfile at", path)

	fd, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("error creating generated mainfile: %w", err)
	}
	defer func() { _ = fd.Close() }()
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

	debug.Println("writing new file at", path)
	if err := mainfileTemplate.Execute(fd, data); err != nil {
		return fmt.Errorf("can't execute mainfile template: %w", err)
	}
	if err := fd.Close(); err != nil {
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

func hashFile(fn string) (string, error) {
	fd, err := os.Open(fn)
	if err != nil {
		return "", fmt.Errorf("can't open input file for hashing: %#w", err)
	}
	defer func() { _ = fd.Close() }()

	h := sha256.New()
	if _, err := io.Copy(h, fd); err != nil {
		return "", fmt.Errorf("can't write data to hash: %w", err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func generateInit(dir string) error {
	debug.Println("generating default stavefile in", dir)
	fd, err := os.Create(filepath.Join(dir, initFile))
	if err != nil {
		return fmt.Errorf("could not create stave template: %w", err)
	}
	defer func() { _ = fd.Close() }()

	if err := initOutput.Execute(fd, nil); err != nil {
		return fmt.Errorf("can't execute stavefile template: %w", err)
	}

	return nil
}

// RunCompiled runs an already-compiled stave command with the given args,.
func RunCompiled(ctx context.Context, runParams RunParams, exePath string, errlog *log.Logger) error {
	debug.Println("running binary", exePath)
	theCmd := dryrun.Wrap(ctx, exePath, runParams.Args...)
	theCmd.Stderr = runParams.Stderr
	theCmd.Stdout = runParams.Stdout
	theCmd.Stdin = runParams.Stdin
	theCmd.Dir = runParams.Dir
	if runParams.WorkDir != runParams.Dir {
		theCmd.Dir = runParams.WorkDir
	}

	envMap, err := setupEnv(runParams)
	if err != nil {
		return fmt.Errorf("setting up environment for stavefile: %w", err)
	}
	theCmd.Env = env.ToAssignments(envMap)

	debug.Print("running stavefile with stave vars:\n", strings.Join(filter(theCmd.Env, "STAVEFILE"), "\n"))
	// catch SIGINT to allow stavefile to handle them
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT)
	defer signal.Stop(sigCh)
	err = theCmd.Run()
	if !sh.CmdRan(err) {
		errlog.Printf("failed to run compiled stavefile: %v", err)
	}
	return err
}

func setupEnv(runParams RunParams) (map[string]string, error) {
	envMap := env.GetMap()

	// We don't want to actually allow dryrun in the outermost invocation of
	// stave, since that will inhibit the very compilation of the stavefile & the
	// use of the resulting binary.
	// But every situation that's within such an execution is one in which dryrun
	// is supported, so we set this environment variable which will be carried
	// over throughout all such situations.
	envMap["STAVEFILE_DRYRUN_POSSIBLE"] = "1"

	if runParams.Verbose {
		envMap["STAVEFILE_VERBOSE"] = "1"
	}
	if runParams.List {
		envMap["STAVEFILE_LIST"] = "1"
	}
	if runParams.Info {
		envMap["STAVEFILE_INFO"] = "1"
	}
	if runParams.Debug {
		envMap["STAVEFILE_DEBUG"] = "1"
	}
	if runParams.GoCmd != "" {
		envMap["STAVEFILE_GOCMD"] = runParams.GoCmd
	}
	if runParams.Timeout > 0 {
		envMap["STAVEFILE_TIMEOUT"] = runParams.Timeout.String()
	}
	if runParams.DryRun {
		envMap["STAVEFILE_DRYRUN"] = "1"
	}

	if err := parallelism.Apply(envMap); err != nil {
		return nil, err
	}

	return envMap, nil
}

func filter(list []string, prefix string) []string {
	var out []string
	for _, s := range list {
		if strings.HasPrefix(s, prefix) {
			out = append(out, s)
		}
	}
	return out
}

// removeContents removes all files but not any subdirectories in the given
// directory.
func removeContents(dir string) error {
	debug.Println("removing all files in", dir)
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
