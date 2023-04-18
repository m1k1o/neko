package xevent

/*
#cgo LDFLAGS: -lX11 -lXfixes

#include "xevent.h"
*/
import "C"

import (
	"strings"
	"unsafe"

	"github.com/kataras/go-events"
)

var Emmiter events.EventEmmiter
var file_chooser_dialog_window uint32 = 0

func init() {
	Emmiter = events.New()
}

func EventLoop(display string) {
	displayUnsafe := C.CString(display)
	defer C.free(unsafe.Pointer(displayUnsafe))

	C.XEventLoop(displayUnsafe)
}

//export goXEventCursorChanged
func goXEventCursorChanged(event C.XFixesCursorNotifyEvent) {
	Emmiter.Emit("cursor-changed", uint64(event.cursor_serial))
}

//export goXEventClipboardUpdated
func goXEventClipboardUpdated() {
	Emmiter.Emit("clipboard-updated")
}

//export goXEventConfigureNotify
func goXEventConfigureNotify(display *C.Display, window C.Window, name *C.char, role *C.char) {
	if C.GoString(role) != "GtkFileChooserDialog" {
		return
	}

	// TODO: Refactor. Right now processing of this dialog relies on identifying
	// via its name. When that changes to role, this condition should be removed.
	if !strings.HasPrefix(C.GoString(name), "Open File") {
		return
	}

	C.XFileChooserHide(display, window)

	if file_chooser_dialog_window == 0 {
		file_chooser_dialog_window = uint32(window)
		Emmiter.Emit("file-chooser-dialog-opened")
	}
}

//export goXEventUnmapNotify
func goXEventUnmapNotify(window C.Window) {
	if uint32(window) != file_chooser_dialog_window {
		return
	}

	file_chooser_dialog_window = 0
	Emmiter.Emit("file-chooser-dialog-closed")
}

//export goXEventError
func goXEventError(event *C.XErrorEvent, message *C.char) {
	Emmiter.Emit("event-error", uint8(event.error_code), C.GoString(message), uint8(event.request_code), uint8(event.minor_code))
}

//export goXEventActive
func goXEventActive() C.int {
	return C.int(1)
}
