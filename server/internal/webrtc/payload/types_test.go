package payload

import "testing"

func TestValidateLength(t *testing.T) {
	tests := []struct {
		name    string
		event   uint8
		length  uint16
		wantErr bool
	}{
		{name: "move", event: OP_MOVE, length: 4},
		{name: "scroll", event: OP_SCROLL, length: 5},
		{name: "key", event: OP_KEY_DOWN, length: 4},
		{name: "ping", event: OP_PING, length: 8},
		{name: "touch", event: OP_TOUCH_BEGIN, length: 13},
		{name: "bad move", event: OP_MOVE, length: 5, wantErr: true},
		{name: "bad scroll", event: OP_SCROLL, length: 4, wantErr: true},
		{name: "bad key", event: OP_KEY_UP, length: 8, wantErr: true},
		{name: "unknown is forward compatible", event: 0xff, length: 1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateLength(test.event, test.length)
			if (err != nil) != test.wantErr {
				t.Fatalf("ValidateLength() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}
