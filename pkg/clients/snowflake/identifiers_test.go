package snowflake

import "testing"

func TestNeedsQuoting(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{name: "starts with digit", input: "783m", expected: true},
		{name: "normal identifier", input: "john_doe", expected: false},
		{name: "starts with underscore", input: "_admin", expected: false},
		{name: "contains dollar sign", input: "user$1", expected: false},
		{name: "empty string", input: "", expected: true},
		{name: "starts with special char", input: "@user", expected: true},
		{name: "contains space", input: "john doe", expected: true},
		{name: "all digits", input: "12345", expected: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := needsQuoting(tt.input)
			if got != tt.expected {
				t.Errorf("needsQuoting(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestQuoteSnowflakeIdentifier(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		pathEscape bool
		expected   string
	}{
		{name: "digit-prefix no escape", input: "783m", pathEscape: false, expected: `"783M"`},
		{name: "digit-prefix path escape", input: "783m", pathEscape: true, expected: "%22783M%22"},
		{name: "all digits no escape", input: "12345", pathEscape: false, expected: `"12345"`},
		{name: "valid identifier unchanged", input: "john_doe", pathEscape: false, expected: "john_doe"},
		{name: "valid identifier path escape unchanged", input: "john_doe", pathEscape: true, expected: "john_doe"},
		{name: "special char uppercased", input: "@user", pathEscape: false, expected: `"@USER"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := quoteSnowflakeIdentifier(tt.input, tt.pathEscape)
			if got != tt.expected {
				t.Errorf("quoteSnowflakeIdentifier(%q, %v) = %q, want %q", tt.input, tt.pathEscape, got, tt.expected)
			}
		})
	}
}
