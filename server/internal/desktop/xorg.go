package desktop

import (
	"os/exec"

	"m1k1o/neko/internal/desktop/xorg"
	"m1k1o/neko/internal/types"
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

func (manager *DesktopManagerCtx) SetKeyboardLayout(layout string) {
	_ = exec.Command("setxkbmap", layout).Run()
}

func (manager *DesktopManagerCtx) SetKeyboardModifiers(NumLock int, CapsLock int, ScrollLock int) {
	xorg.SetKeyboardModifiers(NumLock, CapsLock, ScrollLock)
}
