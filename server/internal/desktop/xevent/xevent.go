package xevent

/*
#cgo LDFLAGS: -lX11 -lXfixes

#include "xevent.h"
*/
import "C"

import (
	"unsafe"

	"m1k1o/neko/internal/types"
)

var CursorChangedChannel chan uint64
var ClipboardUpdatedChannel chan bool
var FileChooserDialogClosedChannel chan bool
var FileChooserDialogOpenedChannel chan bool
var EventErrorChannel chan types.DesktopErrorMessage

func init() {
	CursorChangedChannel = make(chan uint64)
	ClipboardUpdatedChannel = make(chan bool)
	FileChooserDialogClosedChannel = make(chan bool)
	FileChooserDialogOpenedChannel = make(chan bool)
	EventErrorChannel = make(chan types.DesktopErrorMessage)

	go func() {
		for {
			// TODO: Unused.
			<-CursorChangedChannel
		}
	}()
	go func() {
		for {
			// TODO: Unused.
			<-FileChooserDialogClosedChannel
		}
	}()
	go func() {
		for {
			// TODO: Unused.
			<-FileChooserDialogOpenedChannel
		}
	}()
}

func EventLoop(display string) {
	displayUnsafe := C.CString(display)
	defer C.free(unsafe.Pointer(displayUnsafe))

	C.XEventLoop(displayUnsafe)
}

//export goXEventCursorChanged
func goXEventCursorChanged(event C.XFixesCursorNotifyEvent) {
	CursorChangedChannel <- uint64(event.cursor_serial)
}

//export goXEventClipboardUpdated
func goXEventClipboardUpdated() {
	ClipboardUpdatedChannel <- true
}

//export goXEventConfigureNotify
func goXEventConfigureNotify(display *C.Display, window C.Window, name *C.char, role *C.char) {

}

//export goXEventUnmapNotify
func goXEventUnmapNotify(window C.Window) {

}

//export goXEventError
func goXEventError(event *C.XErrorEvent, message *C.char) {
	EventErrorChannel <- types.DesktopErrorMessage{
		Error_code:   uint8(event.error_code),
		Message:      C.GoString(message),
		Request_code: uint8(event.request_code),
		Minor_code:   uint8(event.minor_code),
	}
}

//export goXEventActive
func goXEventActive() C.int {
	return C.int(1)
}
