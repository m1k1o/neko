package gtk

/*
#cgo linux CFLAGS: -I/usr/src -I/usr/local/include/
#cgo pkg-config: gdk-3.0 gtk+-3.0

#include "gtk.h"
*/
import "C"

import (
	"sync"
)

var mu = sync.Mutex{}

func DragWindow(files []string) {
	mu.Lock()
	defer mu.Unlock()

	size := C.int(len(files))
	urisUnsafe := C.uris_make(size);
	defer C.uris_free(urisUnsafe, size)

	for i, file := range files {
		C.uris_set_file(urisUnsafe, C.CString(file), C.int(i))
	}

	C.drag_window(urisUnsafe)
}
