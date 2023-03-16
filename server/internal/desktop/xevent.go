package desktop

import (
	"m1k1o/neko/internal/desktop/xevent"
	"m1k1o/neko/internal/types"
)

func (manager *DesktopManagerCtx) GetCursorChangedChannel() chan uint64 {
	return xevent.CursorChangedChannel
}

func (manager *DesktopManagerCtx) GetClipboardUpdatedChannel() chan struct{} {
	return xevent.ClipboardUpdatedChannel
}

func (manager *DesktopManagerCtx) GetEventErrorChannel() chan types.DesktopErrorMessage {
	return xevent.EventErrorChannel
}
