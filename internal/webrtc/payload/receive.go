package payload

import "math"

const (
	OP_MOVE     = 0x01
	OP_SCROLL   = 0x02
	OP_KEY_DOWN = 0x03
	OP_KEY_UP   = 0x04
	OP_BTN_DOWN = 0x05
	OP_BTN_UP   = 0x06
	OP_PING     = 0x07
)

type Move struct {
	X uint16
	Y uint16
}

type Scroll struct {
	X int16
	Y int16
}

type Key struct {
	Key uint32
}

type Ping struct {
	// client's timestamp split into two uint32
	ClientTs1 uint32
	ClientTs2 uint32
}

func (p Ping) ClientTs() uint64 {
	return (uint64(p.ClientTs1) * uint64(math.MaxUint32)) + uint64(p.ClientTs2)
}
