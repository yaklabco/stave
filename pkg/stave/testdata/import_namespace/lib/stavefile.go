//go:build stave

package lib

import (
	"fmt"
	"github.com/yaklabco/stave/pkg/st"
)

type NS st.Namespace

func (NS) Default() {
	fmt.Println("lib:NS:Default")
}

func (NS) Other() {
	fmt.Println("lib:NS:Other")
}
