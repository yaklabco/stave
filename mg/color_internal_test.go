package mg

import (
	"testing"
)

func TestToLowerCase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		s    string
		want string
	}{
		{
			name: "all uppercase",
			s:    "ABC",
			want: "abc",
		},
		{
			name: "mixed case",
			s:    "aBc",
			want: "abc",
		},
		{
			name: "empty string",
			s:    "",
			want: "",
		},
		{
			name: "already lowercase",
			s:    "abc",
			want: "abc",
		},
		{
			name: "with numbers",
			s:    "ABC123",
			want: "abc123",
		},
		{
			name: "non-alpha chars unchanged",
			s:    "Hello-World_123!",
			want: "hello-world_123!",
		},
		{
			name: "unicode passthrough",
			s:    "Hello世界",
			want: "hello世界",
		},
		{
			name: "special chars",
			s:    "Test@#$%",
			want: "test@#$%",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := toLowerCase(tt.s)
			if got != tt.want {
				t.Errorf("toLowerCase(%q) = %q, want %q", tt.s, got, tt.want)
			}
		})
	}
}

func TestGetAnsiColor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		color     string
		wantCode  string
		wantFound bool
	}{
		// Test all 16 colors
		{"black", "Black", "\u001b[30m", true},
		{"red", "Red", "\u001b[31m", true},
		{"green", "Green", "\u001b[32m", true},
		{"yellow", "Yellow", "\u001b[33m", true},
		{"blue", "Blue", "\u001b[34m", true},
		{"magenta", "Magenta", "\u001b[35m", true},
		{"cyan", "Cyan", "\u001b[36m", true},
		{"white", "White", "\u001b[37m", true},
		{"brightblack", "BrightBlack", "\u001b[30;1m", true},
		{"brightred", "BrightRed", "\u001b[31;1m", true},
		{"brightgreen", "BrightGreen", "\u001b[32;1m", true},
		{"brightyellow", "BrightYellow", "\u001b[33;1m", true},
		{"brightblue", "BrightBlue", "\u001b[34;1m", true},
		{"brightmagenta", "BrightMagenta", "\u001b[35;1m", true},
		{"brightcyan", "BrightCyan", "\u001b[36;1m", true},
		{"brightwhite", "BrightWhite", "\u001b[37;1m", true},

		// Case insensitivity tests
		{"lowercase", "red", "\u001b[31m", true},
		{"uppercase", "RED", "\u001b[31m", true},
		{"mixed case", "rEd", "\u001b[31m", true},
		{"mixed case bright", "BrIgHtReD", "\u001b[31;1m", true},

		// Invalid color
		{"invalid color", "NotAColor", "", false},
		{"empty string", "", "", false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotCode, gotFound := getAnsiColor(tt.color)
			if gotFound != tt.wantFound {
				t.Errorf("getAnsiColor(%q) found = %v, want %v", tt.color, gotFound, tt.wantFound)
			}
			if tt.wantFound && gotCode != tt.wantCode {
				t.Errorf("getAnsiColor(%q) code = %q, want %q", tt.color, gotCode, tt.wantCode)
			}
		})
	}
}

func TestColorStringBoundaries(t *testing.T) {
	t.Parallel()

	// Test that all defined Color constants have valid String() output
	colors := []Color{
		Black, Red, Green, Yellow, Blue, Magenta, Cyan, White,
		BrightBlack, BrightRed, BrightGreen, BrightYellow,
		BrightBlue, BrightMagenta, BrightCyan, BrightWhite,
	}

	for _, c := range colors {
		t.Run(c.String(), func(t *testing.T) {
			s := c.String()
			if s == "" {
				t.Errorf("Color(%d).String() returned empty string", c)
			}
			// Verify it's a valid color name (should be lowercase by convention)
			if _, found := getAnsiColor(s); !found {
				t.Errorf("Color(%d).String() = %q, not found in getAnsiColor", c, s)
			}
		})
	}
}

