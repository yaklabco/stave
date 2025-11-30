package internal_test

import (
	"os"
	"runtime"
	"testing"

	"github.com/yaklabco/stave/internal"
)

func TestSplitEnv(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		env     []string
		want    map[string]string
		wantErr bool
	}{
		{
			name: "valid key=value",
			env:  []string{"KEY=value", "FOO=bar"},
			want: map[string]string{"KEY": "value", "FOO": "bar"},
		},
		{
			name: "empty value",
			env:  []string{"KEY="},
			want: map[string]string{"KEY": ""},
		},
		{
			name: "multiple equals signs",
			env:  []string{"KEY=a=b=c"},
			want: map[string]string{"KEY": "a=b=c"},
		},
		{
			name:    "malformed no equals",
			env:     []string{"NOEQUALS"},
			wantErr: true,
		},
		{
			name: "empty slice",
			env:  []string{},
			want: map[string]string{},
		},
		{
			name: "mixed valid and spaces",
			env:  []string{"PATH=/usr/bin", "HOME=/home/user", "TERM=xterm-256color"},
			want: map[string]string{"PATH": "/usr/bin", "HOME": "/home/user", "TERM": "xterm-256color"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := internal.SplitEnv(tt.env)
			if (err != nil) != tt.wantErr {
				t.Errorf("SplitEnv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("SplitEnv() got %d entries, want %d", len(got), len(tt.want))
			}
			for k, v := range tt.want {
				if gotV, ok := got[k]; !ok {
					t.Errorf("SplitEnv() missing key %q", k)
				} else if gotV != v {
					t.Errorf("SplitEnv() key %q = %q, want %q", k, gotV, v)
				}
			}
		})
	}
}

func TestEnvWithCurrentGOOS(t *testing.T) {
	t.Parallel()

	env, err := internal.EnvWithCurrentGOOS()
	if err != nil {
		t.Fatalf("EnvWithCurrentGOOS() error = %v", err)
	}

	// Parse the returned environment
	parsed, err := internal.SplitEnv(env)
	if err != nil {
		t.Fatalf("failed to parse returned env: %v", err)
	}

	// Verify GOOS is set correctly
	if got := parsed["GOOS"]; got != runtime.GOOS {
		t.Errorf("GOOS = %q, want %q", got, runtime.GOOS)
	}

	// Verify GOARCH is set correctly
	if got := parsed["GOARCH"]; got != runtime.GOARCH {
		t.Errorf("GOARCH = %q, want %q", got, runtime.GOARCH)
	}

	// Verify we have some environment variables (at least PATH or HOME)
	if len(parsed) < 2 {
		t.Errorf("EnvWithCurrentGOOS() returned only %d env vars, expected more", len(parsed))
	}
}

func TestEnvWithGOOS(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		goos      string
		goarch    string
		wantGOOS  string
		wantArch  string
	}{
		{
			name:     "both empty use runtime",
			goos:     "",
			goarch:   "",
			wantGOOS: runtime.GOOS,
			wantArch: runtime.GOARCH,
		},
		{
			name:     "goos set goarch empty",
			goos:     "linux",
			goarch:   "",
			wantGOOS: "linux",
			wantArch: runtime.GOARCH,
		},
		{
			name:     "goos empty goarch set",
			goos:     "",
			goarch:   "arm64",
			wantGOOS: runtime.GOOS,
			wantArch: "arm64",
		},
		{
			name:     "both set",
			goos:     "windows",
			goarch:   "amd64",
			wantGOOS: "windows",
			wantArch: "amd64",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			env, err := internal.EnvWithGOOS(tt.goos, tt.goarch)
			if err != nil {
				t.Fatalf("EnvWithGOOS() error = %v", err)
			}

			parsed, err := internal.SplitEnv(env)
			if err != nil {
				t.Fatalf("failed to parse returned env: %v", err)
			}

			if got := parsed["GOOS"]; got != tt.wantGOOS {
				t.Errorf("GOOS = %q, want %q", got, tt.wantGOOS)
			}

			if got := parsed["GOARCH"]; got != tt.wantArch {
				t.Errorf("GOARCH = %q, want %q", got, tt.wantArch)
			}
		})
	}
}

// Note: TestRunDebug and TestOutputDebug are not included because they require
// debug mode to be enabled, which involves setting up a logger. These functions
// are tested indirectly through integration tests.

func TestSetDebug(t *testing.T) {
	// Save and restore os.Stderr
	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = w
	defer func() {
		os.Stderr = oldStderr
	}()

	// Create a logger that writes to stderr
	logger := internal.SetDebug
	logger(nil) // This just tests it doesn't panic

	// Restore
	w.Close()
	r.Close()
}

