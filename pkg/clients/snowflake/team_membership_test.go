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
