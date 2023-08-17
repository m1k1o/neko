package xorg

/*
#cgo LDFLAGS: -lX11 -lXrandr -lXtst -lXfixes -lxcvt

#include "xorg.h"
*/
import "C"

import (
	"fmt"
	"image"
	"image/color"
	"sync"
	"time"
	"unsafe"

	"github.com/demodesk/neko/pkg/types"
)

//go:generate ./keysymdef.sh

type KbdMod uint8

const (
	KbdModCapsLock KbdMod = 2
	KbdModNumLock  KbdMod = 16
)

type ScreenConfiguration struct {
	Width  int
	Height int
	Rates  map[int]int16
}

var ScreenConfigurations = make(map[int]ScreenConfiguration)

var debounce_button = make(map[uint32]time.Time)
var debounce_key = make(map[uint32]time.Time)
var mu = sync.Mutex{}

func GetScreenConfigurations() {
	mu.Lock()
	defer mu.Unlock()

	C.XGetScreenConfigurations()
}

func DisplayOpen(display string) bool {
	mu.Lock()
	defer mu.Unlock()

	displayUnsafe := C.CString(display)
	defer C.free(unsafe.Pointer(displayUnsafe))

	ok := C.XDisplayOpen(displayUnsafe)
	return int(ok) == 1
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

func GetCursorPosition() (int, int) {
	mu.Lock()
	defer mu.Unlock()

	var x C.int
	var y C.int
	C.XCursorPosition(&x, &y)

	return int(x), int(y)
}

func Scroll(x, y int) {
	mu.Lock()
	defer mu.Unlock()

	C.XScroll(C.int(x), C.int(y))
}

func ButtonDown(code uint32) error {
	mu.Lock()
	defer mu.Unlock()

	if _, ok := debounce_button[code]; ok {
		return fmt.Errorf("debounced button %v", code)
	}

	debounce_button[code] = time.Now()

	C.XButton(C.uint(code), C.int(1))
	return nil
}

func KeyDown(code uint32) error {
	mu.Lock()
	defer mu.Unlock()

	if _, ok := debounce_key[code]; ok {
		return fmt.Errorf("debounced key %v", code)
	}

	debounce_key[code] = time.Now()

	C.XKey(C.KeySym(code), C.int(1))
	return nil
}

func ButtonUp(code uint32) error {
	mu.Lock()
	defer mu.Unlock()

	if _, ok := debounce_button[code]; !ok {
		return fmt.Errorf("debounced button %v", code)
	}

	delete(debounce_button, code)

	C.XButton(C.uint(code), C.int(0))
	return nil
}

func KeyUp(code uint32) error {
	mu.Lock()
	defer mu.Unlock()

	if _, ok := debounce_key[code]; !ok {
		return fmt.Errorf("debounced key %v", code)
	}

	delete(debounce_key, code)

	C.XKey(C.KeySym(code), C.int(0))
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
		C.XKey(C.KeySym(code), C.int(0))
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

		C.XKey(C.KeySym(code), C.int(0))
		delete(debounce_key, code)
	}
}

// set screen configuration, create new one if not exists
func ChangeScreenSize(s types.ScreenSize) (types.ScreenSize, error) {
	mu.Lock()
	defer mu.Unlock()

	// round width to 8, because of Xorg
	s.Width = s.Width - (s.Width % 8)

	// if rate is 0, set it to 60
	if s.Rate == 0 {
		s.Rate = 60
	}

	// convert variables to C types
	c_width, c_height, c_rate := C.int(s.Width), C.int(s.Height), C.short(s.Rate)

	// if screen configuration already exists, just set it
	status := C.XSetScreenConfiguration(c_width, c_height, c_rate)
	if status != C.RRSetConfigSuccess {
		// create new screen configuration
		C.XCreateScreenMode(c_width, c_height, c_rate)

		// screen configuration should exist now, set it
		status = C.XSetScreenConfiguration(c_width, c_height, c_rate)
	}

	var err error

	// if screen configuration was not set successfully, return error
	if status != C.RRSetConfigSuccess {
		err = fmt.Errorf("unknown screen configuration %s", s.String())
	}

	// if specified rate is not supported a BadValue error is returned
	if status == C.BadValue {
		err = fmt.Errorf("unsupported screen rate %d", s.Rate)
	}

	return s, err
}

func GetScreenSize() types.ScreenSize {
	mu.Lock()
	defer mu.Unlock()

	c_width, c_height, c_rate := C.int(0), C.int(0), C.short(0)
	C.XGetScreenConfiguration(&c_width, &c_height, &c_rate)

	return types.ScreenSize{
		Width:  int(c_width),
		Height: int(c_height),
		Rate:   int16(c_rate),
	}
}

func SetKeyboardModifier(mod KbdMod, active bool) {
	mu.Lock()
	defer mu.Unlock()

	num := C.int(0)
	if active {
		num = C.int(1)
	}

	C.XSetKeyboardModifier(C.int(mod), num)
}

func GetKeyboardModifiers() KbdMod {
	mu.Lock()
	defer mu.Unlock()

	return KbdMod(C.XGetKeyboardModifiers())
}

func GetCursorImage() *types.CursorImage {
	mu.Lock()
	defer mu.Unlock()

	cur := C.XGetCursorImage()
	defer C.XFree(unsafe.Pointer(cur))

	width := int(cur.width)
	height := int(cur.height)

	// Xlib stores 32-bit data in longs, even if longs are 64-bits long.
	pixels := C.GoBytes(unsafe.Pointer(cur.pixels), C.int(width*height*8))

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pos := ((y * width) + x) * 8

			img.SetRGBA(x, y, color.RGBA{
				A: pixels[pos+3],
				R: pixels[pos+2],
				G: pixels[pos+1],
				B: pixels[pos+0],
			})
		}
	}

	return &types.CursorImage{
		Width:  uint16(width),
		Height: uint16(height),
		Xhot:   uint16(cur.xhot),
		Yhot:   uint16(cur.yhot),
		Serial: uint64(cur.cursor_serial),
		Image:  img,
	}
}

func GetScreenshotImage() *image.RGBA {
	mu.Lock()
	defer mu.Unlock()

	var w, h C.int
	pixelsUnsafe := C.XGetScreenshot(&w, &h)
	pixels := C.GoBytes(unsafe.Pointer(pixelsUnsafe), w*h*3)
	defer C.free(unsafe.Pointer(pixelsUnsafe))

	width := int(w)
	height := int(h)
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for row := 0; row < height; row++ {
		for col := 0; col < width; col++ {
			pos := ((row * width) + col) * 3

			img.SetRGBA(col, row, color.RGBA{
				R: uint8(pixels[pos]),
				G: uint8(pixels[pos+1]),
				B: uint8(pixels[pos+2]),
				A: 0xFF,
			})
		}
	}

	return img
}

//export goCreateScreenSize
func goCreateScreenSize(index C.int, width C.int, height C.int, mwidth C.int, mheight C.int) {
	ScreenConfigurations[int(index)] = ScreenConfiguration{
		Width:  int(width),
		Height: int(height),
		Rates:  make(map[int]int16),
	}
}

//export goSetScreenRates
func goSetScreenRates(index C.int, rate_index C.int, rateC C.short) {
	ScreenConfigurations[int(index)].Rates[int(rate_index)] = int16(rateC)
}
