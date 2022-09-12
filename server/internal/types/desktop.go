package types

type DesktopManager interface {
	Start()
	Shutdown() error
	// clipboard
	ReadClipboard() string
	WriteClipboard(data string)
	// xorg
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
}
