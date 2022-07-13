package desktop

import (
	"github.com/demodesk/neko/pkg/xevent"
)

func (manager *DesktopManagerCtx) OnCursorChanged(listener func(serial uint64)) {
	xevent.Emmiter.On("cursor-changed", func(payload ...interface{}) {
		listener(payload[0].(uint64))
	})
}

func (manager *DesktopManagerCtx) OnClipboardUpdated(listener func()) {
	xevent.Emmiter.On("clipboard-updated", func(payload ...interface{}) {
		listener()
	})
}

func (manager *DesktopManagerCtx) OnFileChooserDialogOpened(listener func()) {
	xevent.Emmiter.On("file-chooser-dialog-opened", func(payload ...interface{}) {
		listener()
	})
}

func (manager *DesktopManagerCtx) OnFileChooserDialogClosed(listener func()) {
	xevent.Emmiter.On("file-chooser-dialog-closed", func(payload ...interface{}) {
		listener()
	})
}

func (manager *DesktopManagerCtx) OnEventError(listener func(error_code uint8, message string, request_code uint8, minor_code uint8)) {
	xevent.Emmiter.On("event-error", func(payload ...interface{}) {
		listener(payload[0].(uint8), payload[1].(string), payload[2].(uint8), payload[3].(uint8))
	})
}
