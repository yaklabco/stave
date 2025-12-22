package watch

import (
	"io"
	"os"

	"github.com/yaklabco/stave/internal/ish"
	"github.com/yaklabco/stave/pkg/watch/wctx"
)

func RunCmd(cmd string, args ...string) func(args ...string) error {
	return func(args2 ...string) error {
		return Run(cmd, append(args, args2...)...)
	}
}

func OutCmd(cmd string, args ...string) func(args ...string) (string, error) {
	return func(args2 ...string) (string, error) {
		return Output(cmd, append(args, args2...)...)
	}
}

func Run(cmd string, args ...string) error {
	return RunWith(nil, cmd, args...)
}

func RunV(cmd string, args ...string) error {
	_, err := Exec(nil, os.Stdin, os.Stdout, os.Stderr, cmd, args...)
	return err
}

func RunWith(env map[string]string, cmd string, args ...string) error {
	return ish.Run(wctx.GetActiveContext(), env, cmd, args...)
}

func RunWithV(env map[string]string, cmd string, args ...string) error {
	return ish.RunV(wctx.GetActiveContext(), env, cmd, args...)
}

func Output(cmd string, args ...string) (string, error) {
	return ish.Output(wctx.GetActiveContext(), nil, cmd, args...)
}

func OutputWith(env map[string]string, cmd string, args ...string) (string, error) {
	return ish.Output(wctx.GetActiveContext(), env, cmd, args...)
}

func Piper(stdin io.Reader, stdout, stderr io.Writer, cmd string, args ...string) error {
	return ish.Piper(wctx.GetActiveContext(), nil, stdin, stdout, stderr, cmd, args...)
}

func PiperWith(env map[string]string, stdin io.Reader, stdout, stderr io.Writer, cmd string, args ...string) error {
	return ish.Piper(wctx.GetActiveContext(), env, stdin, stdout, stderr, cmd, args...)
}

func Exec(env map[string]string, stdin io.Reader, stdout, stderr io.Writer, cmd string, args ...string) (bool, error) {
	return ish.Exec(wctx.GetActiveContext(), env, stdin, stdout, stderr, cmd, args...)
}

func Rm(path string) error {
	ctx := wctx.GetActiveContext()
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return ish.Rm(path)
}

func Copy(dst string, src string) error {
	ctx := wctx.GetActiveContext()
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return ish.Copy(dst, src)
}

func CmdRan(err error) bool {
	return ish.CmdRan(err)
}

func ExitStatus(err error) int {
	return ish.ExitStatus(err)
}
