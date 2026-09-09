package config

import "testing"

func TestParseEphemeralPortRange(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantMin uint16
		wantMax uint16
		wantErr bool
	}{
		{name: "range", value: "52000-52100", wantMin: 52000, wantMax: 52100},
		{name: "spaces", value: " 52000 - 52100 ", wantMin: 52000, wantMax: 52100},
		{name: "single port is invalid", value: "52000", wantErr: true},
		{name: "reversed range", value: "52100-52000", wantErr: true},
		{name: "zero minimum", value: "0-52000", wantErr: true},
		{name: "out of range", value: "52000-65536", wantErr: true},
		{name: "non numeric", value: "low-high", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			min, max, err := parseEphemeralPortRange(test.value)
			if (err != nil) != test.wantErr {
				t.Fatalf("parseEphemeralPortRange() error = %v, wantErr %v", err, test.wantErr)
			}
			if err == nil && (min != test.wantMin || max != test.wantMax) {
				t.Fatalf("parseEphemeralPortRange() = %d-%d, want %d-%d", min, max, test.wantMin, test.wantMax)
			}
		})
	}
}
