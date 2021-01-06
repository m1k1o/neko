package drop

/*
#cgo linux CFLAGS: -I/usr/src -I/usr/local/include/
#cgo pkg-config: gdk-3.0 gtk+-3.0

#include "drop.h"
*/
import "C"

import (
	"sync"
)

var mu = sync.Mutex{}

func FileDrop(x int, y int, uris []string) {
	mu.Lock()
	defer mu.Unlock()

	size := C.int(len(uris))
	urisUnsafe := C.uris_make(size);
	defer C.uris_free(urisUnsafe, size)

	for i, uri := range uris {
		C.uris_set(urisUnsafe, C.CString(uri), C.int(i))
	}

	C.drag_window(urisUnsafe)
}
