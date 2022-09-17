package types

import "image"

type CursorImage struct {
	Width  uint16
	Height uint16
	Xhot   uint16
	Yhot   uint16
	Serial uint64
	Image  *image.RGBA
}

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

type KeyboardModifiers struct {
	NumLock  *bool
	CapsLock *bool
}

type KeyboardMap struct {
	Layout  string
	Variant string
}

type DesktopManager interface {
	Start()
	Shutdown() error
	OnBeforeScreenSizeChange(listener func())
	OnAfterScreenSizeChange(listener func())

	// clipboard
	ReadClipboard() string
	WriteClipboard(data string)

	// xorg
	Move(x, y int)
	GetCursorPosition() (int, int)
	Scroll(x, y int)
	ButtonDown(code uint32) error
	KeyDown(code uint32) error
	ButtonUp(code uint32) error
	KeyUp(code uint32) error
	ButtonPress(code uint32) error
	KeyPress(codes ...uint32) error
	ResetKeys()
	ScreenConfigurations() map[int]ScreenConfiguration
	SetScreenSize(ScreenSize) error
	GetScreenSize() *ScreenSize
	SetKeyboardMap(KeyboardMap) error
	GetKeyboardMap() (*KeyboardMap, error)
	SetKeyboardModifiers(mod KeyboardModifiers)
	GetKeyboardModifiers() KeyboardModifiers
	GetCursorImage() *CursorImage
	GetScreenshotImage() *image.RGBA

	// xevent
	OnCursorChanged(listener func(serial uint64))
	OnClipboardUpdated(listener func())
	OnFileChooserDialogOpened(listener func())
	OnFileChooserDialogClosed(listener func())
	OnEventError(listener func(error_code uint8, message string, request_code uint8, minor_code uint8))
}
