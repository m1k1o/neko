package chat

import (
	"strings"
	"testing"
)

func TestCheckLength(t *testing.T) {
	tests := []struct {
		name string
		max  int
		text string
		ok   bool
	}{
		{"under", 5, "abcd", true},
		{"at", 5, "abcde", true},
		{"over", 5, "abcdef", false},
		{"empty", 5, "", true},
		{"multi-byte at", 5, "ééééé", true},
		{"multi-byte over", 5, "éééééé", false},
		{"emoji counts as one", 1, "😀", true},
		{"zero means unlimited", 0, strings.Repeat("a", 100000), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Manager{config: &Config{MaxLength: tt.max}}
			if err := m.checkLength(tt.text); (err == nil) != tt.ok {
				t.Errorf("checkLength(%d, %q) = %v, want ok=%v", tt.max, tt.text, err, tt.ok)
			}
		})
	}
}
