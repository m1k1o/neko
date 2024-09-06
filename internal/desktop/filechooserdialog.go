package desktop

import (
	"errors"
	"os/exec"

	"github.com/demodesk/neko/pkg/xorg"
)

// name of the window that is being controlled
const fileChooserDialogName = "Open File"

// short sleep value between fake user interactions
const fileChooserDialogShortSleep = "0.2"

// long sleep value between fake user interactions
const fileChooserDialogLongSleep = "0.4"

func (manager *DesktopManagerCtx) HandleFileChooserDialog(uri string) error {
	mu.Lock()
	defer mu.Unlock()

	// TODO: Use native API.
	err1 := exec.Command(
		"xdotool",
		"search", "--name", fileChooserDialogName, "windowfocus",
		"sleep", fileChooserDialogShortSleep,
		"key", "--clearmodifiers", "ctrl+l",
		"type", "--args", "1", uri+"//",
		"sleep", fileChooserDialogShortSleep,
		"key", "Delete", // remove autocomplete results
		"sleep", fileChooserDialogShortSleep,
		"key", "Return",
		"sleep", fileChooserDialogLongSleep,
		"key", "Down",
		"key", "--clearmodifiers", "ctrl+a",
		"key", "Return",
		"sleep", fileChooserDialogLongSleep,
	).Run()

	if err1 != nil {
		return err1
	}

	// TODO: Use native API.
	err2 := exec.Command(
		"xdotool",
		"search", "--name", fileChooserDialogName,
	).Run()

	// if last command didn't return error, consider dialog as still open
	if err2 == nil {
		return errors.New("unable to select files in dialog")
	}

	return nil
}

func (manager *DesktopManagerCtx) CloseFileChooserDialog() {
	for i := 0; i < 5; i++ {
		mu.Lock()

		manager.logger.Debug().Msg("attempting to close file chooser dialog")

		// TODO: Use native API.
		err := exec.Command(
			"xdotool",
			"search", "--name", fileChooserDialogName, "windowfocus",
		).Run()

		if err != nil {
			mu.Unlock()
			manager.logger.Info().Msg("file chooser dialog is closed")
			return
		}

		// custom press Alt + F4
		// because xdotool is failing to send proper Alt+F4

		//nolint
		manager.KeyPress(xorg.XK_Alt_L, xorg.XK_F4)

		mu.Unlock()
	}
}

func (manager *DesktopManagerCtx) IsFileChooserDialogEnabled() bool {
	return manager.config.FileChooserDialog
}

func (manager *DesktopManagerCtx) IsFileChooserDialogOpened() bool {
	mu.Lock()
	defer mu.Unlock()

	// TODO: Use native API.
	err := exec.Command(
		"xdotool",
		"search", "--name", fileChooserDialogName,
	).Run()

	return err == nil
}
