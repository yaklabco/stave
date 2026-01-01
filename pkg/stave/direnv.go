package stave

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/yaklabco/direnv/v2/pkg/callable"
)

func delegateToDirEnv(ctx context.Context, params RunParams) error {
	env, err := setupEnv(params)
	if err != nil {
		return fmt.Errorf("failed to setup environment: %w", err)
	}

	// Temporarily replace os.Args[0] with the `--direnv`-supplied variant
	origArgsZero := os.Args[0]
	os.Args[0] += " --direnv --"
	defer func() {
		os.Args[0] = origArgsZero
	}()

	// Temporarily inhibit all logging
	origLogger := slog.Default()
	slog.SetDefault(slog.New(slog.DiscardHandler))
	defer func() {
		slog.SetDefault(origLogger)
	}()

	args := []string{os.Args[0]}
	args = append(args, params.Args...)
	if err := callable.CallableMain(ctx, args, env); err != nil {
		return fmt.Errorf("direnv run failed: %w", err)
	}

	return nil
}
