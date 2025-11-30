package stave

import (
	"bytes"
	"testing"
)

func TestStaveImportsList(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/staveimport",
		Stdout: stdout,
		Stderr: stderr,
		List:   true,
	}

	code := Invoke(inv)
	if code != 0 {
		t.Fatalf("expected to exit with code 0, but got %v, stderr:\n%s", code, stderr)
	}
	actual := stdout.String()
	expected := `
Targets:
  buildSubdir        Builds stuff.
  ns:deploy          deploys stuff.
  root               
  zz:buildSubdir2    Builds stuff.
  zz:ns:deploy2*     deploys stuff.

* default target
`[1:]

	if actual != expected {
		t.Logf("expected: %q", expected)
		t.Logf("  actual: %q", actual)
		t.Fatalf("expected:\n%v\n\ngot:\n%v", expected, actual)
	}
}

func TestStaveImportsHelp(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/staveimport",
		Stdout: stdout,
		Stderr: stderr,
		Help:   true,
		Args:   []string{"buildSubdir"},
	}

	code := Invoke(inv)
	if code != 0 {
		t.Fatalf("expected to exit with code 0, but got %v, stderr:\n%s", code, stderr)
	}
	actual := stdout.String()
	expected := `
BuildSubdir Builds stuff.

Usage:

	stave buildsubdir

`[1:]

	if actual != expected {
		t.Logf("expected: %q", expected)
		t.Logf("  actual: %q", actual)
		t.Fatalf("expected:\n%v\n\ngot:\n%v", expected, actual)
	}
}

func TestStaveImportsHelpNamed(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/staveimport",
		Stdout: stdout,
		Stderr: stderr,
		Help:   true,
		Args:   []string{"zz:buildSubdir2"},
	}

	code := Invoke(inv)
	if code != 0 {
		t.Fatalf("expected to exit with code 0, but got %v, stderr:\n%s", code, stderr)
	}
	actual := stdout.String()
	expected := `
BuildSubdir2 Builds stuff.

Usage:

	stave zz:buildsubdir2

`[1:]

	if actual != expected {
		t.Logf("expected: %q", expected)
		t.Logf("  actual: %q", actual)
		t.Fatalf("expected:\n%v\n\ngot:\n%v", expected, actual)
	}
}

func TestStaveImportsHelpNamedNS(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/staveimport",
		Stdout: stdout,
		Stderr: stderr,
		Help:   true,
		Args:   []string{"zz:ns:deploy2"},
	}

	code := Invoke(inv)
	if code != 0 {
		t.Fatalf("expected to exit with code 0, but got %v, stderr:\n%s", code, stderr)
	}
	actual := stdout.String()
	expected := `
Deploy2 deploys stuff.

Usage:

	stave zz:ns:deploy2

Aliases: nsd2

`[1:]

	if actual != expected {
		t.Logf("expected: %q", expected)
		t.Logf("  actual: %q", actual)
		t.Fatalf("expected:\n%v\n\ngot:\n%v", expected, actual)
	}
}

func TestStaveImportsRoot(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/staveimport",
		Stdout: stdout,
		Stderr: stderr,
		Args:   []string{"root"},
	}

	code := Invoke(inv)
	if code != 0 {
		t.Fatalf("expected to exit with code 0, but got %v, stderr:\n%s", code, stderr)
	}
	actual := stdout.String()
	expected := "root\n"
	if actual != expected {
		t.Fatalf("expected: %q got: %q", expected, actual)
	}
}

func TestStaveImportsNamedNS(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/staveimport",
		Stdout: stdout,
		Stderr: stderr,
		Args:   []string{"zz:nS:deploy2"},
	}

	code := Invoke(inv)
	if code != 0 {
		t.Fatalf("expected to exit with code 0, but got %v, stderr:\n%s", code, stderr)
	}
	actual := stdout.String()
	expected := "deploy2\n"
	if actual != expected {
		t.Fatalf("expected: %q got: %q", expected, actual)
	}
}

func TestStaveImportsNamedRoot(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/staveimport",
		Stdout: stdout,
		Stderr: stderr,
		Args:   []string{"zz:buildSubdir2"},
	}

	code := Invoke(inv)
	if code != 0 {
		t.Fatalf("expected to exit with code 0, but got %v, stderr:\n%s", code, stderr)
	}
	actual := stdout.String()
	expected := "buildsubdir2\n"
	if actual != expected {
		t.Fatalf("expected: %q got: %q", expected, actual)
	}
	if stderr := stderr.String(); stderr != "" {
		t.Fatal("unexpected output to stderr: ", stderr)
	}
}

