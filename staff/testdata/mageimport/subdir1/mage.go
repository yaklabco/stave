package staff

import (
	"fmt"

	"github.com/yaklabco/staff/mg"
)

// BuildSubdir Builds stuff.
func BuildSubdir() {
	fmt.Println("buildsubdir")
}

// NS is a namespace.
type NS mg.Namespace

// Deploy deploys stuff.
func (NS) Deploy() {
	fmt.Println("deploy")
}
