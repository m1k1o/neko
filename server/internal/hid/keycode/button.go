package keycode

type Button struct {
	Name   string
	Code   int
	Keysym int
}

var LEFT_BUTTON = Button{
	Name:   "LEFT",
	Code:   0,
	Keysym: 1,
}

var CENTER_BUTTON = Button{
	Name:   "CENTER",
	Code:   1,
	Keysym: 2,
}

var RIGHT_BUTTON = Button{
	Name:   "RIGHT",
	Code:   2,
	Keysym: 3,
}

var SCROLL_UP_BUTTON = Button{
	Name:   "SCROLL_UP",
	Code:   3,
	Keysym: 4,
}

var SCROLL_DOWN_BUTTON = Button{
	Name:   "SCROLL_DOWN",
	Code:   4,
	Keysym: 5,
}

var SCROLL_LEFT_BUTTON = Button{
	Name:   "SCROLL_LEFT",
	Code:   5,
	Keysym: 6,
}

var SCROLL_RIGHT_BUTTON = Button{
	Name:   "SCROLL_RIGHT",
	Code:   6,
	Keysym: 7,
}
