package xorg

/*
#cgo linux CFLAGS: -I/usr/src -I/usr/local/include/
#cgo linux LDFLAGS: /usr/local/lib/libclipboard.a -L/usr/src -L/usr/local/lib -lX11 -lXtst -lXrandr -lxcb

#include "xorg.h"
*/
import "C"

import (
	"fmt"
	"sync"
	"time"
	"unsafe"

	"n.eko.moe/neko/internal/types"
)

var ScreenConfigurations = make(map[int]types.ScreenConfiguration)

var debounce = make(map[int]time.Time)
var mu = sync.Mutex{}

func init() {
	C.XGetScreenConfigurations()
}

func Display(display string) {
	mu.Lock()
	defer mu.Unlock()

	displayUnsafe := C.CString(display)
	defer C.free(unsafe.Pointer(displayUnsafe))

	C.XDisplaySet(displayUnsafe)
}

func Move(x, y int) {
	mu.Lock()
	defer mu.Unlock()

	C.XMove(C.int(x), C.int(y))
}

func Scroll(x, y int) {
	mu.Lock()
	defer mu.Unlock()

	C.XScroll(C.int(x), C.int(y))
}

func ButtonDown(code int) error {
	mu.Lock()
	defer mu.Unlock()

	if _, ok := debounce[code]; ok {
		return fmt.Errorf("debounced button %v", code)
	}

	debounce[code] = time.Now()

	C.XButton(C.uint(code), C.int(1))
	return nil
}

func KeyDown(code int) error {
	mu.Lock()
	defer mu.Unlock()

	if _, ok := debounce[code]; ok {
		return fmt.Errorf("debounced key %v", code)
	}

	debounce[code] = time.Now()

	C.XKey(C.ulong(code), C.int(1))
	return nil
}

func ButtonUp(code int) error {
	mu.Lock()
	defer mu.Unlock()

	if _, ok := debounce[code]; !ok {
		return fmt.Errorf("debounced button %v", code)
	}

	delete(debounce, code)

	C.XButton(C.uint(code), C.int(0))
	return nil
}

func KeyUp(code int) error {
	mu.Lock()
	defer mu.Unlock()

	if _, ok := debounce[code]; !ok {
		return fmt.Errorf("debounced key %v", code)
	}

	delete(debounce, code)

	C.XKey(C.ulong(code), C.int(0))
	return nil
}

func ReadClipboard() string {
	mu.Lock()
	defer mu.Unlock()

	clipboardUnsafe := C.XClipboardGet()
	defer C.free(unsafe.Pointer(clipboardUnsafe))

	return C.GoString(clipboardUnsafe)
}

func WriteClipboard(data string) {
	mu.Lock()
	defer mu.Unlock()

	clipboardUnsafe := C.CString(data)
	defer C.free(unsafe.Pointer(clipboardUnsafe))

	C.XClipboardSet(clipboardUnsafe)
}

func ResetKeys() {
	for code := range debounce {
		if code < 8 {
			ButtonUp(code)
		} else {
			KeyUp(code)
		}

		delete(debounce, code)
	}
}

func CheckKeys(duration time.Duration) {
	t := time.Now()
	for code, start := range debounce {
		if t.Sub(start) < duration {
			continue
		}

		if code < 8 {
			ButtonUp(code)
		} else {
			KeyUp(code)
		}

		delete(debounce, code)
	}
}

func ValidScreenSize(width int, height int, rate int) bool {
	for _, size := range ScreenConfigurations {
		if size.Width == width && size.Height == height {
			for _, fps := range size.Rates {
				if int16(rate) == fps {
					return true
				}
			}
		}
	}

	return false
}

func ChangeScreenSize(width int, height int, rate int) error {
	mu.Lock()
	defer mu.Unlock()

	for index, size := range ScreenConfigurations {
		if size.Width == width && size.Height == height {
			for _, fps := range size.Rates {
				if int16(rate) == fps {
					C.XSetScreenConfiguration(C.int(index), C.short(fps))
					return nil
				}
			}
		}
	}

	return fmt.Errorf("unknown configuration")
}

func GetScreenSize() *types.ScreenSize {
	mu.Lock()
	defer mu.Unlock()

	index := int(C.XGetScreenSize())
	rate := int16(C.XGetScreenRate())

	if conf, ok := ScreenConfigurations[index]; ok {
		return &types.ScreenSize{
			Width:  conf.Width,
			Height: conf.Height,
			Rate:   rate,
		}
	}

	return nil
}

//export goCreateScreenSize
func goCreateScreenSize(index C.int, width C.int, height C.int, mwidth C.int, mheight C.int) {
	ScreenConfigurations[int(index)] = types.ScreenConfiguration{
		Width:  int(width),
		Height: int(height),
		Rates:  make(map[int]int16),
	}
}

//export goSetScreenRates
func goSetScreenRates(index C.int, rate_index C.int, rate C.short) {
	ScreenConfigurations[int(index)].Rates[int(rate_index)] = int16(rate)
}
