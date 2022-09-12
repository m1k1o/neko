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

func Read() string {
	mu.Lock()
	defer mu.Unlock()

	clipboardUnsafe := C.ClipboardGet()
	defer C.free(unsafe.Pointer(clipboardUnsafe))

	return C.GoString(clipboardUnsafe)
}

func Write(data string) {
	mu.Lock()
	defer mu.Unlock()

	clipboardUnsafe := C.CString(data)
	defer C.free(unsafe.Pointer(clipboardUnsafe))

	C.ClipboardSet(clipboardUnsafe)
}
