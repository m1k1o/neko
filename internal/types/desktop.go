package types

import (
	"image"
)

type CursorImage struct {
	Width  uint16
	Height uint16
	Xhot   uint16
	Yhot   uint16
	Serial uint64
	Pixels []byte
}

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

	// xorg
	Move(x, y int)
	Scroll(x, y int)
	ButtonDown(code int) error
	KeyDown(code uint64) error
	ButtonUp(code int) error
	KeyUp(code uint64) error
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

	// clipboard
	ClipboardGetBinary(mime string) ([]byte, error)
	ClipboardSetBinary(mime string, data []byte) error
	ClipboardGetTargets() ([]string, error)
	ClipboardGetPlainText() (string, error)
	ClipboardSetPlainText(data string) error
	ClipboardGetRichText() (string, error)
	ClipboardSetRichText(data string) error

	// drop
	DropFiles(x int, y int, files []string) bool

	// filechooser
	HandleFileChooserDialog(uri string) error
	CloseFileChooserDialog()
	IsFileChooserDialogOpened() bool
}
