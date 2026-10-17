package utils

import (
	"strings"
	"testing"
)

func TestSafeID(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"plain", "john.doe_1-2", "john.doe_1-2"},
		{"path traversal", "x/../../logout?", "x-..-..-logout"},
		{"symbol suffix", "victim-abc12#", "victim-abc12"},
		{"spaces", "John Doe  Jr", "John-Doe-Jr"},
		{"all symbols", "!@#$%^&*()", "user"},
		{"only dots", "..", "user"},
		{"empty", "", "user"},
		{"long", strings.Repeat("a", 70), strings.Repeat("a", 64)},
		{"long cut on dash", strings.Repeat("a", 63) + "/" + strings.Repeat("b", 10), strings.Repeat("a", 63)},
		{"unicode", "Zoë Müller", "Zoë-Müller"},
		{"unicode only", "山田太郎", "山田太郎"},
		{"long unicode cut by characters", strings.Repeat("ä", 70), strings.Repeat("ä", 64)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SafeID(tt.in); got != tt.want {
				t.Errorf("SafeID(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
