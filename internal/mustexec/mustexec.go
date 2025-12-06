package mustexec

import (
	"context"
	"io"
	"os/exec"

	"github.com/samber/lo"
	"github.com/yaklabco/stave/internal/env"
)

type mustExecOption struct {
	ctx        context.Context
	workingDir *string
	env        map[string]string
	stdin      *io.Reader
	stdout     *io.Writer
	stderr     *io.Writer
}

type Option func(*mustExecOption)

func WithContext(ctx context.Context) Option {
	return func(o *mustExecOption) {
		o.ctx = ctx
	}
}

func WithWorkingDir(dir string) Option {
	return func(o *mustExecOption) {
		o.workingDir = &dir
	}
}

func WithEnv(theEnv map[string]string) Option {
	return func(o *mustExecOption) {
		o.env = theEnv
	}
}

func WithStdin(r io.Reader) Option {
	return func(o *mustExecOption) {
		o.stdin = &r
	}
}

func WithStdout(w io.Writer) Option {
	return func(o *mustExecOption) {
		o.stdout = &w
	}
}

func WithStderr(w io.Writer) Option {
	return func(o *mustExecOption) {
		o.stderr = &w
	}
}

func MustExec(cmd string, args []string, options ...Option) {
	opts := mustExecOption{}
	for _, opt := range options {
		opt(&opts)
	}

	if opts.ctx == nil {
		opts.ctx = context.Background()
	}

	theCmd := exec.CommandContext(opts.ctx, cmd, args...)
	if !lo.IsNil(opts.workingDir) {
		theCmd.Dir = *opts.workingDir
	}
	if !lo.IsNil(opts.env) {
		theCmd.Env = env.ToAssignments(opts.env)
	}
	if !lo.IsNil(opts.stdin) {
		theCmd.Stdin = *opts.stdin
	}
	if !lo.IsNil(opts.stdout) {
		theCmd.Stdout = *opts.stdout
	}
	if !lo.IsNil(opts.stderr) {
		theCmd.Stderr = *opts.stderr
	}

	err := theCmd.Run()
	if err != nil {
		panic(err)
	}
}
