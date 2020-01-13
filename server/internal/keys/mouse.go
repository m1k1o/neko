package keys

const MOUSE_LEFT = 0
const MOUSE_MIDDLE = 1
const MOUSE_RIGHT = 2
const MOUSE_WHEEL_UP = 4
const MOUSE_WHEEL_DOWN = 5
const MOUSE_WHEEL_RIGH = 6
const MOUSE_WHEEL_LEFT = 7

var Mouse = map[int]string{}

func init() {
	Mouse[MOUSE_LEFT] = "left"
	Mouse[MOUSE_MIDDLE] = "center"
	Mouse[MOUSE_RIGHT] = "right"
	Mouse[MOUSE_WHEEL_UP] = "wheelUp"
	Mouse[MOUSE_WHEEL_DOWN] = "wheelDown"
	Mouse[MOUSE_WHEEL_RIGH] = "wheelRight"
	Mouse[MOUSE_WHEEL_LEFT] = "wheelLeft"
}
