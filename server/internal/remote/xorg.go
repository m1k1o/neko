package remote

import (
	"m1k1o/neko/internal/remote/xorg"
	"m1k1o/neko/internal/types"
	"os/exec"
)

func (manager *RemoteManager) Move(x, y int) {
	xorg.Move(x, y)
}

func (manager *RemoteManager) Scroll(x, y int) {
	xorg.Scroll(x, y)
}

func (manager *RemoteManager) ButtonDown(code int) error {
	return xorg.ButtonDown(code)
}

func (manager *RemoteManager) KeyDown(code uint64) error {
	return xorg.KeyDown(code)
}

func (manager *RemoteManager) ButtonUp(code int) error {
	return xorg.ButtonUp(code)
}

func (manager *RemoteManager) KeyUp(code uint64) error {
	return xorg.KeyUp(code)
}

func (manager *RemoteManager) ResetKeys() {
	xorg.ResetKeys()
}

func (manager *RemoteManager) ScreenConfigurations() map[int]types.ScreenConfiguration {
	return xorg.ScreenConfigurations
}

func (manager *RemoteManager) GetScreenSize() *types.ScreenSize {
	return xorg.GetScreenSize()
}

func (manager *RemoteManager) SetKeyboardLayout(layout string) {
	_ = exec.Command("setxkbmap", layout).Run()
}

func (manager *RemoteManager) SetKeyboardModifiers(NumLock int, CapsLock int, ScrollLock int) {
	xorg.SetKeyboardModifiers(NumLock, CapsLock, ScrollLock)
}
