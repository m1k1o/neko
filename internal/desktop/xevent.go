package desktop

import (
	"demodesk/neko/internal/desktop/xevent"
)

func (manager *DesktopManagerCtx) OnCursorChanged(listener func(serial uint64)) {
	xevent.OnCursorChanged(listener)
}

func (manager *DesktopManagerCtx) OnClipboardUpdated(listener func()) {
	xevent.OnClipboardUpdated(listener)
}

func (manager *DesktopManagerCtx) OnEventError(listener func(error_code uint8, message string, request_code uint8, minor_code uint8)) {
	xevent.OnEventError(listener)
}
