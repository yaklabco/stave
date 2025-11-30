package internal

import (
	"reflect"
	"testing"
)

func TestJoinEnv(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		env  map[string]string
		want int // number of expected entries
	}{
		{
			name: "empty map",
			env:  map[string]string{},
			want: 0,
		},
		{
			name: "single entry",
			env:  map[string]string{"KEY": "value"},
			want: 1,
		},
		{
			name: "multiple entries",
			env: map[string]string{
				"PATH":  "/usr/bin",
				"HOME":  "/home/user",
				"SHELL": "/bin/bash",
			},
			want: 3,
		},
		{
			name: "entry with equals in value",
			env:  map[string]string{"KEY": "a=b=c"},
			want: 1,
		},
		{
			name: "empty value",
			env:  map[string]string{"KEY": ""},
			want: 1,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := joinEnv(tt.env)
			if len(got) != tt.want {
				t.Errorf("joinEnv() returned %d entries, want %d", len(got), tt.want)
			}

			// Verify all entries are in KEY=VALUE format
			for _, entry := range got {
				if !contains(entry, "=") {
					t.Errorf("joinEnv() entry %q does not contain '='", entry)
				}
			}

			// Verify we can split it back correctly
			if len(got) > 0 {
				parsed, err := SplitEnv(got)
				if err != nil {
					t.Errorf("joinEnv() produced entries that can't be split: %v", err)
				}
				if !reflect.DeepEqual(parsed, tt.env) {
					t.Errorf("joinEnv() -> SplitEnv() roundtrip failed: got %v, want %v", parsed, tt.env)
				}
			}
		})
	}
}

// contains is a simple helper to check if a string contains a substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

