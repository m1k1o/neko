package desktop

import (
	"image"
	"os/exec"
	"regexp"
	"time"

	"github.com/demodesk/neko/pkg/types"
	"github.com/demodesk/neko/pkg/xorg"
)

func (manager *DesktopManagerCtx) Move(x, y int) {
	xorg.Move(x, y)
}

func (manager *DesktopManagerCtx) GetCursorPosition() (int, int) {
	return xorg.GetCursorPosition()
}

func (manager *DesktopManagerCtx) Scroll(deltaX, deltaY int, controlKey bool) {
	xorg.Scroll(deltaX, deltaY, controlKey)
}

func (manager *DesktopManagerCtx) ButtonDown(code uint32) error {
	return xorg.ButtonDown(code)
}

func (manager *DesktopManagerCtx) KeyDown(code uint32) error {
	return xorg.KeyDown(code)
}

func (manager *DesktopManagerCtx) ButtonUp(code uint32) error {
	return xorg.ButtonUp(code)
}

func (manager *DesktopManagerCtx) KeyUp(code uint32) error {
	return xorg.KeyUp(code)
}

func (manager *DesktopManagerCtx) ButtonPress(code uint32) error {
	xorg.ResetKeys()
	defer xorg.ResetKeys()

	return xorg.ButtonDown(code)
}

func (manager *DesktopManagerCtx) KeyPress(codes ...uint32) error {
	xorg.ResetKeys()
	defer xorg.ResetKeys()

	for _, code := range codes {
		if err := xorg.KeyDown(code); err != nil {
			return err
		}
	}

	if len(codes) > 1 {
		time.Sleep(10 * time.Millisecond)
	}

	return nil
}

func (manager *DesktopManagerCtx) ResetKeys() {
	xorg.ResetKeys()
}

func (manager *DesktopManagerCtx) ScreenConfigurations() []types.ScreenSize {
	var configs []types.ScreenSize
	for _, size := range xorg.ScreenConfigurations {
		for _, fps := range size.Rates {
			// filter out all irrelevant rates
			if fps > 60 || (fps > 30 && fps%10 != 0) {
				continue
			}

			configs = append(configs, types.ScreenSize{
				Width:  size.Width,
				Height: size.Height,
				Rate:   fps,
			})
		}
	}
	return configs
}

func (manager *DesktopManagerCtx) SetScreenSize(screenSize types.ScreenSize) (types.ScreenSize, error) {
	mu.Lock()
	manager.emmiter.Emit("before_screen_size_change")

	defer func() {
		manager.emmiter.Emit("after_screen_size_change")
		mu.Unlock()
	}()

	screenSize, err := xorg.ChangeScreenSize(screenSize)
	if err == nil {
		// cache the new screen size
		manager.screenSize = screenSize
	}

	return screenSize, err
}

func (manager *DesktopManagerCtx) GetScreenSize() types.ScreenSize {
	return xorg.GetScreenSize()
}

func (manager *DesktopManagerCtx) SetKeyboardMap(kbd types.KeyboardMap) error {
	// TOOD: Use native API.
	cmd := exec.Command("setxkbmap", "-layout", kbd.Layout, "-variant", kbd.Variant)
	_, err := cmd.Output()
	return err
}

func (manager *DesktopManagerCtx) GetKeyboardMap() (*types.KeyboardMap, error) {
	// TOOD: Use native API.
	cmd := exec.Command("setxkbmap", "-query")
	res, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	kbd := types.KeyboardMap{}

	re := regexp.MustCompile(`layout:\s+(.*)\n`)
	arr := re.FindStringSubmatch(string(res))
	if len(arr) > 1 {
		kbd.Layout = arr[1]
	}

	re = regexp.MustCompile(`variant:\s+(.*)\n`)
	arr = re.FindStringSubmatch(string(res))
	if len(arr) > 1 {
		kbd.Variant = arr[1]
	}

	return &kbd, nil
}

func (manager *DesktopManagerCtx) SetKeyboardModifiers(mod types.KeyboardModifiers) {
	if mod.Shift != nil {
		xorg.SetKeyboardModifier(xorg.KbdModShift, *mod.Shift)
	}

	if mod.CapsLock != nil {
		xorg.SetKeyboardModifier(xorg.KbdModCapsLock, *mod.CapsLock)
	}

	if mod.Control != nil {
		xorg.SetKeyboardModifier(xorg.KbdModControl, *mod.Control)
	}

	if mod.Alt != nil {
		xorg.SetKeyboardModifier(xorg.KbdModAlt, *mod.Alt)
	}

	if mod.NumLock != nil {
		xorg.SetKeyboardModifier(xorg.KbdModNumLock, *mod.NumLock)
	}

	if mod.Meta != nil {
		xorg.SetKeyboardModifier(xorg.KbdModMeta, *mod.Meta)
	}

	if mod.Super != nil {
		xorg.SetKeyboardModifier(xorg.KbdModSuper, *mod.Super)
	}

	if mod.AltGr != nil {
		xorg.SetKeyboardModifier(xorg.KbdModAltGr, *mod.AltGr)
	}
}

func (manager *DesktopManagerCtx) GetKeyboardModifiers() types.KeyboardModifiers {
	modifiers := xorg.GetKeyboardModifiers()

	isset := func(mod xorg.KbdMod) *bool {
		x := modifiers&mod != 0
		return &x
	}

	return types.KeyboardModifiers{
		Shift:    isset(xorg.KbdModShift),
		CapsLock: isset(xorg.KbdModCapsLock),
		Control:  isset(xorg.KbdModControl),
		Alt:      isset(xorg.KbdModAlt),
		NumLock:  isset(xorg.KbdModNumLock),
		Meta:     isset(xorg.KbdModMeta),
		Super:    isset(xorg.KbdModSuper),
		AltGr:    isset(xorg.KbdModAltGr),
	}
}

func (manager *DesktopManagerCtx) GetCursorImage() *types.CursorImage {
	return xorg.GetCursorImage()
}

func (manager *DesktopManagerCtx) GetScreenshotImage() *image.RGBA {
	return xorg.GetScreenshotImage()
}
