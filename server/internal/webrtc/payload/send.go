package payload

import "math"

const (
	OP_CURSOR_POSITION = 0x01
	OP_CURSOR_IMAGE    = 0x02
	OP_PONG            = 0x03
)

type CursorPosition struct {
	X uint16
	Y uint16
}

type CursorImage struct {
	Width  uint16
	Height uint16
	Xhot   uint16
	Yhot   uint16
}

type Pong struct {
	Ping

	// server's timestamp split into two uint32
	ServerTs1 uint32
	ServerTs2 uint32
}

func (p Pong) ServerTs() uint64 {
	return (uint64(p.ServerTs1) * uint64(math.MaxUint32)) + uint64(p.ServerTs2)
}
