package media

import "testing"

func TestDecodeControl(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		kind byte
		ok   bool
	}{
		{"ack", EncodeAck(Ack{Sequence: 9}), ControlAck, true},
		{"keyframe request", EncodeControl(ControlKeyframeRequest, 0), ControlKeyframeRequest, true},
		{"short message", []byte{1, 1, 0}, 0, false},
		{"empty message", nil, 0, false},
		{"wrong version", []byte{2, 1, 0, 0, 0, 0, 0, 1}, 0, false},
		{"unknown kind", []byte{1, 99, 0, 0, 0, 0, 0, 1}, 0, false},
		// A sample header must never be mistaken for a control message.
		{"sample header", make([]byte, SampleHeaderSize), 0, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			control, err := DecodeControl(test.data)
			if test.ok != (err == nil) {
				t.Fatalf("error = %v, want ok = %v", err, test.ok)
			}
			if test.ok && control.Kind != test.kind {
				t.Fatalf("kind = %d, want %d", control.Kind, test.kind)
			}
		})
	}
}

func TestDecodeAckRejectsOtherControls(t *testing.T) {
	if _, err := DecodeAck(EncodeControl(ControlKeyframeRequest, 0)); err == nil {
		t.Fatal("a keyframe request was decoded as an ack")
	}
	ack, err := DecodeAck(EncodeAck(Ack{Sequence: 77}))
	if err != nil || ack.Sequence != 77 {
		t.Fatalf("ack = %+v, err = %v", ack, err)
	}
}
