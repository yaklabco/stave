package ish

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/yaklabco/stave/internal/dryrun"
	"github.com/yaklabco/stave/internal/log"
	"github.com/yaklabco/stave/pkg/st"
)

// Exec executes the command, piping its stdout and stderr to the given
// writers.
func Exec(ctx context.Context, env map[string]string, stdin io.Reader, stdout, stderr io.Writer, cmd string, args ...string) (bool, error) {
	expand := func(varName string) string {
		if env != nil {
			s2, ok := env[varName]
			if ok {
				return s2
			}
		}
		return os.Getenv(varName)
	}

	cmd = os.Expand(cmd, expand)

	for i := range args {
		args[i] = os.Expand(args[i], expand)
	}

	ran, code, err := run(ctx, env, stdin, stdout, stderr, cmd, args...)
	if err == nil {
		return true, nil
	}
	if ran {
		return ran, st.Fatalf(code, `running "%s %s" failed with exit code %d`, cmd, strings.Join(args, " "), code)
	}
	return ran, fmt.Errorf(`failed to run "%s %s: %w"`, cmd, strings.Join(args, " "), err)
}

func run(ctx context.Context, env map[string]string, stdin io.Reader, stdout, stderr io.Writer, cmd string, args ...string) (bool, int, error) {
	theCmd := dryrun.Wrap(ctx, cmd, args...)
	theCmd.Env = os.Environ()
	for k, v := range env {
		theCmd.Env = append(theCmd.Env, k+"="+v)
	}
	theCmd.Stderr = stderr
	theCmd.Stdout = stdout
	theCmd.Stdin = stdin

	quoted := make([]string, 0, len(args))
	for i := range args {
		quoted = append(quoted, fmt.Sprintf("%q", args[i]))
	}
	// To protect against logging from doing exec in global variables
	if st.Verbose() {
		log.SimpleConsoleLogger.Println("exec:", cmd, strings.Join(quoted, " "))
	}
	err := theCmd.Run()

	return CmdRan(err), ExitStatus(err), err
}

// CmdRan examines the error to determine if it was generated as a result of a
// command running via os/exec.Command.
func CmdRan(err error) bool {
	if err == nil {
		return true
	}
	var ee *exec.ExitError
	ok := errors.As(err, &ee)
	if ok {
		return ee.Exited()
	}
	return false
}

// ExitStatus returns the exit status of the error if it is an exec.ExitError
// or if it implements ExitStatus() int.
func ExitStatus(err error) int {
	if err == nil {
		return 0
	}
	var exit st.ExitStatuser
	if errors.As(err, &exit) {
		return exit.ExitStatus()
	}
	var e *exec.ExitError
	if errors.As(err, &e) {
		if ex, ok := e.Sys().(st.ExitStatuser); ok {
			return ex.ExitStatus()
		}
	}
	return 1
}

// Rm removes the given file or directory even if non-empty.
func Rm(path string) error {
	if dryrun.IsDryRun() {
		_, err := fmt.Println("DRYRUN: rm", path) //nolint:forbidigo // This is intentional console output.
		return err
	}

	err := os.RemoveAll(path)
	if err == nil || os.IsNotExist(err) {
		return nil
	}
	return fmt.Errorf(`failed to remove %s: %w`, path, err)
}

// Copy robustly copies the source file to the destination.
func Copy(dst string, src string) error {
	if dryrun.IsDryRun() {
		_, err := fmt.Println("DRYRUN: cp", src, dst) //nolint:forbidigo // This is intentional console output.
		return err
	}

	from, err := os.Open(src)
	if err != nil {
		return fmt.Errorf(`can't copy %s: %w`, src, err)
	}
	defer func() { _ = from.Close() }()
	finfo, err := from.Stat()
	if err != nil {
		return fmt.Errorf(`can't stat %s: %w`, src, err)
	}
	to, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, finfo.Mode())
	if err != nil {
		return fmt.Errorf(`can't copy to %s: %w`, dst, err)
	}
	defer func() { _ = to.Close() }()
	_, err = io.Copy(to, from)
	if err != nil {
		return fmt.Errorf(`error copying %s to %s: %w`, src, dst, err)
	}
	return nil
}

// Higher-level functions

func Run(ctx context.Context, env map[string]string, cmd string, args ...string) error {
	var output io.Writer
	if st.Verbose() || dryrun.IsDryRun() {
		output = os.Stdout
	}
	_, err := Exec(ctx, env, os.Stdin, output, os.Stderr, cmd, args...)
	return err
}

func RunV(ctx context.Context, env map[string]string, cmd string, args ...string) error {
	_, err := Exec(ctx, env, os.Stdin, os.Stdout, os.Stderr, cmd, args...)
	return err
}

func Output(ctx context.Context, env map[string]string, cmd string, args ...string) (string, error) {
	buf := &bytes.Buffer{}
	_, err := Exec(ctx, env, os.Stdin, buf, os.Stderr, cmd, args...)
	return strings.TrimSuffix(buf.String(), "\n"), err
}

func Piper(ctx context.Context, env map[string]string, stdin io.Reader, stdout, stderr io.Writer, cmd string, args ...string) error {
	_, err := Exec(ctx, env, stdin, stdout, stderr, cmd, args...)
	return err
}
