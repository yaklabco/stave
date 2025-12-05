package st

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"testing"
)

func TestDepsLogging(t *testing.T) {
	t.Setenv("STAVEFILE_VERBOSE", "1")
	buf := &bytes.Buffer{}

	defaultLogger := simpleConsoleLogger
	simpleConsoleLogger = log.New(buf, "", 0)
	defer func() { simpleConsoleLogger = defaultLogger }()

	foo()

	if strings.Count(buf.String(), "github.com/yaklabco/stave/pkg/st.baz") != 1 {
		t.Fatalf("expected one baz to be logged, but got\n%s", buf)
	}
}

func foo() {
	Deps(bar, baz)
}

func bar() {
	Deps(baz)
}

func baz() {}

func TestDepWasNotInvoked(t *testing.T) {
	fn1 := func() error {
		return nil
	}
	defer func() {
		err := recover()
		if err == nil {
			t.Fatal("expected panic, but didn't get one")
		}
		gotErr := fmt.Sprint(err)
		wantErr := "non-function used as a target dependency: <nil>. The st.Deps, st.SerialDeps and st.CtxDeps functions accept function names, such as st.Deps(TargetA, TargetB)"
		if !strings.Contains(gotErr, wantErr) {
			t.Fatalf(`expected to get "%s" but got "%s"`, wantErr, gotErr)
		}
	}()
	func(fns ...interface{}) {
		checkFns(fns)
	}(fn1())
}
