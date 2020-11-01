package types

type ScreenSize struct {
	Width  int   `json:"width"`
	Height int   `json:"height"`
	Rate   int16 `json:"rate"`
}

type ScreenConfiguration struct {
	Width  int           `json:"width"`
	Height int           `json:"height"`
	Rates  map[int]int16 `json:"rates"`
}

type DesktopManager interface {
	Start()
	Shutdown() error

	// xorg
	ChangeScreenSize(width int, height int, rate int) error
	Move(x, y int)
	Scroll(x, y int)
	ButtonDown(code int) error
	KeyDown(code uint64) error
	ButtonUp(code int) error
	KeyUp(code uint64) error
	ResetKeys()
	ScreenConfigurations() map[int]ScreenConfiguration
	GetScreenSize() *ScreenSize
	SetKeyboardLayout(layout string)
	SetKeyboardModifiers(NumLock int, CapsLock int, ScrollLock int)

	// clipboard
	ReadClipboard() string
	WriteClipboard(data string)
}
