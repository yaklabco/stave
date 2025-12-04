package deps

import (
	"fmt"

	"github.com/yaklabco/stave/pkg/st"
)

// All code in this package belongs to @na4ma4 in GitHub https://github.com/na4ma4/stavefile-test-import
// reproduced here for ease of testing regression on bug 508

type Docker st.Namespace

func (Docker) Test() {
	fmt.Println("docker")
}

func Test() {
	fmt.Println("test")
}
