package internal

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"runtime"
	"strings"

	"github.com/yaklabco/stave/internal/dryrun"
	"github.com/yaklabco/stave/internal/env"
)

const (
	GoOSEnvVar   = "GOOS"
	GoArchEnvVar = "GOARCH"
)

func SetDebug(l *log.Logger) {
	debug = l
}

func RunDebug(ctx context.Context, cmd string, args ...string) error {
	envMap := EnvWithCurrentGOOS()

	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}

	debug.Println("running", cmd, strings.Join(args, " "))
	theCmd := dryrun.Wrap(ctx, cmd, args...)
	theCmd.Env = env.ToAssignments(envMap)
	theCmd.Stderr = errBuf
	theCmd.Stdout = outBuf

	if err := theCmd.Run(); err != nil {
		debug.Print("error running '", cmd, strings.Join(args, " "), "': ", err, ": ", errBuf)
		return err
	}

	debug.Println(outBuf)

	return nil
}

func OutputDebug(ctx context.Context, cmd string, args ...string) (string, error) {
	envMap := EnvWithCurrentGOOS()

	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}

	debug.Println("running", cmd, strings.Join(args, " "))
	theCmd := dryrun.Wrap(ctx, cmd, args...)
	theCmd.Env = env.ToAssignments(envMap)
	theCmd.Stderr = errBuf
	theCmd.Stdout = outBuf

	if err := theCmd.Run(); err != nil {
		errMsg := strings.TrimSpace(errBuf.String())
		debug.Print("error running '", cmd, strings.Join(args, " "), "': ", err, ": ", errMsg)
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
