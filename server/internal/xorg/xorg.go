// NOTE: I have no fucking clue what I'm doing with this,
// it works, but I am positive I'm doing this very wrong...
// should I be freeing these strings? does go gc them?
// pretty sure this *isn't* thread safe either.... /shrug
// if you know a better way to get this done *please* make a pr <3

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
	"n.eko.moe/neko/internal/xorg/keycode"
)

var ScreenConfigurations = make(map[int]types.ScreenConfiguration)

var debounce = make(map[int]time.Time)
var buttons = make(map[int]types.Button)
var keys = make(map[int]types.Key)
var mu = sync.Mutex{}

func init() {
	keys[keycode.BACKSPACE.Code] = keycode.BACKSPACE
	keys[keycode.TAB.Code] = keycode.TAB
	keys[keycode.CLEAR.Code] = keycode.CLEAR
	keys[keycode.ENTER.Code] = keycode.ENTER
	keys[keycode.SHIFT.Code] = keycode.SHIFT
	keys[keycode.CTRL.Code] = keycode.CTRL
	keys[keycode.ALT.Code] = keycode.ALT
	keys[keycode.PAUSE.Code] = keycode.PAUSE
	keys[keycode.CAPS_LOCK.Code] = keycode.CAPS_LOCK
	keys[keycode.ESCAPE.Code] = keycode.ESCAPE
	keys[keycode.SPACE.Code] = keycode.SPACE
	keys[keycode.PAGE_UP.Code] = keycode.PAGE_UP
	keys[keycode.PAGE_DOWN.Code] = keycode.PAGE_DOWN
	keys[keycode.END.Code] = keycode.END
	keys[keycode.HOME.Code] = keycode.HOME
	keys[keycode.LEFT_ARROW.Code] = keycode.LEFT_ARROW
	keys[keycode.UP_ARROW.Code] = keycode.UP_ARROW
	keys[keycode.RIGHT_ARROW.Code] = keycode.RIGHT_ARROW
	keys[keycode.DOWN_ARROW.Code] = keycode.DOWN_ARROW
	keys[keycode.INSERT.Code] = keycode.INSERT
	keys[keycode.DELETE.Code] = keycode.DELETE
	keys[keycode.KEY_0.Code] = keycode.KEY_0
	keys[keycode.KEY_1.Code] = keycode.KEY_1
	keys[keycode.KEY_2.Code] = keycode.KEY_2
	keys[keycode.KEY_3.Code] = keycode.KEY_3
	keys[keycode.KEY_4.Code] = keycode.KEY_4
	keys[keycode.KEY_5.Code] = keycode.KEY_5
	keys[keycode.KEY_6.Code] = keycode.KEY_6
	keys[keycode.KEY_7.Code] = keycode.KEY_7
	keys[keycode.KEY_8.Code] = keycode.KEY_8
	keys[keycode.KEY_9.Code] = keycode.KEY_9
	keys[keycode.KEY_A.Code] = keycode.KEY_A
	keys[keycode.KEY_B.Code] = keycode.KEY_B
	keys[keycode.KEY_C.Code] = keycode.KEY_C
	keys[keycode.KEY_D.Code] = keycode.KEY_D
	keys[keycode.KEY_E.Code] = keycode.KEY_E
	keys[keycode.KEY_F.Code] = keycode.KEY_F
	keys[keycode.KEY_G.Code] = keycode.KEY_G
	keys[keycode.KEY_H.Code] = keycode.KEY_H
	keys[keycode.KEY_I.Code] = keycode.KEY_I
	keys[keycode.KEY_J.Code] = keycode.KEY_J
	keys[keycode.KEY_K.Code] = keycode.KEY_K
	keys[keycode.KEY_L.Code] = keycode.KEY_L
	keys[keycode.KEY_M.Code] = keycode.KEY_M
	keys[keycode.KEY_N.Code] = keycode.KEY_N
	keys[keycode.KEY_O.Code] = keycode.KEY_O
	keys[keycode.KEY_P.Code] = keycode.KEY_P
	keys[keycode.KEY_Q.Code] = keycode.KEY_Q
	keys[keycode.KEY_R.Code] = keycode.KEY_R
	keys[keycode.KEY_S.Code] = keycode.KEY_S
	keys[keycode.KEY_T.Code] = keycode.KEY_T
	keys[keycode.KEY_U.Code] = keycode.KEY_U
	keys[keycode.KEY_V.Code] = keycode.KEY_V
	keys[keycode.KEY_W.Code] = keycode.KEY_W
	keys[keycode.KEY_X.Code] = keycode.KEY_X
	keys[keycode.KEY_Y.Code] = keycode.KEY_Y
	keys[keycode.KEY_Z.Code] = keycode.KEY_Z
	keys[keycode.WIN_LEFT.Code] = keycode.WIN_LEFT
	keys[keycode.WIN_RIGHT.Code] = keycode.WIN_RIGHT
	keys[keycode.PAD_0.Code] = keycode.PAD_0
	keys[keycode.PAD_1.Code] = keycode.PAD_1
	keys[keycode.PAD_2.Code] = keycode.PAD_2
	keys[keycode.PAD_3.Code] = keycode.PAD_3
	keys[keycode.PAD_4.Code] = keycode.PAD_4
	keys[keycode.PAD_5.Code] = keycode.PAD_5
	keys[keycode.PAD_6.Code] = keycode.PAD_6
	keys[keycode.PAD_7.Code] = keycode.PAD_7
	keys[keycode.PAD_8.Code] = keycode.PAD_8
	keys[keycode.PAD_9.Code] = keycode.PAD_9
	keys[keycode.MULTIPLY.Code] = keycode.MULTIPLY
	keys[keycode.ADD.Code] = keycode.ADD
	keys[keycode.SUBTRACT.Code] = keycode.SUBTRACT
	keys[keycode.DECIMAL.Code] = keycode.DECIMAL
	keys[keycode.DIVIDE.Code] = keycode.DIVIDE
	keys[keycode.KEY_F1.Code] = keycode.KEY_F1
	keys[keycode.KEY_F2.Code] = keycode.KEY_F2
	keys[keycode.KEY_F3.Code] = keycode.KEY_F3
	keys[keycode.KEY_F4.Code] = keycode.KEY_F4
	keys[keycode.KEY_F5.Code] = keycode.KEY_F5
	keys[keycode.KEY_F6.Code] = keycode.KEY_F6
	keys[keycode.KEY_F7.Code] = keycode.KEY_F7
	keys[keycode.KEY_F8.Code] = keycode.KEY_F8
	keys[keycode.KEY_F9.Code] = keycode.KEY_F9
	keys[keycode.KEY_F10.Code] = keycode.KEY_F10
	keys[keycode.KEY_F11.Code] = keycode.KEY_F11
	keys[keycode.KEY_F12.Code] = keycode.KEY_F12
	keys[keycode.NUM_LOCK.Code] = keycode.NUM_LOCK
	keys[keycode.SCROLL_LOCK.Code] = keycode.SCROLL_LOCK
	keys[keycode.SEMI_COLON.Code] = keycode.SEMI_COLON
	keys[keycode.EQUAL.Code] = keycode.EQUAL
	keys[keycode.COMMA.Code] = keycode.COMMA
	keys[keycode.DASH.Code] = keycode.DASH
	keys[keycode.PERIOD.Code] = keycode.PERIOD
	keys[keycode.FORWARD_SLASH.Code] = keycode.FORWARD_SLASH
	keys[keycode.GRAVE.Code] = keycode.GRAVE
	keys[keycode.OPEN_BRACKET.Code] = keycode.OPEN_BRACKET
	keys[keycode.BACK_SLASH.Code] = keycode.BACK_SLASH
	keys[keycode.CLOSE_BRAKET.Code] = keycode.CLOSE_BRAKET
	keys[keycode.SINGLE_QUOTE.Code] = keycode.SINGLE_QUOTE

	buttons[keycode.LEFT_BUTTON.Code] = keycode.LEFT_BUTTON
	buttons[keycode.CENTER_BUTTON.Code] = keycode.CENTER_BUTTON
	buttons[keycode.RIGHT_BUTTON.Code] = keycode.RIGHT_BUTTON
	buttons[keycode.SCROLL_UP_BUTTON.Code] = keycode.SCROLL_UP_BUTTON
	buttons[keycode.SCROLL_DOWN_BUTTON.Code] = keycode.SCROLL_DOWN_BUTTON
	buttons[keycode.SCROLL_LEFT_BUTTON.Code] = keycode.SCROLL_LEFT_BUTTON
	buttons[keycode.SCROLL_RIGHT_BUTTON.Code] = keycode.SCROLL_RIGHT_BUTTON

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

func ButtonDown(code int) (*types.Button, error) {
	mu.Lock()
	defer mu.Unlock()

	button, ok := buttons[code]
	if !ok {
		return nil, fmt.Errorf("invalid button %v", code)
	}

	if _, ok := debounce[code]; ok {
		return nil, fmt.Errorf("debounced button %v(%v)", button.Name, code)
	}

	debounce[code] = time.Now()

	C.XButton(C.uint(button.Keysym), C.int(1))
	return &button, nil
}

func KeyDown(code int) (*types.Key, error) {
	mu.Lock()
	defer mu.Unlock()

	key, ok := keys[code]
	if !ok {
		return nil, fmt.Errorf("invalid key %v", code)
	}

	if _, ok := debounce[code]; ok {
		return nil, fmt.Errorf("debounced key %v(%v)", key.Name, code)
	}

	debounce[code] = time.Now()

	C.XKey(C.ulong(key.Keysym), C.int(1))
	return &key, nil
}

func ButtonUp(code int) (*types.Button, error) {
	mu.Lock()
	defer mu.Unlock()

	button, ok := buttons[code]
	if !ok {
		return nil, fmt.Errorf("invalid button %v", code)
	}

	if _, ok := debounce[code]; !ok {
		return nil, fmt.Errorf("debounced button %v(%v)", button.Name, code)
	}

	delete(debounce, code)

	C.XButton(C.uint(button.Keysym), C.int(0))
	return &button, nil
}

func KeyUp(code int) (*types.Key, error) {
	mu.Lock()
	defer mu.Unlock()

	key, ok := keys[code]
	if !ok {
		return nil, fmt.Errorf("invalid key %v", code)
	}

	if _, ok := debounce[code]; !ok {
		return nil, fmt.Errorf("debounced key %v(%v)", key.Name, code)
	}

	delete(debounce, code)

	C.XKey(C.ulong(key.Keysym), C.int(0))
	return &key, nil
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
	for key := range debounce {
		if key < 8 {
			ButtonUp(key)
		} else {
			KeyUp(key)
		}

		delete(debounce, key)
	}
}

func CheckKeys(duration time.Duration) {
	t := time.Now()
	for key, start := range debounce {
		if t.Sub(start) < duration {
			continue
		}

		if key < 8 {
			ButtonUp(key)
		} else {
			KeyUp(key)
		}

		delete(debounce, key)
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
