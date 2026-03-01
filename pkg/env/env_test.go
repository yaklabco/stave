package env

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseBool(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    bool
		wantErr bool
	}{
		// True values
		{"1", "1", true, false},
		{"true lowercase", "true", true, false},
		{"true uppercase", "TRUE", true, false},
		{"true titlecase", "True", true, false},
		{"true mixedcase", "tRuE", true, false},
		{"yes lowercase", "yes", true, false},
		{"yes uppercase", "YES", true, false},
		{"yes titlecase", "Yes", true, false},
		{"yes mixedcase", "yEs", true, false},

		// False values
		{"0", "0", false, false},
		{"false lowercase", "false", false, false},
		{"false uppercase", "FALSE", false, false},
		{"false titlecase", "False", false, false},
		{"false mixedcase", "fAlSe", false, false},
		{"no lowercase", "no", false, false},
		{"no uppercase", "NO", false, false},
		{"no titlecase", "No", false, false},
		{"no mixedcase", "nO", false, false},

		// Empty string
		{"empty", "", false, false},

		// Whitespace handling
		{"true with spaces", "  true  ", true, false},
		{"false with spaces", "  false  ", false, false},
		{"yes with spaces", "  yes  ", true, false},
		{"no with spaces", "  no  ", false, false},
		{"1 with spaces", "  1  ", true, false},
		{"0 with spaces", "  0  ", false, false},
		{"true with tabs and newlines", "\ttrue\n", true, false},
		{"YES with tabs", "\tYES\n", true, false},

		// Invalid values
		{"t", "t", false, true},
		{"f", "f", false, true},
		{"T", "T", false, true},
		{"F", "F", false, true},
		{"on", "on", false, true},
		{"off", "off", false, true},
		{"enabled", "enabled", false, true},
		{"disabled", "disabled", false, true},
		{"invalid", "invalid", false, true},
		{"2", "2", false, true},
		{"-1", "-1", false, true},
		{"y", "y", false, true},
		{"n", "n", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseBool(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseBool(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if tt.wantErr && err != nil && !errors.Is(err, ErrInvalidBool) {
				t.Errorf("ParseBool(%q) error should wrap ErrInvalidBool, got %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("ParseBool(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseBoolEnv(t *testing.T) {
	const envVar = "TEST_PARSE_BOOL_ENV"

	tests := []struct {
		name    string
		value   string
		setEnv  bool
		want    bool
		wantErr bool
	}{
		{"unset", "", false, false, false},
		{"empty", "", true, false, false},
		{"true", "true", true, true, false},
		{"false", "false", true, false, false},
		{"invalid", "enabled", true, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				t.Setenv(envVar, tt.value)
			}

			got, err := ParseBoolEnv(envVar)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseBoolEnv(%q) error = %v, wantErr %v", envVar, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("ParseBoolEnv(%q) = %v, want %v", envVar, got, tt.want)
			}
		})
	}
}

func TestFailsafeParseBoolEnvDefaultFalse(t *testing.T) {
	const envVar = "TEST_FAILSAFE_PARSE_BOOL_ENV_DEFAULT_FALSE"

	tests := []struct {
		name   string
		value  string
		setEnv bool
		want   bool
	}{
		{"unset", "", false, false},
		{"empty", "", true, false},
		{"true", "true", true, true},
		{"TRUE", "TRUE", true, true},
		{"yes", "yes", true, true},
		{"YES", "YES", true, true},
		{"1", "1", true, true},
		{"false", "false", true, false},
		{"no", "no", true, false},
		{"0", "0", true, false},
		{"invalid", "enabled", true, false},
		{"whitespace", "  true  ", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				t.Setenv(envVar, tt.value)
			}

			if got := FailsafeParseBoolEnv(envVar, false); got != tt.want {
				t.Errorf("FailsafeParseBoolEnv(%q, false) = %v, want %v", envVar, got, tt.want)
			}
		})
	}
}

func TestFailsafeParseBoolEnvDefaultTrue(t *testing.T) {
	const envVar = "TEST_FAILSAFE_PARSE_BOOL_ENV_DEFAULT_TRUE"

	tests := []struct {
		name   string
		value  string
		setEnv bool
		want   bool
	}{
		{"unset", "", false, true},
		{"empty", "", true, true},
		{"true", "true", true, true},
		{"yes", "yes", true, true},
		{"1", "1", true, true},
		{"false", "false", true, false},
		{"no", "no", true, false},
		{"0", "0", true, false},
		{"invalid", "enabled", true, true},
		{"whitespace false", "  false  ", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				t.Setenv(envVar, tt.value)
			}

			if got := FailsafeParseBoolEnv(envVar, true); got != tt.want {
				t.Errorf("FailsafeParseBoolEnv(%q, true) = %v, want %v", envVar, got, tt.want)
			}
		})
	}
}

// clearCIEnv unsets all CI environment variables for test isolation.
// It uses t.Setenv to register cleanup (restoring original values),
// then os.Unsetenv to actually remove them from the environment.
func clearCIEnv(t *testing.T) {
	t.Helper()

	for _, v := range CIEnvVarNames() {
		t.Setenv(v, "")
		require.NoError(t, os.Unsetenv(v))
	}
}

func TestInCI(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want bool
	}{
		{
			name: "no CI env vars",
			env:  nil,
			want: false,
		},
		{
			name: "CI=true",
			env:  map[string]string{"CI": "true"},
			want: true,
		},
		{
			name: "CI=1",
			env:  map[string]string{"CI": "1"},
			want: true,
		},
		{
			name: "CI=false",
			env:  map[string]string{"CI": "false"},
			want: false,
		},
		{
			name: "CI empty",
			env:  map[string]string{"CI": ""},
			want: false,
		},
		{
			name: "GITHUB_ACTIONS=true",
			env:  map[string]string{"GITHUB_ACTIONS": "true"},
			want: true,
		},
		{
			name: "GITLAB_CI=true",
			env:  map[string]string{"GITLAB_CI": "true"},
			want: true,
		},
		{
			name: "JENKINS_URL set",
			env:  map[string]string{"JENKINS_URL": "http://jenkins.example.com"},
			want: true,
		},
		{
			name: "CIRCLECI=true",
			env:  map[string]string{"CIRCLECI": "true"},
			want: true,
		},
		{
			name: "BUILDKITE=true",
			env:  map[string]string{"BUILDKITE": "true"},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearCIEnv(t)

			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			assert.Equal(t, tt.want, InCI())
		})
	}
}