func TestStaveImportsRootImportNS(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/staveimport",
		Stdout: stdout,
		Stderr: stderr,
		Args:   []string{"nS:deploy"},
	}

	code := Invoke(inv)
	if code != 0 {
		t.Fatalf("expected to exit with code 0, but got %v, stderr:\n%s", code, stderr)
	}
	actual := stdout.String()
	expected := "deploy\n"
	if actual != expected {
		t.Fatalf("expected: %q got: %q", expected, actual)
	}
}

func TestStaveImportsRootImport(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/staveimport",
		Stdout: stdout,
		Stderr: stderr,
		Args:   []string{"buildSubdir"},
	}

	code := Invoke(inv)
	if code != 0 {
		t.Fatalf("expected to exit with code 0, but got %v, stderr:\n%s", code, stderr)
	}
	actual := stdout.String()
	expected := "buildsubdir\n"
	if actual != expected {
		t.Fatalf("expected: %q got: %q", expected, actual)
	}
}

func TestStaveImportsAliasToNS(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/staveimport",
		Stdout: stdout,
		Stderr: stderr,
		Args:   []string{"nsd2"},
	}

	code := Invoke(inv)
	if code != 0 {
		t.Fatalf("expected to exit with code 0, but got %v, stderr:\n%s", code, stderr)
	}
	actual := stdout.String()
	expected := "deploy2\n"
	if actual != expected {
		t.Fatalf("expected: %q got: %q", expected, actual)
	}
}

func TestStaveImportsOneLine(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/staveimport/oneline",
		Stdout: stdout,
		Stderr: stderr,
		Args:   []string{"build"},
	}

	code := Invoke(inv)
	if code != 0 {
		t.Fatalf("expected to exit with code 0, but got %v, stderr:\n%s", code, stderr)
	}
	actual := stdout.String()
	expected := "build\n"
	if actual != expected {
		t.Fatalf("expected: %q got: %q", expected, actual)
	}
}
func TestStaveImportsTrailing(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/staveimport/trailing",
		Stdout: stdout,
		Stderr: stderr,
		Args:   []string{"build"},
	}

	code := Invoke(inv)
	if code != 0 {
		t.Fatalf("expected to exit with code 0, but got %v, stderr:\n%s", code, stderr)
	}
	actual := stdout.String()
	expected := "build\n"
	if actual != expected {
		t.Fatalf("expected: %q got: %q", expected, actual)
	}
}

func TestStaveImportsTaggedPackage(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/staveimport/tagged",
		Stdout: stdout,
		Stderr: stderr,
		List:   true,
	}

	code := Invoke(inv)
	if code != 1 {
		t.Fatalf("expected to exit with code 1, but got %v, stdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}

	actual := stderr.String()
	// Match a shorter version of the error message, since the output from go list differs between versions
	expected := `
Error parsing stavefiles: error running "go list -f {{.Dir}}||{{.Name}} github.com/yaklabco/stave/stave/testdata/staveimport/tagged/pkg": exit status 1`[1:]
	actualShortened := actual[:len(expected)]
	if actualShortened != expected {
		t.Logf("expected: %q", expected)
		t.Logf("actual: %q", actualShortened)
		t.Fatalf("expected:\n%s\n\ngot:\n%s", expected, actualShortened)
	}
}

func TestStaveImportsSameNamespaceUniqueTargets(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/staveimport/samenamespace/uniquetargets",
		Stdout: stdout,
		Stderr: stderr,
		List:   true,
	}

	code := Invoke(inv)
	if code != 0 {
		t.Fatalf("expected to exit with code 0, but got %v, stderr:\n%s", code, stderr)
	}
	actual := stdout.String()
	expected := `
Targets:
  samenamespace:build1    
  samenamespace:build2    
`[1:]

	if actual != expected {
		t.Logf("expected: %q", expected)
		t.Logf("  actual: %q", actual)
		t.Fatalf("expected:\n%v\n\ngot:\n%v", expected, actual)
	}
}

func TestStaveImportsSameNamespaceDupTargets(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	inv := Invocation{
		Dir:    "./testdata/staveimport/samenamespace/duptargets",
		Stdout: stdout,
		Stderr: stderr,
		List:   true,
	}

	code := Invoke(inv)
	if code != 1 {
		t.Fatalf("expected to exit with code 1, but got %v, stderr:\n%s", code, stderr)
	}
	actual := stderr.String()
	expected := `
Error parsing stavefiles: "samenamespace:build" target has multiple definitions: github.com/yaklabco/stave/stave/testdata/staveimport/samenamespace/duptargets/package1.Build, github.com/yaklabco/stave/stave/testdata/staveimport/samenamespace/duptargets/package2.Build

`[1:]
	if actual != expected {
		t.Logf("expected: %q", expected)
		t.Logf("  actual: %q", actual)
		t.Fatalf("expected:\n%v\n\ngot:\n%v", expected, actual)
	}
}
