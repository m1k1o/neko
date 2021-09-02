package payload

const (
	OP_MOVE     = 0x01
	OP_SCROLL   = 0x02
	OP_KEY_DOWN = 0x03
	OP_KEY_UP   = 0x04
	OP_BTN_DOWN = 0x05
	OP_BTN_UP   = 0x06
)

type Move struct {
	Header

	X uint16
	Y uint16
}

type Scroll struct {
	Header

	X int16
	Y int16
}

type Key struct {
	Header

	Key uint32
}
