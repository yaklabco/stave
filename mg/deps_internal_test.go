package mg

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"testing"
)

func TestDepsLogging(t *testing.T) {
	os.Setenv("MAGEFILE_VERBOSE", "1")
	defer os.Unsetenv("MAGEFILE_VERBOSE")
	buf := &bytes.Buffer{}

	defaultLogger := logger
	logger = log.New(buf, "", 0)
	defer func() { logger = defaultLogger }()

	foo()

	if strings.Count(buf.String(), "Running dependency: github.com/yaklabco/stave/mg.baz") != 1 {
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
		wantErr := "non-function used as a target dependency: <nil>. The mg.Deps, mg.SerialDeps and mg.CtxDeps functions accept function names, such as mg.Deps(TargetA, TargetB)"
		if !strings.Contains(gotErr, wantErr) {
			t.Fatalf(`expected to get "%s" but got "%s"`, wantErr, gotErr)
		}
	}()
	func(fns ...interface{}) {
		checkFns(fns)
	}(fn1())
}

func TestChangeExit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		old  int
		new  int
		want int
	}{
		{"both zero", 0, 0, 0},
		{"old zero new nonzero", 0, 1, 1},
		{"old nonzero new zero", 1, 0, 1},
		{"same nonzero", 1, 1, 1},
		{"different nonzero", 1, 2, 1},
		{"preserves code", 99, 99, 99},
		{"new overwrites zero", 0, 42, 42},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := changeExit(tt.old, tt.new)
			if got != tt.want {
				t.Errorf("changeExit(%d, %d) = %d, want %d", tt.old, tt.new, got, tt.want)
			}
		})
	}
}

func TestDisplayName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		funcName string
		want     string
	}{
		{
			name:     "main package function",
			funcName: "main.Foo",
			want:     "Foo",
		},
		{
			name:     "qualified package function",
			funcName: "pkg.Foo",
			want:     "pkg.Foo",
		},
		{
			name:     "fully qualified path",
			funcName: "github.com/user/repo/pkg.Foo",
			want:     "github.com/user/repo/pkg.Foo",
		},
		{
			name:     "nested package",
			funcName: "a/b/c.Foo",
			want:     "a/b/c.Foo",
		},
		{
			name:     "single name",
			funcName: "Foo",
			want:     "Foo",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := displayName(tt.funcName)
			if got != tt.want {
				t.Errorf("displayName(%q) = %q, want %q", tt.funcName, got, tt.want)
			}
		})
	}
}

func TestFuncName(t *testing.T) {
	t.Parallel()

	// Test with an actual function
	testFunc := func() {}
	name := funcName(testFunc)

	// The name should contain the package and function identifier
	if !strings.Contains(name, "mg") {
		t.Errorf("funcName() = %q, expected to contain 'mg'", name)
	}

	// Verify it returns the same name when called twice
	name2 := funcName(testFunc)
	if name != name2 {
		t.Errorf("funcName() not deterministic: first=%q, second=%q", name, name2)
	}
}

func TestFuncNameWithKnownFunc(t *testing.T) {
	t.Parallel()

	// Use a known function from this package
	name := funcName(baz)

	// Should contain the full package path and function name
	expected := "github.com/yaklabco/stave/mg.baz"
	if name != expected {
		t.Errorf("funcName(baz) = %q, want %q", name, expected)
	}
}

// Helper function to get the current function name for testing
func getCurrentFuncName() string {
	pc, _, _, _ := runtime.Caller(0)
	return runtime.FuncForPC(pc).Name()
}
