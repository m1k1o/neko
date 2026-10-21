package media

import (
	"encoding/binary"
	"errors"
)

// Client to server control messages. These share the websocket with the binary
// sample stream, so the first byte distinguishes them from a sample header.
const (
	controlVersion byte = 1
	// ControlAck reports the newest painted video frame.
	ControlAck byte = 1
	// ControlKeyframeRequest asks for a keyframe after the client dropped a
	// frame and can no longer decode deltas.
	ControlKeyframeRequest byte = 2

	ControlAckSize = 8
)

var errInvalidControl = errors.New("invalid media control message")

// Control is a decoded client message.
type Control struct {
	Kind     byte
	Sequence uint32
}

// Ack reports the newest video sequence number the client has painted. The
// client sends it after presenting a frame so the server can bound how many
// frames are in flight.
type Ack struct {
	Sequence uint32
}

func DecodeControl(data []byte) (Control, error) {
	if len(data) < ControlAckSize || data[0] != controlVersion {
		return Control{}, errInvalidControl
	}
	switch data[1] {
	case ControlAck, ControlKeyframeRequest:
		return Control{Kind: data[1], Sequence: binary.BigEndian.Uint32(data[4:8])}, nil
	default:
		return Control{}, errInvalidControl
	}
}

func DecodeAck(data []byte) (Ack, error) {
	control, err := DecodeControl(data)
	if err != nil || control.Kind != ControlAck {
		return Ack{}, errInvalidControl
	}
	return Ack{Sequence: control.Sequence}, nil
}

func EncodeControl(kind byte, sequence uint32) []byte {
	data := make([]byte, ControlAckSize)
	data[0] = controlVersion
	data[1] = kind
	binary.BigEndian.PutUint32(data[4:8], sequence)
	return data
}

func EncodeAck(ack Ack) []byte {
	return EncodeControl(ControlAck, ack.Sequence)
}
