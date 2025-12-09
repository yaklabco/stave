package internal

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"strings"

	"github.com/yaklabco/stave/internal/dryrun"
	"github.com/yaklabco/stave/internal/log"
	"github.com/yaklabco/stave/pkg/env"
)

const (
	GoOSEnvVar   = "GOOS"
	GoArchEnvVar = "GOARCH"
)

func RunDebug(ctx context.Context, cmd string, args ...string) error {
	envMap := EnvWithCurrentGOOS()

	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}

	slog.Debug("running command", slog.String(log.Cmd, cmd), slog.Any(log.Args, args))
	theCmd := dryrun.Wrap(ctx, cmd, args...)
	theCmd.Env = env.ToAssignments(envMap)
	theCmd.Stderr = errBuf
	theCmd.Stdout = outBuf

	if err := theCmd.Run(); err != nil {
		slog.Debug(
			"error running command",
			slog.String(log.Cmd, cmd),
			slog.Any(log.Args, args),
			slog.Any(log.Error, err),
			slog.String(log.Stderr, errBuf.String()),
		)
		return err
	}

	slog.Debug(
		"command ran successfully",
		slog.String(log.Cmd, cmd),
		slog.Any(log.Args, args),
		slog.String(log.Stdout, outBuf.String()),
	)

	return nil
}

func OutputDebug(ctx context.Context, cmd string, args ...string) (string, error) {
	envMap := EnvWithCurrentGOOS()

	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}

	slog.Debug("running command", slog.String(log.Cmd, cmd), slog.Any(log.Args, args))
	theCmd := dryrun.Wrap(ctx, cmd, args...)
	theCmd.Env = env.ToAssignments(envMap)
	theCmd.Stderr = errBuf
	theCmd.Stdout = outBuf

	if err := theCmd.Run(); err != nil {
		errMsg := strings.TrimSpace(errBuf.String())
		slog.Debug(
			"error running command",
			slog.String(log.Cmd, cmd),
			slog.Any(log.Args, args),
			slog.Any(log.Error, err),
			slog.String(log.Stderr, errBuf.String()),
		)
		return "", fmt.Errorf("error running \"%s %s\": %w\n%s", cmd, strings.Join(args, " "), err, errMsg)
	}

	return strings.TrimSpace(outBuf.String()), nil
}

// EnvWithCurrentGOOS creates an env map using the current GOOS and GOARCH.
func EnvWithCurrentGOOS() map[string]string {
	return EnvWithGOOS(runtime.GOOS, runtime.GOARCH)
}

// EnvWithGOOS creates an env map with GOOS and GOARCH values set explicitly.
// If goos or goarch is empty, defaults to runtime.GOOS or runtime.GOARCH.
// Returns a modified environment map based on input and current settings.
func EnvWithGOOS(goos, goarch string) map[string]string {
	envMap := env.GetMap()
	if goos == "" {
		envMap[GoOSEnvVar] = runtime.GOOS
	} else {
		envMap[GoOSEnvVar] = goos
	}
	if goarch == "" {
		envMap[GoArchEnvVar] = runtime.GOARCH
	} else {
		envMap[GoArchEnvVar] = goarch
	}

	return envMap
}
