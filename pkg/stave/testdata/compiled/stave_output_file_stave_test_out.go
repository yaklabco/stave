//go:build ignore

package main

import (
	"context"
	_flag "flag"
	_fmt "fmt"
	_io "io"
	_log "log"
	"os"
	"os/signal"
	_filepath "path/filepath"
	"strconv"
	_strings "strings"
	"syscall"
	"time"

	st_staveimport "github.com/yaklabco/stave/pkg/st"
)

func main() { // Use local types and functions in order to avoid name conflicts with additional stavefiles.
	_ = st_staveimport.ResetOnces
	type arguments struct {
		Verbose bool          // print out log statements
		Debug   bool          // print out more detailed logs
		Info    bool          // print out docstring for a specific target
		Timeout time.Duration // set a timeout to running the targets
		Args    []string      // args contain the non-flag command-line arguments
	}

	// parseBool implements the same semantics as internal/env.ParseBool:
	// true: "true", "yes", "1"; false: "false", "no", "0" (case- and
	// whitespace-insensitive). Any other non-empty value falls back to false.
	parseBool := func(envVar string) bool {
		val := _strings.TrimSpace(os.Getenv(envVar))
		if val == "" {
			return false
		}
		switch _strings.ToLower(val) {
		case "true", "yes", "1":
			return true
		case "false", "no", "0":
			return false
		default:
			return false
		}
	}

	parseDuration := func(env string) time.Duration {
		val := os.Getenv(env)
		if val == "" {
			return 0
		}
		d, err := time.ParseDuration(val)
		if err != nil {
			_log.Printf("warning: environment variable %s is not a valid duration value: %v", env, val)
			return 0
		}
		return d
	}

	args := arguments{}
	fs := _flag.FlagSet{}
	fs.SetOutput(os.Stdout)

	// default flag set with ExitOnError and auto generated PrintDefaults should be sufficient
	var verboseLong bool
	fs.BoolVar(&args.Verbose, "v", parseBool("STAVEFILE_VERBOSE"), "show verbose output when running targets")
	fs.BoolVar(&verboseLong, "verbose", parseBool("STAVEFILE_VERBOSE"), "show verbose output when running targets")
	var debugLong bool
	fs.BoolVar(&args.Debug, "d", parseBool("STAVEFILE_DEBUG"), "print out more detailed logs")
	fs.BoolVar(&debugLong, "debug", parseBool("STAVEFILE_DEBUG"), "print out more detailed logs")
	var infoLong bool
	fs.BoolVar(&args.Info, "i", parseBool("STAVEFILE_INFO"), "print out docstring for a specific target")
	fs.BoolVar(&infoLong, "info", parseBool("STAVEFILE_INFO"), "print out docstring for a specific target")
	var timeoutLong time.Duration
	fs.DurationVar(&args.Timeout, "t", parseDuration("STAVEFILE_TIMEOUT"), "timeout in duration parsable format (e.g. 5m30s)")
	fs.DurationVar(&timeoutLong, "timeout", parseDuration("STAVEFILE_TIMEOUT"), "timeout in duration parsable format (e.g. 5m30s)")

	fs.Usage = func() {
		_fmt.Fprintf(os.Stdout, `
		%s [options] [target]

	Commands:
		-h --info      show this help

	Options:
		-i --info      show description of a target
		-t             <string>
                   timeout in duration parsable format (e.g. 5m30s)
		-v --verbose   show verbose output when running targets
		-d --debug     emit detailed logs
		`[1:], _filepath.Base(os.Args[0]))
	}
	if err := fs.Parse(os.Args[1:]); err != nil {
		// flag will have printed out an error already.
		return
	}
	args.Args = fs.Args()
	if verboseLong != parseBool("STAVEFILE_VERBOSE") {
		args.Verbose = verboseLong
	}
	if debugLong != parseBool("STAVEFILE_DEBUG") {
		args.Debug = debugLong
	}
	if infoLong != parseBool("STAVEFILE_INFO") {
		args.Info = infoLong
	}
	if timeoutLong != parseDuration("STAVEFILE_TIMEOUT") {
		args.Timeout = timeoutLong
	}
	if args.Info && len(args.Args) == 0 {
		fs.Usage()
		return
	}

	// Set the outermost target name.
	outermost := ""
	if len(args.Args) > 0 {
		outermost = args.Args[0]
		// Resolve alias
		switch _strings.ToLower(outermost) {

		}
	} else {
		outermost = "Deploy"
	}

	// color is ANSI color type
	type color int

	const (
		black color = iota
		red
		green
		yellow
		blue
		staventa
		cyan
		white
		brightblack
		brightred
		brightgreen
		brightyellow
		brightblue
		brightstaventa
		brightcyan
		brightwhite
	)

	// AnsiColor are ANSI color codes for supported terminal colors.
	var ansiColor = map[color]string{
		black:          "\u001b[30m",
		red:            "\u001b[31m",
		green:          "\u001b[32m",
		yellow:         "\u001b[33m",
		blue:           "\u001b[34m",
		staventa:       "\u001b[35m",
		cyan:           "\u001b[36m",
		white:          "\u001b[37m",
		brightblack:    "\u001b[30;1m",
		brightred:      "\u001b[31;1m",
		brightgreen:    "\u001b[32;1m",
		brightyellow:   "\u001b[33;1m",
		brightblue:     "\u001b[34;1m",
		brightstaventa: "\u001b[35;1m",
		brightcyan:     "\u001b[36;1m",
		brightwhite:    "\u001b[37;1m",
	}

	const _color_name = "blackredgreenyellowbluestaventacyanwhitebrightblackbrightredbrightgreenbrightyellowbrightbluebrightstaventabrightcyanbrightwhite"

	var _color_index = [...]uint8{0, 5, 8, 13, 19, 23, 30, 34, 39, 50, 59, 70, 82, 92, 105, 115, 126}

	colorToLowerString := func(i color) string {
		if i < 0 || i >= color(len(_color_index)-1) {
			return "color(" + strconv.FormatInt(int64(i), 10) + ")"
		}
		return _color_name[_color_index[i]:_color_index[i+1]]
	}

	// ansiColorReset is an ANSI color code to reset the terminal color.
	const ansiColorReset = "\033[0m"

	// defaultTargetAnsiColor is a default ANSI color for colorizing targets.
	// It is set to Cyan as an arbitrary color, because it has a neutral meaning
	var defaultTargetAnsiColor = ansiColor[cyan]

	getAnsiColor := func(color string) (string, bool) {
		colorLower := _strings.ToLower(color)
		for k, v := range ansiColor {
			colorConstLower := colorToLowerString(k)
			if colorConstLower == colorLower {
				return v, true
			}
		}
		return "", false
	}

	// Terminals which don't support color (populated from pkg/st.NoColorTERMs).
	var noColorTerms = map[string]struct{}{
		"cygwin":     {},
		"dumb":       {},
		"vt100":      {},
		"xterm-mono": {},
	}

	// terminalSupportsColor checks if the current console supports color output
	//
	// Supported:
	// 	linux, mac, or windows's ConEmu, Cmder, putty, git-bash.exe, pwsh.exe
	// Not supported:
	// 	windows cmd.exe, powerShell.exe
	terminalSupportsColor := func() bool {
		envTerm := os.Getenv("TERM")
		if _, ok := noColorTerms[envTerm]; ok {
			return false
		}
		return true
	}

	// enableColor reports whether the user has requested to enable a color output.
	enableColor := func() bool {
		return parseBool("STAVEFILE_ENABLE_COLOR")
	}

	// targetColor returns the ANSI color which should be used to colorize targets.
	targetColor := func() string {
		s, exists := os.LookupEnv("STAVEFILE_TARGET_COLOR")
		if exists == true {
			if c, ok := getAnsiColor(s); ok == true {
				return c
			}
		}
		return defaultTargetAnsiColor
	}

	// store the color terminal variables, so that the detection isn't repeated for each target
	var enableColorValue = enableColor() && terminalSupportsColor()
	var targetColorValue = targetColor()

	printName := func(str string) string {
		if enableColorValue {
			return _fmt.Sprintf("%s%s%s", targetColorValue, str, ansiColorReset)
		} else {
			return str
		}
	}

	var ctx context.Context
	ctxCancel := func() {}

	mainCtx, mainCancel := context.WithCancel(context.Background())
	defer mainCancel()

	// by deferring in a closure, we let the cancel function get replaced
	// by the getContext function.
	defer func() {
		ctxCancel()
	}()

	getContext := func() (context.Context, func()) {
		if ctx == nil || ctx.Err() != nil {
			if args.Timeout != 0 {
				ctx, ctxCancel = context.WithTimeout(mainCtx, args.Timeout)
			} else {
				ctx, ctxCancel = context.WithCancel(mainCtx)
			}
		}

		return ctx, ctxCancel
	}

	runTarget := func(logger *_log.Logger, name string, fn func(context.Context) error) any {
		var err any
		ctx, _ := getContext()
		d := make(chan any, 2)
		go func() {
			var err any
			defer func() {
				if r := recover(); r != nil {
					err = r
				}
				d <- err
			}()
			err = fn(ctx)
			d <- err
		}()
		select {
		case <-ctx.Done():
			logger.Println("cancelling stave targets, waiting up to 5 seconds for cleanup...")
			cleanupCh := time.After(5 * time.Second)

			select {
			// target exited by itself
			case err = <-d:
				if err == nil && ctx.Err() == context.DeadlineExceeded {
					err = ctx.Err()
				}
				return err
			// cleanup timeout exceeded
			case <-cleanupCh:
				if ctx.Err() == context.DeadlineExceeded {
					return _fmt.Errorf("cleanup timeout exceeded: %w", ctx.Err())
				}
				return _fmt.Errorf("cleanup timeout exceeded")
			}
		case err = <-d:
			// we intentionally don't cancel the context here, because
			// the next target will need to run with the same context.
			return err
		}
	}
	// This is necessary in case there aren't any targets, to avoid an unused
	// variable error.
	_ = runTarget

	handleError := func(logger *_log.Logger, err any) {
		if err != nil {
			logger.Printf("Error: %+v\n", err)
			type code interface {
				ExitStatus() int
			}
			if c, ok := err.(code); ok {
				os.Exit(c.ExitStatus())
			}
			os.Exit(1)
		}
	}
	_ = handleError

	// Set STAVEFILE_VERBOSE so st.Verbose() reflects the flag value.
	if args.Verbose {
		os.Setenv("STAVEFILE_VERBOSE", "1")
	} else {
		os.Setenv("STAVEFILE_VERBOSE", "0")
	}

	// Set STAVEFILE_DEBUG so st.Debug() reflects the flag value.
	if args.Debug {
		os.Setenv("STAVEFILE_DEBUG", "1")
	} else {
		os.Setenv("STAVEFILE_DEBUG", "0")
	}

	if args.Debug {
		// Debug
	} else {
		if args.Verbose {
			// Info
		} else {
			// Warn
		}
	}

	_log.SetFlags(0)
	if !args.Verbose {
		_log.SetOutput(_io.Discard)
	}
	logger := _log.New(os.Stderr, "", 0)
	globalSigCh := make(chan os.Signal, 1)
	signal.Notify(globalSigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-globalSigCh
		mainCancel()
		<-globalSigCh
		_fmt.Fprintln(os.Stderr, "exiting stave")
		handleError(logger, _fmt.Errorf("exit forced"))
	}()
	if args.Info {
		if len(args.Args) < 1 {
			logger.Println("no target specified")
			os.Exit(2)
		}
		switch _strings.ToLower(args.Args[0]) {
		case "deploy":
			_fmt.Println("This is the synopsis for Deploy. This part shouldn't show up.")
			_fmt.Println()

			_fmt.Print("Usage:\n\n\tstave_test_out deploy\n\n")
			var aliases []string
			if len(aliases) > 0 {
				_fmt.Printf("Aliases: %s\n\n", _strings.Join(aliases, ", "))
			}
			return
		case "printverboseflag":
			_fmt.Println("PrintVerboseFlag prints the value of st.Verbose() to stdout.")
			_fmt.Println()

			_fmt.Print("Usage:\n\n\tstave_test_out printverboseflag\n\n")
			var aliases []string
			if len(aliases) > 0 {
				_fmt.Printf("Aliases: %s\n\n", _strings.Join(aliases, ", "))
			}
			return
		case "sleep":
			_fmt.Println("Sleep sleeps 5 seconds.")
			_fmt.Println()

			_fmt.Print("Usage:\n\n\tstave_test_out sleep\n\n")
			var aliases []string
			if len(aliases) > 0 {
				_fmt.Printf("Aliases: %s\n\n", _strings.Join(aliases, ", "))
			}
			return
		case "testverbose":
			_fmt.Println("This is very verbose.")
			_fmt.Println()

			_fmt.Print("Usage:\n\n\tstave_test_out testverbose\n\n")
			var aliases []string
			if len(aliases) > 0 {
				_fmt.Printf("Aliases: %s\n\n", _strings.Join(aliases, ", "))
			}
			return
		default:
			logger.Printf("Unknown target: %q\n", args.Args[0])
			os.Exit(2)
		}
	}
	runAllTargets := func() any {
		if len(args.Args) < 1 {
			if parseBool("STAVEFILE_IGNOREDEFAULT") {
				logger.Println("Error: STAVEFILE_IGNOREDEFAULT is on and no target specified.")
				os.Exit(1)
			}
			run := func() any {
				_targetArgs := []string{}
				_ = _targetArgs

				wrapFn := func(ctx context.Context) error {
					Deploy()
					return nil
				}
				ret := runTarget(logger, "Deploy", wrapFn)
				return ret
			}
			return run()
		}

		hooksAreRunning := parseBool("STAVEFILE_HOOKS_RUNNING")
		for iArg := 0; iArg < len(args.Args); {
			target := args.Args[iArg]
			iArg++

			// resolve aliases
			switch _strings.ToLower(target) {

			}

			var ret any
			switch _strings.ToLower(target) {

			case "deploy":
				expected := iArg + 0
				if expected > len(args.Args) {
					// note that expected and args at this point include the arg for the target itself
					// so we subtract 1 here to show the number of args without the target.
					logger.Printf("not enough arguments for target \"Deploy\", expected %v, got %v\n", expected-1, len(args.Args)-1)
					os.Exit(2)
				}
				if args.Verbose {
					logger.Println("Running target: <Deploy>")
				}
				_targetArgs := args.Args[iArg:expected]
				iArg = expected
				run := func() any {
					_ = _targetArgs

					wrapFn := func(ctx context.Context) error {
						Deploy()
						return nil
					}
					ret := runTarget(logger, "Deploy", wrapFn)
					return ret
				}
				ret = run()
			case "printverboseflag":
				expected := iArg + 0
				if expected > len(args.Args) {
					// note that expected and args at this point include the arg for the target itself
					// so we subtract 1 here to show the number of args without the target.
					logger.Printf("not enough arguments for target \"PrintVerboseFlag\", expected %v, got %v\n", expected-1, len(args.Args)-1)
					os.Exit(2)
				}
				if args.Verbose {
					logger.Println("Running target: <PrintVerboseFlag>")
				}
				_targetArgs := args.Args[iArg:expected]
				iArg = expected
				run := func() any {
					_ = _targetArgs

					wrapFn := func(ctx context.Context) error {
						PrintVerboseFlag()
						return nil
					}
					ret := runTarget(logger, "PrintVerboseFlag", wrapFn)
					return ret
				}
				ret = run()
			case "sleep":
				expected := iArg + 0
				if expected > len(args.Args) {
					// note that expected and args at this point include the arg for the target itself
					// so we subtract 1 here to show the number of args without the target.
					logger.Printf("not enough arguments for target \"Sleep\", expected %v, got %v\n", expected-1, len(args.Args)-1)
					os.Exit(2)
				}
				if args.Verbose {
					logger.Println("Running target: <Sleep>")
				}
				_targetArgs := args.Args[iArg:expected]
				iArg = expected
				run := func() any {
					_ = _targetArgs

					wrapFn := func(ctx context.Context) error {
						Sleep()
						return nil
					}
					ret := runTarget(logger, "Sleep", wrapFn)
					return ret
				}
				ret = run()
			case "testverbose":
				expected := iArg + 0
				if expected > len(args.Args) {
					// note that expected and args at this point include the arg for the target itself
					// so we subtract 1 here to show the number of args without the target.
					logger.Printf("not enough arguments for target \"TestVerbose\", expected %v, got %v\n", expected-1, len(args.Args)-1)
					os.Exit(2)
				}
				if args.Verbose {
					logger.Println("Running target: <TestVerbose>")
				}
				_targetArgs := args.Args[iArg:expected]
				iArg = expected
				run := func() any {
					_ = _targetArgs

					wrapFn := func(ctx context.Context) error {
						TestVerbose()
						return nil
					}
					ret := runTarget(logger, "TestVerbose", wrapFn)
					return ret
				}
				ret = run()

			default:
				logger.Printf("Unknown target specified: %q\n", target)
				os.Exit(2)
			}

			if ret != nil {
				return ret
			}

			// If hooks are running, the remainder of the command-line might just be unused hook arguments; instead of treating them as targets, we ignore them.
			if hooksAreRunning {
				if args.Verbose && iArg < len(args.Args) {
					logger.Println(_fmt.Sprintf("Skipping unused arguments in target %s: %v", printName(target), args.Args[iArg:]))
				}
				break
			}
		}
		return nil
	}

	ret := runAllTargets()

	handleError(logger, ret)
}
