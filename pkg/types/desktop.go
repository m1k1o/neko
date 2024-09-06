package types

import (
	"fmt"
	"image"
)

type CursorImage struct {
	Width  uint16
	Height uint16
	Xhot   uint16
	Yhot   uint16
	Serial uint64
	Image  *image.RGBA
}

type ScreenSize struct {
	Width  int
	Height int
	Rate   int16
}

func (s ScreenSize) String() string {
	return fmt.Sprintf("%dx%d@%d", s.Width, s.Height, s.Rate)
}

type KeyboardModifiers struct {
	Shift    *bool `json:"shift"`
	CapsLock *bool `json:"capslock"`
	Control  *bool `json:"control"`
	Alt      *bool `json:"alt"`
	NumLock  *bool `json:"numlock"`
	Meta     *bool `json:"meta"`
	Super    *bool `json:"super"`
	AltGr    *bool `json:"altgr"`
}

type KeyboardMap struct {
	Layout  string `json:"layout"`
	Variant string `json:"variant"`
}

type ClipboardText struct {
	Text string
	HTML string
}

type DesktopManager interface {
	Start()
	Shutdown() error
	OnBeforeScreenSizeChange(listener func())
	OnAfterScreenSizeChange(listener func())

	// xorg
	Move(x, y int)
	GetCursorPosition() (int, int)
	Scroll(deltaX, deltaY int, controlKey bool)
	ButtonDown(code uint32) error
	KeyDown(code uint32) error
	ButtonUp(code uint32) error
	KeyUp(code uint32) error
	ButtonPress(code uint32) error
	KeyPress(codes ...uint32) error
	ResetKeys()
	ScreenConfigurations() []ScreenSize
	SetScreenSize(ScreenSize) (ScreenSize, error)
	GetScreenSize() ScreenSize
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

	// input driver
	HasTouchSupport() bool
	TouchBegin(touchId uint32, x, y int, pressure uint8) error
	TouchUpdate(touchId uint32, x, y int, pressure uint8) error
	TouchEnd(touchId uint32, x, y int, pressure uint8) error

	// clipboard
	ClipboardGetText() (*ClipboardText, error)
	ClipboardSetText(data ClipboardText) error
	ClipboardGetBinary(mime string) ([]byte, error)
	ClipboardSetBinary(mime string, data []byte) error
	ClipboardGetTargets() ([]string, error)

	// drop
	DropFiles(x int, y int, files []string) bool
	IsUploadDropEnabled() bool

	// filechooser
	HandleFileChooserDialog(uri string) error
	CloseFileChooserDialog()
	IsFileChooserDialogEnabled() bool
	IsFileChooserDialogOpened() bool
}
