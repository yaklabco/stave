//go:build ignore
// +build ignore

package main

import (
	"context"
	"os"

	"github.com/yaklabco/stave/pkg/stave"
)

func main() {
	os.Exit(stave.Main(context.Background()))
}
