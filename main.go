package main

import (
	"context"
	"os"

	"github.com/yaklabco/stave/stave"
)

func main() {
	ctx := context.Background()
	os.Exit(stave.Main(ctx))
}
