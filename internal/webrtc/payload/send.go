package payload

const (
	OP_CURSOR_POSITION = 0x01
	OP_CURSOR_IMAGE    = 0x02
)

type CursorPosition struct {
	Header

	X uint16
	Y uint16
}

type CursorImage struct {
	Header

	Width  uint16
	Height uint16
	Xhot   uint16
	Yhot   uint16
}
