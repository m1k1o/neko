package xevent

/*
#cgo LDFLAGS: -lX11 -lXfixes

#include "xevent.h"
*/
import "C"

import (
	"unsafe"

	"github.com/kataras/go-events"
)

var Emmiter events.EventEmmiter

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

}

//export goXEventUnmapNotify
func goXEventUnmapNotify(window C.Window) {

}

//export goXEventError
func goXEventError(event *C.XErrorEvent, message *C.char) {
	Emmiter.Emit("event-error", uint8(event.error_code), C.GoString(message), uint8(event.request_code), uint8(event.minor_code))
}

//export goXEventActive
func goXEventActive() C.int {
	return C.int(1)
}
