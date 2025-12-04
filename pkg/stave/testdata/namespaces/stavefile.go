//go:build stave
// +build stave

package main

import (
	"context"
	"fmt"

	"github.com/yaklabco/stave/pkg/st"
)

var Default = NS.Error

func TestNamespaceDep() {
	st.Deps(NS.Error, NS.Bare, NS.BareCtx, NS.CtxErr)
}

type NS st.Namespace

func (NS) Error() error {
	fmt.Println("hi!")
	return nil
}

func (NS) Bare() {
}

func (NS) BareCtx(ctx context.Context) {
}
func (NS) CtxErr(ctx context.Context) error {
	return nil
}
