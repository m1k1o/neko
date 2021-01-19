package xevent

/*
#cgo linux LDFLAGS: -lX11 -lXfixes

#include "xevent.h"
*/
import "C"

import (
	"time"
	"unsafe"

	"github.com/kataras/go-events"
)

var emmiter events.EventEmmiter
var file_chooser_dialog_window uint32 = 0

func init() {
	emmiter = events.New()
}

func EventLoop(display string) {
	displayUnsafe := C.CString(display)
	defer C.free(unsafe.Pointer(displayUnsafe))

	C.XEventLoop(displayUnsafe)
}

func OnCursorChanged(listener func(serial uint64)) {
	emmiter.On("cursor-changed", func(payload ...interface{}) {
		listener(payload[0].(uint64))
	})
}

func OnClipboardUpdated(listener func()) {
	emmiter.On("clipboard-updated", func(payload ...interface{}) {
		listener()
	})
}

func OnFileChooserDialogOpened(listener func()) {
	emmiter.On("file-chooser-dialog-opened", func(payload ...interface{}) {
		listener()
	})
}

func OnFileChooserDialogClosed(listener func()) {
	emmiter.On("file-chooser-dialog-closed", func(payload ...interface{}) {
		listener()
	})
}

func OnEventError(listener func(error_code uint8, message string, request_code uint8, minor_code uint8)) {
	emmiter.On("event-error", func(payload ...interface{}) {
		listener(payload[0].(uint8), payload[1].(string), payload[2].(uint8), payload[3].(uint8))
	})
}

//export goXEventCursorChanged
func goXEventCursorChanged(event C.XFixesCursorNotifyEvent) {
	emmiter.Emit("cursor-changed", uint64(event.cursor_serial))
}

//export goXEventClipboardUpdated
func goXEventClipboardUpdated() {
	emmiter.Emit("clipboard-updated")
}

//export goXEventCreateNotify
func goXEventCreateNotify(window C.Window, nameUnsafe *C.char, roleUnsafe *C.char) {
	role := C.GoString(roleUnsafe)
	if role != "GtkFileChooserDialog" {
		return
	}

	file_chooser_dialog_window = uint32(window)
	emmiter.Emit("file-chooser-dialog-opened")
}

//export goXEventConfigureNotify
func goXEventConfigureNotify(display *C.Display, window C.Window, nameUnsafe *C.char, roleUnsafe *C.char) {
	role := C.GoString(roleUnsafe)
	if role != "GtkFileChooserDialog" {
		return
	}

	C.XFileChooserHide(display, window)

	// Because first dialog is not put properly to background
	time.Sleep(10 * time.Millisecond)
	C.XFileChooserHide(display, window)
}

//export goXEventUnmapNotify
func goXEventUnmapNotify(window C.Window) {
	if uint32(window) != file_chooser_dialog_window {
		return
	}

	emmiter.Emit("file-chooser-dialog-closed")
}

//export goXEventError
func goXEventError(event *C.XErrorEvent, message *C.char) {
	emmiter.Emit("event-error", uint8(event.error_code), C.GoString(message), uint8(event.request_code), uint8(event.minor_code))
}

//export goXEventActive
func goXEventActive() C.int {
	return C.int(1)
}
