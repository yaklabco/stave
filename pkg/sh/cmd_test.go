package sh

import (
	"bytes"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOutCmd(t *testing.T) {
	cmd := OutCmd(os.Args[0], "-printArgs", "foo", "bar")
	out, err := cmd("baz", "bat")
	if err != nil {
		t.Fatal(err)
	}
	expected := "[foo bar baz bat]"
	if out != expected {
		t.Fatalf("expected %q but got %q", expected, out)
	}
}

func TestExitCode(t *testing.T) {
	ran, err := Exec(nil, nil, nil, nil, os.Args[0], "-helper", "-exit", "99")
	if err == nil {
		t.Fatal("unexpected nil error from run")
	}
	if !ran {
		t.Errorf("ran returned as false, but should have been true")
	}
	code := ExitStatus(err)
	if code != 99 {
		t.Fatalf("expected exit status 99, but got %v", code)
	}
}

func TestEnv(t *testing.T) {
	env := "SOME_REALLY_LONG_STAVEFILE_SPECIFIC_THING"
	out := &bytes.Buffer{}
	ran, err := Exec(map[string]string{env: "foobar"}, nil, out, nil, os.Args[0], "-printVar", env)
	if err != nil {
		t.Fatalf("unexpected error from runner: %#v", err)
	}
	if !ran {
		t.Errorf("expected ran to be true but was false.")
	}
	if out.String() != "foobar\n" {
		t.Errorf("expected foobar, got %q", out)
	}
}

func TestNotRun(t *testing.T) {
	ran, err := Exec(nil, nil, nil, nil, "thiswontwork")
	if err == nil {
		t.Fatal("unexpected nil error")
	}
	if ran {
		t.Fatal("expected ran to be false but was true")
	}
}

func TestAutoExpand(t *testing.T) {
	if err := os.Setenv("STAVE_FOOBAR", "baz"); err != nil {
		t.Fatal(err)
	}
	s, err := Output("echo", "$STAVE_FOOBAR")
	if err != nil {
		t.Fatal(err)
	}
	if s != "baz" {
		t.Fatalf(`Expected "baz" but got %q`, s)
	}
}

func TestDryRunOutput(t *testing.T) {
	// Invoke test binary with helper flag to exercise dry-run Output path.
	cmd := exec.Command(os.Args[0], "-dryRunOutput")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("dry-run helper failed: %v, out=%s", err, string(out))
	}
	got := strings.TrimSpace(string(out))
	want := "DRYRUN: somecmd arg1 arg two"
	assert.Contains(t, got, want)
}

func TestPiper(t *testing.T) {
	t.Run("pipes stdin to stdout", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("cat not available on windows in CI")
		}
		in := bytes.NewBufferString("hello\nworld\n")
		var out bytes.Buffer
		var errBuf bytes.Buffer
		if err := Piper(in, &out, &errBuf, "cat"); err != nil {
			t.Fatalf("Piper(cat) failed: %v", err)
		}
		assert.Equal(t, "hello\nworld\n", out.String())
		assert.Empty(t, errBuf.String())
	})

	t.Run("captures stderr output", func(t *testing.T) {
		var out bytes.Buffer
		var errBuf bytes.Buffer
		// Use test helper to write to stderr deterministically
		err := Piper(nil, &out, &errBuf, os.Args[0], "-helper", "-stderr", "oops")
		require.NoError(t, err)
		assert.Equal(t, "oops\n", errBuf.String())
		// helper prints a newline to stdout by default (empty string + Fprintln)
		assert.Equal(t, "\n", out.String())
	})
}

func TestPiperWith(t *testing.T) {
	// Ensure env override is respected and stdout is captured
	const key = "STAVE_TEST_PIPERWITH_VAR"
	_ = os.Setenv(key, "outer")
	defer os.Unsetenv(key)

	var out bytes.Buffer
	var errBuf bytes.Buffer
	env := map[string]string{key: "inner"}
	if err := PiperWith(env, nil, &out, &errBuf, os.Args[0], "-printVar", key); err != nil {
		t.Fatalf("PiperWith failed: %v", err)
	}
	assert.Equal(t, "inner\n", out.String())
	assert.Empty(t, errBuf.String())
}
