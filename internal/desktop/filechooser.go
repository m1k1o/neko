package desktop

import (
	"os/exec"
)

func (manager *DesktopManagerCtx) HandleFileChooserDialog(uri string) error {
	mu.Lock()
	defer mu.Unlock()

	// TOOD: Use native API.
    cmd := exec.Command(
		"xdotool",
			"search", "--name", "Open", "windowfocus",
			"sleep", "0.2",
			"key", "--clearmodifiers", "ctrl+l",
			"type", "--args", "1", uri + "//",
			"key", "--clearmodifiers", "Return",
			"sleep", "1",
			"key", "--clearmodifiers", "Down",
			"key", "--clearmodifiers", "ctrl+a",
			"key", "--clearmodifiers", "Return",
			"sleep", "0.3",
	)

	_, err := cmd.Output()
	return err
}

func (manager *DesktopManagerCtx) CloseFileChooserDialog() error {
	mu.Lock()
	defer mu.Unlock()

	// TOOD: Use native API.
    cmd := exec.Command(
		"xdotool",
			"search", "--name", "Open", "windowfocus",
			"sleep", "0.2",
			"key", "--clearmodifiers", "alt+f4",
	)

	_, err := cmd.Output()
	return err
}

func (manager *DesktopManagerCtx) IsFileChooserDialogOpen() bool {
	mu.Lock()
	defer mu.Unlock()

	// TOOD: Use native API.
    cmd := exec.Command(
		"xdotool",
			"search", "--name", "Open", "windowfocus",
	)

	_, err := cmd.Output()
	return err == nil
}
