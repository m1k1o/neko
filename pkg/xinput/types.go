package xinput

import "time"

const (
	// absolute coordinates used in driver
	AbsX = 0xffff
	AbsY = 0xffff
)

const (
	XI_TouchBegin  = 18
	XI_TouchUpdate = 19
	XI_TouchEnd    = 20
)

type Message struct {
	_type    uint16
	touchId  uint32
	x        int32 // can be negative?
	y        int32 // can be negative?
	pressure uint8
}

func (msg *Message) Unpack(buffer []byte) {
	msg._type = uint16(buffer[0])
	msg.touchId = uint32(buffer[1]) | (uint32(buffer[2]) << 8)
	msg.x = int32(buffer[3]) | (int32(buffer[4]) << 8) | (int32(buffer[5]) << 16) | (int32(buffer[6]) << 24)
	msg.y = int32(buffer[7]) | (int32(buffer[8]) << 8) | (int32(buffer[9]) << 16) | (int32(buffer[10]) << 24)
	msg.pressure = uint8(buffer[11])
}

func (msg *Message) Pack() []byte {
	var buffer [12]byte

	buffer[0] = byte(msg._type)
	buffer[1] = byte(msg.touchId)
	buffer[2] = byte(msg.touchId >> 8)
	buffer[3] = byte(msg.x)
	buffer[4] = byte(msg.x >> 8)
	buffer[5] = byte(msg.x >> 16)
	buffer[6] = byte(msg.x >> 24)
	buffer[7] = byte(msg.y)
	buffer[8] = byte(msg.y >> 8)
	buffer[9] = byte(msg.y >> 16)
	buffer[10] = byte(msg.y >> 24)
	buffer[11] = byte(msg.pressure)

	return buffer[:]
}

type Driver interface {
	Connect() error
	Close() error
	// release touches, that were not updated for duration
	Debounce(duration time.Duration)
	// touch events
	TouchBegin(touchId uint32, x, y int, pressure uint8) error
	TouchUpdate(touchId uint32, x, y int, pressure uint8) error
	TouchEnd(touchId uint32, x, y int, pressure uint8) error
}
