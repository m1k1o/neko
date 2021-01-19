package xorg

/*
#cgo linux LDFLAGS: -lX11 -lXrandr -lXtst -lXfixes

#include "xorg.h"
*/
import "C"

import (
	"fmt"
	"sync"
	"time"
	"unsafe"

	"demodesk/neko/internal/types"
)

type KbdModifiers uint8

const (
	KBD_CAPS_LOCK KbdModifiers = 2
	KBD_NUM_LOCK  KbdModifiers = 16
)

var ScreenConfigurations = make(map[int]types.ScreenConfiguration)

var debounce_button = make(map[int]time.Time)
var debounce_key = make(map[uint64]time.Time)
var mu = sync.Mutex{}

func GetScreenConfigurations() {
	mu.Lock()
	defer mu.Unlock()

	C.XGetScreenConfigurations()
}

func DisplayOpen(display string) error {
	mu.Lock()
	defer mu.Unlock()

	displayUnsafe := C.CString(display)
	defer C.free(unsafe.Pointer(displayUnsafe))

	err := C.XDisplayOpen(displayUnsafe)
	if int(err) == 1 {
		return fmt.Errorf("Could not open display %s.", display)
	}

	return nil
}

func DisplayClose() {
	mu.Lock()
	defer mu.Unlock()
	
	C.XDisplayClose()
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

	if _, ok := debounce_button[code]; ok {
		return fmt.Errorf("debounced button %v", code)
	}

	debounce_button[code] = time.Now()

	C.XButton(C.uint(code), C.int(1))
	return nil
}

func KeyDown(code uint64) error {
	mu.Lock()
	defer mu.Unlock()

	if _, ok := debounce_key[code]; ok {
		return fmt.Errorf("debounced key %v", code)
	}

	debounce_key[code] = time.Now()

	C.XKey(C.ulong(code), C.int(1))
	return nil
}

func ButtonUp(code int) error {
	mu.Lock()
	defer mu.Unlock()

	if _, ok := debounce_button[code]; !ok {
		return fmt.Errorf("debounced button %v", code)
	}

	delete(debounce_button, code)

	C.XButton(C.uint(code), C.int(0))
	return nil
}

func KeyUp(code uint64) error {
	mu.Lock()
	defer mu.Unlock()

	if _, ok := debounce_key[code]; !ok {
		return fmt.Errorf("debounced key %v", code)
	}

	delete(debounce_key, code)

	C.XKey(C.ulong(code), C.int(0))
	return nil
}

func ResetKeys() {
	mu.Lock()
	defer mu.Unlock()

	for code := range debounce_button {
		C.XButton(C.uint(code), C.int(0))
		delete(debounce_button, code)
	}

	for code := range debounce_key {
		C.XKey(C.ulong(code), C.int(0))
		delete(debounce_key, code)
	}
}

func CheckKeys(duration time.Duration) {
	mu.Lock()
	defer mu.Unlock()

	t := time.Now()
	for code, start := range debounce_button {
		if t.Sub(start) < duration {
			continue
		}

		C.XButton(C.uint(code), C.int(0))
		delete(debounce_button, code)
	}

	for code, start := range debounce_key {
		if t.Sub(start) < duration {
			continue
		}

		C.XKey(C.ulong(code), C.int(0))
		delete(debounce_key, code)
	}
}

func ChangeScreenSize(width int, height int, rate int16) error {
	mu.Lock()
	defer mu.Unlock()

	for index, size := range ScreenConfigurations {
		if size.Width == width && size.Height == height {
			for _, fps := range size.Rates {
				if rate == fps {
					C.XSetScreenConfiguration(C.int(index), C.short(fps))
					return nil
				}
			}
		}
	}

	return fmt.Errorf("Unknown screen configuration %dx%d@%d.", width, height, rate)
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

func SetKeyboardModifier(mod KbdModifiers, active bool) {
	mu.Lock()
	defer mu.Unlock()

	num := C.int(0)
	if active {
		num = C.int(1)
	}

	C.XSetKeyboardModifier(C.int(mod), num)
}

func GetKeyboardModifiers() KbdModifiers {
	mu.Lock()
	defer mu.Unlock()

	return KbdModifiers(C.XGetKeyboardModifiers())
}

func GetCursorImage() *types.CursorImage {
	mu.Lock()
	defer mu.Unlock()

	var cur *C.XFixesCursorImage
	cur = C.XGetCursorImage()
	defer C.XFree(unsafe.Pointer(cur))

	width := uint16(cur.width)
	height := uint16(cur.height)

	return &types.CursorImage{
		Width: width,
		Height: height,
		Xhot: uint16(cur.xhot),
		Yhot: uint16(cur.yhot),
		Serial: uint64(cur.cursor_serial),
		// Xlib stores 32-bit data in longs, even if longs are 64-bits long.
		Pixels: C.GoBytes(unsafe.Pointer(cur.pixels), C.int(width * height * 8)),
	}
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
