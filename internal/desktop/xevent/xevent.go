package xevent

/*
#cgo linux LDFLAGS: -lX11 -lXfixes

#include "xevent.h"
*/
import "C"

import (
	"unsafe"

	"github.com/kataras/go-events"
)

var emmiter events.EventEmmiter

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

func OnWindowCreated(listener func(window uint32, name string, role string)) {
	emmiter.On("window-created", func(payload ...interface{}) {
		listener(payload[0].(uint32), payload[1].(string), payload[2].(string))
	})
}

func OnWindowConfigured(listener func(window uint32, name string, role string)) {
	emmiter.On("window-configured", func(payload ...interface{}) {
		listener(payload[0].(uint32), payload[1].(string), payload[2].(string))
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

//export goXEventWindowCreated
func goXEventWindowCreated(window C.Window, name *C.char, role *C.char) {
	emmiter.Emit("window-created", uint32(window), C.GoString(name), C.GoString(role))
}

//export goXEventWindowConfigured
func goXEventWindowConfigured(window C.Window, name *C.char, role *C.char) {
	emmiter.Emit("window-configured", uint32(window), C.GoString(name), C.GoString(role))
}

//export goXEventError
func goXEventError(event *C.XErrorEvent, message *C.char) {
	emmiter.Emit("event-error", uint8(event.error_code), C.GoString(message), uint8(event.request_code), uint8(event.minor_code))
}

//export goXEventActive
func goXEventActive() C.int {
	return C.int(1)
}
