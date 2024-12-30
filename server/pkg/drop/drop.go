package drop

/*
#cgo pkg-config: gtk+-3.0

#include "drop.h"
*/
import "C"

import (
	"sync"

	"github.com/kataras/go-events"
)

var Emmiter events.EventEmmiter
var mu = sync.Mutex{}

func init() {
	Emmiter = events.New()
}

func OpenWindow(files []string) {
	mu.Lock()
	defer mu.Unlock()

	size := C.int(len(files))
	urisUnsafe := C.dragUrisMake(size)
	defer C.dragUrisFree(urisUnsafe, size)

	for i, file := range files {
		C.dragUrisSetFile(urisUnsafe, C.CString(file), C.int(i))
	}

	C.dragWindowOpen(urisUnsafe)
}

func CloseWindow() {
	C.dragWindowClose()
}

//export goDragCreate
func goDragCreate(widget *C.GtkWidget, event *C.GdkEvent, user_data C.gpointer) {
	go Emmiter.Emit("create")
}

//export goDragCursorEnter
func goDragCursorEnter(widget *C.GtkWidget, event *C.GdkEvent, user_data C.gpointer) {
	go Emmiter.Emit("cursor-enter")
}

//export goDragButtonPress
func goDragButtonPress(widget *C.GtkWidget, event *C.GdkEvent, user_data C.gpointer) {
	go Emmiter.Emit("button-press")
}

//export goDragBegin
func goDragBegin(widget *C.GtkWidget, context *C.GdkDragContext, user_data C.gpointer) {
	go Emmiter.Emit("begin")
}

//export goDragFinish
func goDragFinish(succeeded C.gboolean) {
	go Emmiter.Emit("finish", bool(succeeded == C.int(1)))
}
