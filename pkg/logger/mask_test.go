package logger

import "testing"

func TestMaskEmail(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"standard email", "user@example.com", "us***@example.com"},
		{"long local part", "firstname.lastname@corp.com", "fi***@corp.com"},
		{"short local part", "a@example.com", "a***@example.com"},
		{"two char local", "ab@example.com", "ab***@example.com"},
		{"empty string", "", ""},
		{"no at sign", "username", "us***"},
		{"single char no at", "u", "u***"},
		{"just at sign", "@domain.com", "***@domain.com"},
		{"multiple at signs", "user@sub@domain.com", "us***@domain.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MaskEmail(tt.input)
			if got != tt.want {
				t.Errorf("MaskEmail(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
