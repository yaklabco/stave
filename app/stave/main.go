package main

import (
	"context"
	"fmt"
	"os"

	"github.com/yaklabco/stave/cmd/stave"
)

func main() {
	os.Exit(actualMain())
}

func actualMain() int {
	ctx := context.Background()

	rootCmd := stave.NewRootCmd(ctx)

	if err := stave.ExecuteWithFang(ctx, rootCmd); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		return 1
	}

	return 0
}
