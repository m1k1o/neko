package webrtc

type dataHeader struct {
	Event  uint8
	Length uint16
}

type dataMouseMove struct {
	dataHeader
	X int16
	Y int16
}

type dataMouseKey struct {
	dataHeader
	Key uint8
}

type dataKeyboardKey struct {
	dataHeader
	Key uint16
}
