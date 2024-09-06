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
	// touch events
	OP_TOUCH_BEGIN  = 0x08
	OP_TOUCH_UPDATE = 0x09
	OP_TOUCH_END    = 0x0a
)

type Move struct {
	X uint16
	Y uint16
}

// TODO: remove this once the client is fixed
type Scroll_Old struct {
	X int16
	Y int16
}

type Scroll struct {
	DeltaX     int16
	DeltaY     int16
	ControlKey bool
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

type Touch struct {
	TouchId  uint32
	X        int32
	Y        int32
	Pressure uint8
}
