package oauth

import "testing"

func TestClaimFalse(t *testing.T) {
	cases := []struct {
		name   string
		claims map[string]any
		want   bool
	}{
		{"absent", map[string]any{}, false},
		{"true", map[string]any{"email_verified": true}, false},
		{"false", map[string]any{"email_verified": false}, true},
		{"string true", map[string]any{"email_verified": "true"}, false},
		{"string false", map[string]any{"email_verified": "False"}, true},
	}
	for _, tc := range cases {
		if got := claimFalse(tc.claims, "email_verified"); got != tc.want {
			t.Errorf("%s: got %v, want %v", tc.name, got, tc.want)
		}
	}
}
