package mg

import (
	"os"
	"testing"
)

func TestGetEnvWithFallback(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		primary    string
		legacy     string
		primaryVal string
		legacyVal  string
		want       string
	}{
		{
			name:       "primary set",
			primary:    "TEST_PRIMARY",
			legacy:     "TEST_LEGACY",
			primaryVal: "primary-value",
			legacyVal:  "",
			want:       "primary-value",
		},
		{
			name:       "legacy set",
			primary:    "TEST_PRIMARY",
			legacy:     "TEST_LEGACY",
			primaryVal: "",
			legacyVal:  "legacy-value",
			want:       "legacy-value",
		},
		{
			name:       "both set primary wins",
			primary:    "TEST_PRIMARY",
			legacy:     "TEST_LEGACY",
			primaryVal: "primary-value",
			legacyVal:  "legacy-value",
			want:       "primary-value",
		},
		{
			name:       "neither set",
			primary:    "TEST_PRIMARY",
			legacy:     "TEST_LEGACY",
			primaryVal: "",
			legacyVal:  "",
			want:       "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// Don't run in parallel - env vars are global

			// Clean up env vars
			os.Unsetenv(tt.primary)
			os.Unsetenv(tt.legacy)
			defer func() {
				os.Unsetenv(tt.primary)
				os.Unsetenv(tt.legacy)
			}()

			if tt.primaryVal != "" {
				os.Setenv(tt.primary, tt.primaryVal)
			}
			if tt.legacyVal != "" {
				os.Setenv(tt.legacy, tt.legacyVal)
			}

			got := getEnvWithFallback(tt.primary, tt.legacy)
			if got != tt.want {
				t.Errorf("getEnvWithFallback(%q, %q) = %q, want %q", tt.primary, tt.legacy, got, tt.want)
			}
		})
	}
}

func TestGetBoolEnvWithFallback(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		primary    string
		legacy     string
		primaryVal string
		legacyVal  string
		want       bool
	}{
		// Primary env tests
		{"primary true", "TEST_P", "TEST_L", "true", "", true},
		{"primary false", "TEST_P", "TEST_L", "false", "", false},
		{"primary one", "TEST_P", "TEST_L", "1", "", true},
		{"primary zero", "TEST_P", "TEST_L", "0", "", false},
		{"primary TRUE", "TEST_P", "TEST_L", "TRUE", "", true},
		{"primary FALSE", "TEST_P", "TEST_L", "FALSE", "", false},

		// Legacy fallback tests
		{"legacy true", "TEST_P", "TEST_L", "", "true", true},
		{"legacy false", "TEST_P", "TEST_L", "", "false", false},
		{"legacy 1", "TEST_P", "TEST_L", "", "1", true},
		{"legacy 0", "TEST_P", "TEST_L", "", "0", false},

		// Primary takes precedence
		{"primary wins over legacy true", "TEST_P", "TEST_L", "0", "1", false},
		{"primary wins over legacy false", "TEST_P", "TEST_L", "1", "0", true},

		// Neither set or invalid
		{"neither set", "TEST_P", "TEST_L", "", "", false},
		{"invalid value", "TEST_P", "TEST_L", "invalid", "", false},
		{"empty string", "TEST_P", "TEST_L", "", "", false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// Don't run in parallel - env vars are global

			// Clean up env vars
			os.Unsetenv(tt.primary)
			os.Unsetenv(tt.legacy)
			defer func() {
				os.Unsetenv(tt.primary)
				os.Unsetenv(tt.legacy)
			}()

			if tt.primaryVal != "" {
				os.Setenv(tt.primary, tt.primaryVal)
			}
			if tt.legacyVal != "" {
				os.Setenv(tt.legacy, tt.legacyVal)
			}

			got := getBoolEnvWithFallback(tt.primary, tt.legacy)
			if got != tt.want {
				t.Errorf("getBoolEnvWithFallback(%q, %q) = %v, want %v", tt.primary, tt.legacy, got, tt.want)
			}
		})
	}
}

