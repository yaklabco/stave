//go:build stave
// +build stave

package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/yaklabco/stave/st"
)

// Returns a non-nil error.
func TakesContextNoError(ctx context.Context) {
	deadline, _ := ctx.Deadline()
	fmt.Printf("Context timeout: %v\n", deadline)
}

func Timeout(ctx context.Context) {
	time.Sleep(200 * time.Millisecond)
}

func TakesContextWithError(ctx context.Context) error {
	return errors.New("Something went sideways")
}

func CtxDeps(ctx context.Context) {
	st.CtxDeps(ctx, TakesContextNoError)
}
