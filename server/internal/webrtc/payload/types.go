package payload

import "fmt"

const HeaderSize = 3

type Header struct {
	Event  uint8
	Length uint16
}

// ValidateLength checks the body size declared by the v3 data-channel
// protocol. Length is the number of bytes after the three-byte header and is
// encoded in network byte order by both the browser and server.
func ValidateLength(event uint8, length uint16) error {
	var valid bool
	switch event {
	case OP_MOVE:
		valid = length == 4
	case OP_SCROLL:
		valid = length == 5
	case OP_KEY_DOWN, OP_KEY_UP, OP_BTN_DOWN, OP_BTN_UP:
		valid = length == 4
	case OP_PING:
		valid = length == 8
	case OP_TOUCH_BEGIN, OP_TOUCH_UPDATE, OP_TOUCH_END:
		valid = length == 13
	default:
		// Unknown events are left to the handler for forward compatibility.
		return nil
	}

	if !valid {
		return fmt.Errorf("invalid payload length %d for event 0x%02x", length, event)
	}
	return nil
}
