package env

import (
	"errors"
	"testing"
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

func TestParseBoolEnvDefaultFalse(t *testing.T) {
	const envVar = "TEST_PARSE_BOOL_ENV_DEFAULT_FALSE"

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

			if got := ParseBoolEnvDefaultFalse(envVar); got != tt.want {
				t.Errorf("ParseBoolEnvDefaultFalse(%q) = %v, want %v", envVar, got, tt.want)
			}
		})
	}
}

func TestParseBoolEnvDefaultTrue(t *testing.T) {
	const envVar = "TEST_PARSE_BOOL_ENV_DEFAULT_TRUE"

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

			if got := ParseBoolEnvDefaultTrue(envVar); got != tt.want {
				t.Errorf("ParseBoolEnvDefaultTrue(%q) = %v, want %v", envVar, got, tt.want)
			}
		})
	}
}
