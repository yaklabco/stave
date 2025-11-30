package sh

import (
	"bytes"
	"fmt"
	"os"
	"testing"
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
	ran, err := Exec(nil, nil, nil, os.Args[0], "-helper", "-exit", "99")
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
	env := "SOME_REALLY_LONG_MAGEFILE_SPECIFIC_THING"
	out := &bytes.Buffer{}
	ran, err := Exec(map[string]string{env: "foobar"}, out, nil, os.Args[0], "-printVar", env)
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
	ran, err := Exec(nil, nil, nil, "thiswontwork")
	if err == nil {
		t.Fatal("unexpected nil error")
	}
	if ran {
		t.Fatal("expected ran to be false but was true")
	}
}

func TestAutoExpand(t *testing.T) {
	if err := os.Setenv("MAGE_FOOBAR", "baz"); err != nil {
		t.Fatal(err)
	}
	s, err := Output("echo", "$MAGE_FOOBAR")
	if err != nil {
		t.Fatal(err)
	}
	if s != "baz" {
		t.Fatalf(`Expected "baz" but got %q`, s)
	}

}

func TestRun(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cmd     string
		args    []string
		wantErr bool
	}{
		{
			name:    "successful command",
			cmd:     "go",
			args:    []string{"version"},
			wantErr: false,
		},
		{
			name:    "failing command",
			cmd:     "go",
			args:    []string{"build", "nonexistent-package-xyz123"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := Run(tt.cmd, tt.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunV(t *testing.T) {
	// Don't run in parallel - modifies global os.Stdout

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		os.Stdout = oldStdout
	}()

	// Run a command
	err := RunV("go", "version")
	w.Close()

	if err != nil {
		t.Fatalf("RunV() error = %v", err)
	}

	// Read captured output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !bytes.Contains([]byte(output), []byte("go version")) {
		t.Errorf("RunV() did not write to stdout, got: %q", output)
	}
}

func TestRunWith(t *testing.T) {
	t.Parallel()

	env := map[string]string{"GOOS": "linux"}
	err := RunWith(env, "go", "version")
	if err != nil {
		t.Errorf("RunWith() error = %v", err)
	}
}

func TestRunWithV(t *testing.T) {
	// Don't run in parallel - modifies global os.Stdout

	env := map[string]string{"GOOS": "linux"}
	err := RunWithV(env, "go", "version")
	if err != nil {
		t.Errorf("RunWithV() error = %v", err)
	}
}

func TestRunCmd(t *testing.T) {
	t.Parallel()

	// Test curried args
	goVersion := RunCmd("go", "version")
	err := goVersion()
	if err != nil {
		t.Errorf("RunCmd()() error = %v", err)
	}

	// Test additional args appended
	goBuild := RunCmd("go", "build")
	err = goBuild("-n", ".")
	if err != nil {
		t.Errorf("RunCmd() with additional args error = %v", err)
	}
}

func TestOutput(t *testing.T) {
	t.Parallel()

	// Test capture stdout
	out, err := Output("go", "version")
	if err != nil {
		t.Fatalf("Output() error = %v", err)
	}
	if !bytes.Contains([]byte(out), []byte("go version")) {
		t.Errorf("Output() = %q, expected to contain 'go version'", out)
	}

	// Test newline trimmed
	if bytes.HasSuffix([]byte(out), []byte("\n")) {
		t.Errorf("Output() = %q, should not end with newline", out)
	}
}

func TestOutputWith(t *testing.T) {
	t.Parallel()

	env := map[string]string{"GOOS": "linux"}
	out, err := OutputWith(env, "go", "version")
	if err != nil {
		t.Fatalf("OutputWith() error = %v", err)
	}
	if !bytes.Contains([]byte(out), []byte("go version")) {
		t.Errorf("OutputWith() = %q, want to contain 'go version'", out)
	}
}

func TestCmdRan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error returns true",
			err:  nil,
			want: true,
		},
		{
			name: "regular error returns false",
			err:  fmt.Errorf("regular error"),
			want: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := CmdRan(tt.err)
			if got != tt.want {
				t.Errorf("CmdRan() = %v, want %v", got, tt.want)
			}
		})
	}
}
