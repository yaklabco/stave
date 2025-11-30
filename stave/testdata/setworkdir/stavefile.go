//go:build stave
// +build stave

package main

import (
	"fmt"
	"os"
	"strings"
)

func TestWorkingDir() error {
	entries, err := os.ReadDir(".")
	if err != nil {
		return err
	}
	var out []string
	for _, entry := range entries {
		out = append(out, entry.Name())
	}

	fmt.Println(strings.Join(out, ", "))
	return nil
}

