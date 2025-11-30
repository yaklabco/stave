package mg

import (
	"errors"
	"testing"
)

func TestFatalExit(t *testing.T) {
	expected := 99
	code := ExitStatus(Fatal(expected))
	if code != expected {
		t.Fatalf("Expected code %v but got %v", expected, code)
	}
}

func TestFatalfExit(t *testing.T) {
	expected := 99
	code := ExitStatus(Fatalf(expected, "boo!"))
	if code != expected {
		t.Fatalf("Expected code %v but got %v", expected, code)
	}
}

func TestFatalMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		code    int
		args    []interface{}
		wantMsg string
	}{
		{
			name:    "simple message",
			code:    1,
			args:    []interface{}{"test error"},
			wantMsg: "test error",
		},
		{
			name:    "multiple args",
			code:    2,
			args:    []interface{}{"error:", " ", "details"},
			wantMsg: "error: details",
		},
		{
			name:    "with numbers",
			code:    3,
			args:    []interface{}{"code ", 42, " failed"},
			wantMsg: "code 42 failed",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := Fatal(tt.code, tt.args...)
			if err.Error() != tt.wantMsg {
				t.Errorf("Fatal().Error() = %q, want %q", err.Error(), tt.wantMsg)
			}

			// Verify exit status
			if code := ExitStatus(err); code != tt.code {
				t.Errorf("ExitStatus(Fatal(%d, ...)) = %d, want %d", tt.code, code, tt.code)
			}
		})
	}
}

func TestFatalfFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		code    int
		format  string
		args    []interface{}
		wantMsg string
	}{
		{
			name:    "string format",
			code:    1,
			format:  "error: %s",
			args:    []interface{}{"test"},
			wantMsg: "error: test",
		},
		{
			name:    "int format",
			code:    2,
			format:  "code %d failed",
			args:    []interface{}{42},
			wantMsg: "code 42 failed",
		},
		{
			name:    "multiple formats",
			code:    3,
			format:  "%s: %d - %v",
			args:    []interface{}{"error", 123, true},
			wantMsg: "error: 123 - true",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := Fatalf(tt.code, tt.format, tt.args...)
			if err.Error() != tt.wantMsg {
				t.Errorf("Fatalf().Error() = %q, want %q", err.Error(), tt.wantMsg)
			}

			// Verify exit status
			if code := ExitStatus(err); code != tt.code {
				t.Errorf("ExitStatus(Fatalf(%d, ...)) = %d, want %d", tt.code, code, tt.code)
			}
		})
	}
}

func TestExitStatusNil(t *testing.T) {
	t.Parallel()

	code := ExitStatus(nil)
	if code != 0 {
		t.Errorf("ExitStatus(nil) = %d, want 0", code)
	}
}

func TestExitStatusNonExit(t *testing.T) {
	t.Parallel()

	// Regular errors should return 1
	err := errors.New("regular error")
	code := ExitStatus(err)
	if code != 1 {
		t.Errorf("ExitStatus(errors.New(...)) = %d, want 1", code)
	}
}

func TestFatalErrExitStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		code int
	}{
		{"code 0", 0},
		{"code 1", 1},
		{"code 99", 99},
		{"code 255", 255},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := Fatal(tt.code, "test")
			if fe, ok := err.(interface{ ExitStatus() int }); ok {
				if got := fe.ExitStatus(); got != tt.code {
					t.Errorf("fatalErr.ExitStatus() = %d, want %d", got, tt.code)
				}
			} else {
				t.Error("Fatal() does not implement ExitStatus() int")
			}
		})
	}
}
