package types

type ScreenSize struct {
	Width  int
	Height int
	Rate   int16
}

type ScreenConfiguration struct {
	Width  int
	Height int
	Rates  map[int]int16
}

type DesktopManager interface {
	Start()
	Shutdown() error
	OnBeforeScreenSizeChange(listener func())
	OnAfterScreenSizeChange(listener func())

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
