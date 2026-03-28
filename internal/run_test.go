package internal

import (
	"runtime"
	"testing"
)

func TestEnvWithCurrentGOOS(t *testing.T) {
	env := EnvWithCurrentGOOS()
	var foundGOOS, foundGOARCH bool
	for key, value := range env {
		switch key {
		case "GOOS":
			foundGOOS = true
			if value != runtime.GOOS {
				t.Errorf("expected GOOS=%s, got %s", runtime.GOOS, value)
			}
		case "GOARCH":
			foundGOARCH = true
			if value != runtime.GOARCH {
				t.Errorf("expected GOARCH=%s, got %s", runtime.GOARCH, value)
			}
		default:
			// ignore other env vars
			continue
		}
	}
	if !foundGOOS {
		t.Error("GOOS not found in env")
	}
	if !foundGOARCH {
		t.Error("GOARCH not found in env")
	}
}

func TestRunDebug(t *testing.T) {
	ctx := t.Context()

	// Test successful command
	err := RunDebug(ctx, "echo", "hello")
	if err != nil {
		t.Fatalf("RunDebug with valid command failed: %v", err)
	}

	// Test failed command
	err = RunDebug(ctx, "false")
	if err == nil {
		t.Fatal("RunDebug with failing command should return error")
	}
}

func TestOutputDebug(t *testing.T) {
	ctx := t.Context()

	// Test successful command
	out, err := OutputDebug(ctx, "echo", "hello")
	if err != nil {
		t.Fatalf("OutputDebug with valid command failed: %v", err)
	}
	if out != "hello" {
		t.Errorf("expected 'hello', got %q", out)
	}

	// Test failed command
	_, err = OutputDebug(ctx, "false")
	if err == nil {
		t.Fatal("OutputDebug with failing command should return error")
	}
}
