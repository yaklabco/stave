package stave

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"

	"go/build"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/yaklabco/stave/internal"
	"github.com/yaklabco/stave/internal/dryrun"
	"github.com/yaklabco/stave/parse"
	"github.com/yaklabco/stave/sh"
	"github.com/yaklabco/stave/st"
)

const longAgoShift = -time.Hour * 24 * 365 * 10

// magicRebuildKey is used when hashing the output binary to ensure that we get
// a new binary even if nothing in the input files or generated mainfile has
// changed. This can be used when we change how we parse files, or otherwise
// change the inputs to the compiling process.
const magicRebuildKey = "v0.3"

// (Aaaa)(Bbbb) -> aaaaBbbb.
var firstWordRx = regexp.MustCompile(`^([[:upper:]][^[:upper:]]+)([[:upper:]].*)$`)

// (AAAA)(Bbbb) -> aaaaBbbb.
var firstAbbrevRx = regexp.MustCompile(`^([[:upper:]]+)([[:upper:]][^[:upper:]].*)$`)

func lowerFirstWord(str string) string {
	if match := firstWordRx.FindStringSubmatch(str); match != nil {
		return strings.ToLower(match[1]) + match[2]
	}
	if match := firstAbbrevRx.FindStringSubmatch(str); match != nil {
		return strings.ToLower(match[1]) + match[2]
	}
	return strings.ToLower(str)
}

//go:generate stringer -type=Command

// Command tracks invocations of stave that run without targets or other flags.
type Command int

// The various command types.
const (
	None          Command = iota
	Version               // report the current version of stave
	Init                  // create a starting template for stave
	Clean                 // clean out old compiled stave binaries from the cache
	CompileStatic         // compile a static binary of the current directory
)

// Main is the entrypoint for running stave.  It exists external to stave's main
// function to allow it to be used from other programs, specifically so you can
// go run a simple file that run's stave's Main.
func Main(ctx context.Context) int {
	return ParseAndRun(ctx, os.Stdout, os.Stderr, os.Stdin, os.Args[1:])
}

// Invocation contains the args for invoking a run of Stave.
type Invocation struct {
	Debug      bool          // turn on debug messages
	Dir        string        // directory to read stavefiles from
	WorkDir    string        // directory where stavefiles will run
	Force      bool          // forces recreation of the compiled binary
	Verbose    bool          // tells the stavefile to print out log statements
	List       bool          // tells the stavefile to print out a list of targets
	Help       bool          // tells the stavefile to print out help for a specific target
	Keep       bool          // tells stave to keep the generated main file after compiling
	DryRun     bool          // tells stave that all sh.Run* commands should print, but not execute
	Timeout    time.Duration // tells stave to set a timeout to running the targets
	CompileOut string        // tells stave to compile a static binary to this path, but not execute
	GOOS       string        // sets the GOOS when producing a binary with -compileout
	GOARCH     string        // sets the GOARCH when producing a binary with -compileout
	Ldflags    string        // sets the ldflags when producing a binary with -compileout
	Stdout     io.Writer     // writer to write stdout messages to
	Stderr     io.Writer     // writer to write stderr messages to
	Stdin      io.Reader     // reader to read stdin from
	Args       []string      // args to pass to the compiled binary
	GoCmd      string        // the go binary command to run
	CacheDir   string        // the directory where we should store compiled binaries
	HashFast   bool          // don't rely on GOCACHE, just hash the stavefiles
}

// StavefilesDirName is the name of the default folder to look for if no directory was specified,
// if this folder exists it will be assumed stave package lives inside it.
const StavefilesDirName = "stavefiles"

// UsesStavefiles returns true if we are getting our stave files from a stavefiles directory.
func (i Invocation) UsesStavefiles() bool {
	return filepath.Base(i.Dir) == StavefilesDirName
}

// ParseAndRun parses the command line, and then compiles and runs the stave
// files in the given directory with the given args (do not include the command
// name in the args).
func ParseAndRun(ctx context.Context, stdout, stderr io.Writer, stdin io.Reader, args []string) int {
	errlog := log.New(stderr, "", 0)
	out := log.New(stdout, "", 0)
	inv, cmd, err := Parse(stderr, stdout, args)
	if errors.Is(err, flag.ErrHelp) {
		return 0
	}
	if err != nil {
		errlog.Println("Error:", err)
		return 2
	}

	inv.Stderr = stderr
	inv.Stdin = stdin

	switch cmd {
	case Version:
		out.Println("Stave Build Tool", gitTag)
		out.Println("Build Date:", timestamp)
		out.Println("Commit:", commitHash)
		out.Println("Built with:", runtime.Version())
		return 0
	case Init:
		if err := generateInit(inv.Dir); err != nil {
			errlog.Println("Error:", err)
			return 1
		}
		out.Println(initFile, "created")
		return 0
	case Clean:
		if err := removeContents(inv.CacheDir); err != nil {
			out.Println("Error:", err)
			return 1
		}
		out.Println(inv.CacheDir, "cleaned")
		return 0
	case CompileStatic:
		return Invoke(ctx, inv)
	case None:
		return Invoke(ctx, inv)
	default:
		panic(fmt.Errorf("unknown command type: %v", cmd))
	}
}

