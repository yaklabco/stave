package main

import (
	"context"

	"github.com/yaklabco/stave/stave"
	"os"
)

func main() {
	ctx := context.Background()
	os.Exit(stave.Main(ctx))
}
