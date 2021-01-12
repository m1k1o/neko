package desktop

import (
	"os/exec"

	"demodesk/neko/internal/types"
	"demodesk/neko/internal/desktop/xorg"
)

func (manager *DesktopManagerCtx) Move(x, y int) {
	xorg.Move(x, y)
}

func (manager *DesktopManagerCtx) Scroll(x, y int) {
	xorg.Scroll(x, y)
}

func (manager *DesktopManagerCtx) ButtonDown(code int) error {
	return xorg.ButtonDown(code)
}

func (manager *DesktopManagerCtx) KeyDown(code uint64) error {
	return xorg.KeyDown(code)
}

func (manager *DesktopManagerCtx) ButtonUp(code int) error {
	return xorg.ButtonUp(code)
}

func (manager *DesktopManagerCtx) KeyUp(code uint64) error {
	return xorg.KeyUp(code)
}

func (manager *DesktopManagerCtx) ResetKeys() {
	xorg.ResetKeys()
}

func (manager *DesktopManagerCtx) ScreenConfigurations() map[int]types.ScreenConfiguration {
	return xorg.ScreenConfigurations
}

func (manager *DesktopManagerCtx) GetScreenSize() *types.ScreenSize {
	return xorg.GetScreenSize()
}

func (manager *DesktopManagerCtx) ChangeScreenSize(width int, height int, rate int) error {
	mu.Lock()
	manager.emmiter.Emit("before_screen_size_change")

	defer func() {
		manager.emmiter.Emit("after_screen_size_change")
		mu.Unlock()
	}()

	return xorg.ChangeScreenSize(width, height, rate)
}

func (manager *DesktopManagerCtx) SetKeyboardLayout(layout string) error {
	// TOOD: Use native API.
    cmd := exec.Command("setxkbmap", layout)
	_, err := cmd.Output()
	return err
}

func (manager *DesktopManagerCtx) SetKeyboardModifiers(mod types.KeyboardModifiers) {
	if mod.NumLock != nil {
		xorg.SetKeyboardModifier(xorg.KBD_NUM_LOCK, *mod.NumLock)
	}

	if mod.CapsLock != nil {
		xorg.SetKeyboardModifier(xorg.KBD_CAPS_LOCK, *mod.CapsLock)
	}
}

func (manager *DesktopManagerCtx) GetKeyboardModifiers() types.KeyboardModifiers {
	modifiers := xorg.GetKeyboardModifiers()

	NumLock := (modifiers & xorg.KBD_NUM_LOCK) != 0
	CapsLock := (modifiers & xorg.KBD_CAPS_LOCK) != 0

	return types.KeyboardModifiers{
		NumLock: &NumLock,
		CapsLock: &CapsLock,
	}
}

func (manager *DesktopManagerCtx) GetCursorImage() *types.CursorImage {
	return xorg.GetCursorImage()
}