// Parse parses the given args and returns structured data.  If parse returns
// flag.ErrHelp, the calling process should exit with code 0.
func Parse(stderr, stdout io.Writer, args []string) (Invocation, Command, error) {
	var inv Invocation
	var cmd Command

	inv.Stdout = stdout
	fs := flag.FlagSet{}
	fs.SetOutput(stdout)

	// options flags

	fs.BoolVar(&inv.Force, "f", false, "force recreation of compiled stavefile")
	fs.BoolVar(&inv.Debug, "debug", st.Debug(), "turn on debug messages")
	fs.BoolVar(&inv.Verbose, "v", st.Verbose(), "show verbose output when running stave targets")
	fs.BoolVar(&inv.Help, "h", false, "show this help")
	fs.DurationVar(&inv.Timeout, "t", 0, "timeout in duration parsable format (e.g. 5m30s)")
	fs.BoolVar(&inv.Keep, "keep", false, "keep intermediate stave files around after running")
	fs.BoolVar(&inv.DryRun, "dryrun", false, "print commands instead of executing them")
	fs.StringVar(&inv.Dir, "d", "", "directory to read stavefiles from")
	fs.StringVar(&inv.WorkDir, "w", "", "working directory where stavefiles will run")
	fs.StringVar(&inv.GoCmd, "gocmd", st.GoCmd(), "use the given go binary to compile the output")
	fs.StringVar(&inv.GOOS, "goos", "", "set GOOS for binary produced with -compile")
	fs.StringVar(&inv.GOARCH, "goarch", "", "set GOARCH for binary produced with -compile")
	fs.StringVar(&inv.Ldflags, "ldflags", "", "set ldflags for binary produced with -compile")

	// commands below

	fs.BoolVar(&inv.List, "l", false, "list stave targets in this directory")
	var showVersion bool
	fs.BoolVar(&showVersion, "version", false, "show version info for the stave binary")
	var staveInit bool
	fs.BoolVar(&staveInit, "init", false, "create a starting template if no stave files exist")
	var clean bool
	fs.BoolVar(&clean, "clean", false, "clean out old generated binaries from CACHE_DIR")
	var compileOutPath string
	fs.StringVar(&compileOutPath, "compile", "", "output a static binary to the given path")

	fs.Usage = func() {
		_, _ = fmt.Fprint(stdout, `
stave [options] [target]

Stave is a make-like command runner. Fork of Mage. See https://github.com/yaklabco/stave

Commands:
  -clean    clean out old generated binaries from CACHE_DIR
  -compile <string>
            output a static binary to the given path
  -h        show this help
  -init     create a starting template if no stave files exist
  -l        list targets in this directory
  -version  show version info for the stave binary

Options:
  -d <string> 
            directory to read stavefiles from (default "." or "stavefiles" if exists)
  -debug    turn on debug messages
  -dryrun   print commands instead of executing them
  -f        force recreation of compiled stavefile
  -goarch   sets the GOARCH for the binary created by -compile (default: current arch)
  -gocmd <string>
            use the given go binary to compile the output (default: "go")
  -goos     sets the GOOS for the binary created by -compile (default: current OS)
  -ldflags  sets the ldflags for the binary created by -compile (default: "")
  -h        show description of a target
  -keep     keep intermediate stave files around after running
  -t <string>
            timeout in duration parsable format (e.g. 5m30s)
  -v        show verbose output when running targets
  -w <string>
            working directory where stavefiles will run (default -d value)
`[1:])
	}
	err := fs.Parse(args)
	if errors.Is(err, flag.ErrHelp) {
		// parse will have already called fs.Usage()
		return inv, cmd, err
	}
	if err == nil && inv.Help && len(fs.Args()) == 0 {
		fs.Usage()
		// tell upstream, to just exit
		return inv, cmd, flag.ErrHelp
	}

	numCommands := 0
	switch {
	case staveInit:
		numCommands++
		cmd = Init
	case compileOutPath != "":
		numCommands++
		cmd = CompileStatic
		inv.CompileOut = compileOutPath
		inv.Force = true
	case showVersion:
		numCommands++
		cmd = Version
	case clean:
		numCommands++
		cmd = Clean
		if fs.NArg() > 0 {
			// Temporary dupe of below check until we refactor the other commands to use this check
			return inv, cmd, errors.New("-h, -init, -clean, -compile and -version cannot be used simultaneously")
		}
	}
	if inv.Help {
		numCommands++
	}

	if inv.Debug {
		debug.SetOutput(stderr)
	}

	if inv.DryRun {
		dryrun.SetRequested(true)
	}

	inv.CacheDir = st.CacheDir()

	if numCommands > 1 {
		debug.Printf("%d commands defined", numCommands)
		return inv, cmd, errors.New("-h, -init, -clean, -compile and -version cannot be used simultaneously")
	}

	if cmd != CompileStatic && (inv.GOARCH != "" || inv.GOOS != "") {
		return inv, cmd, errors.New("-goos and -goarch only apply when running with -compile")
	}

	inv.Args = fs.Args()
	if inv.Help && len(inv.Args) > 1 {
		return inv, cmd, errors.New("-h can only show help for a single target")
	}

	if len(inv.Args) > 0 && cmd != None {
		return inv, cmd, fmt.Errorf("unexpected arguments to command: %q", inv.Args)
	}
	inv.HashFast = st.HashFast()
	return inv, cmd, err
}

