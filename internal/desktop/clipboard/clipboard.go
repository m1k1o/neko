package clipboard

/*
#cgo linux CFLAGS: -I/usr/src -I/usr/local/include/
#cgo linux LDFLAGS: /usr/local/lib/libclipboard.a -L/usr/src -L/usr/local/lib -lxcb

#include "clipboard.h"
*/
import "C"

import (
	"sync"
	"unsafe"
)

var mu = sync.Mutex{}

func ReadClipboard() string {
	mu.Lock()
	defer mu.Unlock()

	clipboardUnsafe := C.ClipboardGet()
	defer C.free(unsafe.Pointer(clipboardUnsafe))

	return C.GoString(clipboardUnsafe)
}

func WriteClipboard(data string) {
	mu.Lock()
	defer mu.Unlock()

	clipboardUnsafe := C.CString(data)
	defer C.free(unsafe.Pointer(clipboardUnsafe))

	C.ClipboardSet(clipboardUnsafe)
}
