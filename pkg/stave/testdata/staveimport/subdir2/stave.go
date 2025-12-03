package stave

import (
	"fmt"

	"github.com/yaklabco/stave/pkg/st"
)

// BuildSubdir2 Builds stuff.
func BuildSubdir2() {
	fmt.Println("buildsubdir2")
}

// NS is a namespace.
type NS st.Namespace

// Deploy2 deploys stuff.
func (NS) Deploy2() {
	fmt.Println("deploy2")
}
