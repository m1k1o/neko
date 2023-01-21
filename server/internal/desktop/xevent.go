package desktop

import (
	"m1k1o/neko/internal/desktop/xevent"
	"m1k1o/neko/internal/types"
)

func (manager *DesktopManagerCtx) GetCursorChangedChannel() (chan uint64) {
	return xevent.CursorChangedChannel
}

func (manager *DesktopManagerCtx) GetClipboardUpdatedChannel() (chan bool) {
	return xevent.ClipboardUpdatedChannel
}

func (manager *DesktopManagerCtx) GetFileChooserDialogOpenedChannel() (chan bool) {
	return xevent.FileChooserDialogOpenedChannel
}

func (manager *DesktopManagerCtx) GetFileChooserDialogClosedChannel() (chan bool) {
	return xevent.FileChooserDialogClosedChannel
}

func (manager *DesktopManagerCtx) GetEventErrorChannel() (chan types.DesktopErrorMessage) {
	return xevent.EventErrorChannel
}
