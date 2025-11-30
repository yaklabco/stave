package st_test

import (
	"fmt"

	"github.com/yaklabco/stave/st"
)

func Example() {
	// Deps will run each dependency exactly once, and will run leaf-dependencies before those
	// functions that depend on them (if you put st.Deps first in the function).

	// Normal (non-serial) Deps runs all dependencies in goroutines, so which one finishes first is
	// non-deterministic. Here we use SerialDeps here to ensure the example always produces the same
	// output.

	st.SerialDeps(st.F(Say, "hi"), Bark)
	// output:
	// hi
	// woof
}

func Say(something string) {
	fmt.Println(something)
}

func Bark() {
	st.Deps(st.F(Say, "woof"))
}
