package clipboard

/*
#cgo linux LDFLAGS: /usr/local/lib/libclipboard.a -lxcb

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
