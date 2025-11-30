package stave

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yaklabco/stave/config"
)

func TestRunConfigCommand_Show(t *testing.T) {
	// Reset global config state
	config.ResetGlobal()

	var stdout, stderr bytes.Buffer

	// Run 'stave config' (which defaults to 'show')
	exitCode := RunConfigCommand(&stdout, &stderr, []string{})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. stderr: %s", exitCode, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "go_cmd:") {
		t.Errorf("Expected output to contain 'go_cmd:', got: %s", output)
	}
	if !strings.Contains(output, "verbose:") {
		t.Errorf("Expected output to contain 'verbose:', got: %s", output)
	}
}

func TestRunConfigCommand_ShowExplicit(t *testing.T) {
	// Reset global config state
	config.ResetGlobal()

	var stdout, stderr bytes.Buffer

	// Run 'stave config show'
	exitCode := RunConfigCommand(&stdout, &stderr, []string{"show"})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. stderr: %s", exitCode, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "Effective Stave Configuration") {
		t.Errorf("Expected output to contain header, got: %s", output)
	}
}

func TestRunConfigCommand_Path(t *testing.T) {
	// Reset global config state
	config.ResetGlobal()

	var stdout, stderr bytes.Buffer

	// Run 'stave config path'
	exitCode := RunConfigCommand(&stdout, &stderr, []string{"path"})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. stderr: %s", exitCode, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "Configuration Paths:") {
		t.Errorf("Expected output to contain 'Configuration Paths:', got: %s", output)
	}
	if !strings.Contains(output, "User config:") {
		t.Errorf("Expected output to contain 'User config:', got: %s", output)
	}
}

func TestRunConfigCommand_Init(t *testing.T) {
	// Reset global config state
	config.ResetGlobal()

	// Use a temp directory for XDG_CONFIG_HOME
	tmpDir := t.TempDir()
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer func() {
		if origXDG == "" {
			os.Unsetenv("XDG_CONFIG_HOME")
		} else {
			os.Setenv("XDG_CONFIG_HOME", origXDG)
		}
	}()

	var stdout, stderr bytes.Buffer

	// Run 'stave config init'
	exitCode := RunConfigCommand(&stdout, &stderr, []string{"init"})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. stderr: %s", exitCode, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "Created config file:") {
		t.Errorf("Expected output to contain 'Created config file:', got: %s", output)
	}

	// Verify file was created
	configPath := filepath.Join(tmpDir, "stave", "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("Expected config file to be created at %s", configPath)
	}
}

func TestRunConfigCommand_InitAlreadyExists(t *testing.T) {
	// Reset global config state
	config.ResetGlobal()

	// Use a temp directory for XDG_CONFIG_HOME
	tmpDir := t.TempDir()
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer func() {
		if origXDG == "" {
			os.Unsetenv("XDG_CONFIG_HOME")
		} else {
			os.Setenv("XDG_CONFIG_HOME", origXDG)
		}
	}()

	// Create the config file first
	configDir := filepath.Join(tmpDir, "stave")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("test"), 0o644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	var stdout, stderr bytes.Buffer

	// Run 'stave config init' - should fail because file exists
	exitCode := RunConfigCommand(&stdout, &stderr, []string{"init"})

	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}

	errOutput := stderr.String()
	if !strings.Contains(errOutput, "already exists") {
		t.Errorf("Expected error about existing file, got: %s", errOutput)
	}
}

func TestRunConfigCommand_UnknownSubcommand(t *testing.T) {
	var stdout, stderr bytes.Buffer

	// Run 'stave config unknown'
	exitCode := RunConfigCommand(&stdout, &stderr, []string{"unknown"})

	if exitCode != 2 {
		t.Errorf("Expected exit code 2, got %d", exitCode)
	}

	errOutput := stderr.String()
	if !strings.Contains(errOutput, "unknown config subcommand") {
		t.Errorf("Expected error about unknown subcommand, got: %s", errOutput)
	}
}

func TestRunConfigCommand_Help(t *testing.T) {
	var stdout, stderr bytes.Buffer

	// Run 'stave config -h'
	exitCode := RunConfigCommand(&stdout, &stderr, []string{"-h"})

	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for help, got %d", exitCode)
	}

	output := stdout.String()
	if !strings.Contains(output, "stave config") {
		t.Errorf("Expected help output, got: %s", output)
	}
}