const dotDirectory = "."

// Invoke runs Stave with the given arguments.
func Invoke(ctx context.Context, inv Invocation) int {
	errlog := log.New(inv.Stderr, "", 0)
	if inv.GoCmd == "" {
		inv.GoCmd = "go"
	}
	if inv.Dir == "" {
		inv.Dir = dotDirectory
	}
	if inv.WorkDir == "" {
		inv.WorkDir = inv.Dir
	}
	stavefilesDir := filepath.Join(inv.Dir, StavefilesDirName)
	// . will be default unless we find a stave folder.
	mfSt, err := os.Stat(stavefilesDir)
	if err == nil {
		if mfSt.IsDir() {
			originalDir := inv.Dir
			inv.Dir = stavefilesDir // preemptive assignment
			// TODO: Remove this fallback and the above Stavefiles invocation when the bw compatibility is removed.
			files, err := Stavefiles(originalDir, inv.GOOS, inv.GOARCH, false)
			if err == nil {
				if len(files) != 0 {
					errlog.Println("[WARNING] You have both a stavefiles directory and stave files in the " +
						"current directory, in future versions the files will be ignored in favor of the directory")
					inv.Dir = originalDir
				}
			}
		}
	}

	if inv.CacheDir == "" {
		inv.CacheDir = st.CacheDir()
	}

	files, err := Stavefiles(inv.Dir, inv.GOOS, inv.GOARCH, inv.UsesStavefiles())
	if err != nil {
		errlog.Println("Error determining list of stavefiles:", err)
		return 1
	}

	if len(files) == 0 {
		errlog.Println("No .go files marked with the stave build tag in this directory.")
		return 1
	}
	debug.Printf("found stavefiles: %s", strings.Join(files, ", "))
	exePath := inv.CompileOut
	if inv.CompileOut == "" {
		exePath, err = ExeName(ctx, inv.GoCmd, inv.CacheDir, files)
		if err != nil {
			errlog.Println("Error getting exe name:", err)
			return 1
		}
	}
	debug.Println("output exe is ", exePath)

	useCache := false
	if inv.HashFast {
		debug.Println("user has set STAVEFILE_HASHFAST, so we'll ignore GOCACHE")
	} else {
		theGoCache, err := internal.OutputDebug(ctx, inv.GoCmd, "env", "GOCACHE")
		if err != nil {
			errlog.Printf("failed to run %s env GOCACHE: %s", inv.GoCmd, err)
			return 1
		}

		// if GOCACHE exists, always rebuild, so we catch transitive
		// dependencies that have changed.
		if theGoCache != "" {
			debug.Println("go build cache exists, will ignore any compiled binary")
			useCache = true
		}
	}

	if !useCache {
		_, err = os.Stat(exePath)
		switch {
		case err == nil:
			if !inv.Force {
				debug.Println("Running existing exe")
				return RunCompiled(ctx, inv, exePath, errlog)
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
	if inv.Debug {
		parse.EnableDebug()
	}
	debug.Println("parsing files")
	info, err := parse.PrimaryPackage(ctx, inv.GoCmd, inv.Dir, fnames)
	if err != nil {
		errlog.Println("Error parsing stavefiles:", err)
		return 1
	}

	// reproducible output for deterministic builds
	sort.Sort(info.Funcs)
	sort.Sort(info.Imports)

	main := filepath.Join(inv.Dir, mainfile)
	binaryName := "stave"
	if inv.CompileOut != "" {
		binaryName = filepath.Base(inv.CompileOut)
	}

	err = GenerateMainfile(binaryName, main, info)
	if err != nil {
		errlog.Println("Error:", err)
		return 1
	}
	if !inv.Keep {
		defer func() { _ = os.RemoveAll(main) }()
	}
	files = append(files, main)
	if err := Compile(ctx, CompileParams{
		Goos:      inv.GOOS,
		Goarch:    inv.GOARCH,
		Ldflags:   inv.Ldflags,
		StavePath: inv.Dir,
		GoCmd:     inv.GoCmd,
		CompileTo: exePath,
		Gofiles:   files,
		Debug:     inv.Debug,
		Stderr:    inv.Stderr,
		Stdout:    inv.Stdout,
	}); err != nil {
		errlog.Println("Error:", err)
		return 1
	}
	if !inv.Keep {
		// move aside this file before we run the compiled version, in case the
		// compiled file screws things up.  Yes this doubles up with the above
		// defer, that's ok.
		_ = os.RemoveAll(main)
	} else {
		debug.Print("keeping mainfile")
	}

	if inv.CompileOut != "" {
		return 0
	}

	return RunCompiled(ctx, inv, exePath, errlog)
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
func listGoFiles(stavePath, tag string, envStr []string) ([]string, error) {
	origStavePath := stavePath
	if !filepath.IsAbs(stavePath) {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("can't get current working directory: %w", err)
		}
		stavePath = filepath.Join(cwd, stavePath)
	}

	env, err := internal.SplitEnv(envStr)
	if err != nil {
		return nil, fmt.Errorf("error parsing environment variables: %w", err)
	}

	bctx := build.Default
	bctx.BuildTags = []string{tag}

	if _, ok := env["GOOS"]; ok {
		bctx.GOOS = env["GOOS"]
	}

	if _, ok := env["GOARCH"]; ok {
		bctx.GOARCH = env["GOARCH"]
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

	env, err := internal.EnvWithGOOS(goos, goarch)
	if err != nil {
		return nil, err
	}

	debug.Println("getting all files including those with stave tag in", stavePath)
	staveFiles, err := listGoFiles(stavePath, "stave", env)
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
	nonStaveFiles, err := listGoFiles(stavePath, "", env)
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
	environ, err := internal.EnvWithGOOS(params.Goos, params.Goarch)
	if err != nil {
		return err
	}
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
	theCmd.Env = environ
	theCmd.Stderr = params.Stderr
	theCmd.Stdout = params.Stdout
	theCmd.Dir = params.StavePath
	start := time.Now()
	err = theCmd.Run()
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
	hash := sha256.Sum256([]byte(strings.Join(hashes, "") + magicRebuildKey + ver))
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
func RunCompiled(ctx context.Context, inv Invocation, exePath string, errlog *log.Logger) int {
	debug.Println("running binary", exePath)
	theCmd := dryrun.Wrap(ctx, exePath, inv.Args...)
	theCmd.Stderr = inv.Stderr
	theCmd.Stdout = inv.Stdout
	theCmd.Stdin = inv.Stdin
	theCmd.Dir = inv.Dir
	if inv.WorkDir != inv.Dir {
		theCmd.Dir = inv.WorkDir
	}

	// intentionally pass through unaltered os.Environ here.. your stavefile has
	// to deal with it.
	theCmd.Env = os.Environ()

	// We don't want to actually allow dryrun in the outermost invocation of
	// stave, since that will inhibit the very compilation of the stavefile & the
	// use of the resulting binary.
	// But every situation that's within such an execution is one in which dryrun
	// is supported, so we set this environment variable which will be carried
	// over throughout all such situations.
	theCmd.Env = append(theCmd.Env, "STAVEFILE_DRYRUN_POSSIBLE=1")

	if inv.Verbose {
		theCmd.Env = append(theCmd.Env, "STAVEFILE_VERBOSE=1")
	}
	if inv.List {
		theCmd.Env = append(theCmd.Env, "STAVEFILE_LIST=1")
	}
	if inv.Help {
		theCmd.Env = append(theCmd.Env, "STAVEFILE_HELP=1")
	}
	if inv.Debug {
		theCmd.Env = append(theCmd.Env, "STAVEFILE_DEBUG=1")
	}
	if inv.GoCmd != "" {
		theCmd.Env = append(theCmd.Env, "STAVEFILE_GOCMD="+inv.GoCmd)
	}
	if inv.Timeout > 0 {
		theCmd.Env = append(theCmd.Env, "STAVEFILE_TIMEOUT="+inv.Timeout.String())
	}
	if inv.DryRun {
		theCmd.Env = append(theCmd.Env, "STAVEFILE_DRYRUN=1")
	}
	debug.Print("running stavefile with stave vars:\n", strings.Join(filter(theCmd.Env, "STAVEFILE"), "\n"))
	// catch SIGINT to allow stavefile to handle them
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT)
	defer signal.Stop(sigCh)
	err := theCmd.Run()
	if !sh.CmdRan(err) {
		errlog.Printf("failed to run compiled stavefile: %v", err)
	}
	return sh.ExitStatus(err)
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
