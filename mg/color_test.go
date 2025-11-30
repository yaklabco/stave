package mg

import (
	"os"
	"testing"

	"github.com/yaklabco/stave/config"
)

func TestValidTargetColor(t *testing.T) {
	// Reset config state
	os.Unsetenv(EnableColorEnv)
	os.Unsetenv(TargetColorEnv)
	os.Unsetenv(LegacyEnableColorEnv)
	os.Unsetenv(LegacyTargetColorEnv)
	config.ResetGlobal()
	t.Cleanup(func() {
		os.Unsetenv(EnableColorEnv)
		os.Unsetenv(TargetColorEnv)
		config.ResetGlobal()
	})

	os.Setenv(EnableColorEnv, "true")
	os.Setenv(TargetColorEnv, "Yellow")
	config.ResetGlobal()

	expected := "\u001b[33m"
	if actual := TargetColor(); actual != expected {
		t.Fatalf("expected %v but got %s", expected, actual)
	}
}

func TestValidTargetColorCaseInsensitive(t *testing.T) {
	// Reset config state
	os.Unsetenv(EnableColorEnv)
	os.Unsetenv(TargetColorEnv)
	os.Unsetenv(LegacyEnableColorEnv)
	os.Unsetenv(LegacyTargetColorEnv)
	config.ResetGlobal()
	t.Cleanup(func() {
		os.Unsetenv(EnableColorEnv)
		os.Unsetenv(TargetColorEnv)
		config.ResetGlobal()
	})

	os.Setenv(EnableColorEnv, "true")
	os.Setenv(TargetColorEnv, "rED")
	config.ResetGlobal()

	expected := "\u001b[31m"
	if actual := TargetColor(); actual != expected {
		t.Fatalf("expected %v but got %s", expected, actual)
	}
}

func TestInvalidTargetColor(t *testing.T) {
	// Reset config state
	os.Unsetenv(EnableColorEnv)
	os.Unsetenv(TargetColorEnv)
	os.Unsetenv(LegacyEnableColorEnv)
	os.Unsetenv(LegacyTargetColorEnv)
	config.ResetGlobal()
	t.Cleanup(func() {
		os.Unsetenv(EnableColorEnv)
		os.Unsetenv(TargetColorEnv)
		config.ResetGlobal()
	})

	os.Setenv(EnableColorEnv, "true")
	// NOTE: Brown is not a defined Color constant
	os.Setenv(TargetColorEnv, "Brown")
	config.ResetGlobal()

	expected := DefaultTargetAnsiColor
	if actual := TargetColor(); actual != expected {
		t.Fatalf("expected %v but got %s", expected, actual)
	}
}
