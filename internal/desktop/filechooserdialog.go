package desktop

import (
	"fmt"
	"os/exec"
	"time"
)

const (
	FILE_CHOOSER_DIALOG_NAME        = "Open File"
	FILE_CHOOSER_DIALOG_SHORT_SLEEP = "0.2"
	FILE_CHOOSER_DIALOG_LONG_SLEEP  = "0.4"
)

func (manager *DesktopManagerCtx) HandleFileChooserDialog(uri string) error {
	mu.Lock()
	defer mu.Unlock()

	// TODO: Use native API.
	err1 := exec.Command(
		"xdotool",
		"search", "--name", FILE_CHOOSER_DIALOG_NAME, "windowfocus",
		"sleep", FILE_CHOOSER_DIALOG_SHORT_SLEEP,
		"key", "--clearmodifiers", "ctrl+l",
		"type", "--args", "1", uri+"//",
		"sleep", FILE_CHOOSER_DIALOG_SHORT_SLEEP,
		"key", "Delete", // remove autocomplete results
		"sleep", FILE_CHOOSER_DIALOG_SHORT_SLEEP,
		"key", "Return",
		"sleep", FILE_CHOOSER_DIALOG_LONG_SLEEP,
		"key", "Down",
		"key", "--clearmodifiers", "ctrl+a",
		"key", "Return",
		"sleep", FILE_CHOOSER_DIALOG_LONG_SLEEP,
	).Run()

	if err1 != nil {
		return err1
	}

	// TODO: Use native API.
	err2 := exec.Command(
		"xdotool",
		"search", "--name", FILE_CHOOSER_DIALOG_NAME,
	).Run()

	// if last command didn't return error, consider dialog as still open
	if err2 == nil {
		return fmt.Errorf("unable to select files in dialog")
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
			"search", "--name", FILE_CHOOSER_DIALOG_NAME, "windowfocus",
		).Run()

		if err != nil {
			mu.Unlock()
			manager.logger.Info().Msg("file chooser dialog is closed")
			return
		}

		// custom press Alt + F4
		// because xdotool is failing to send proper Alt+F4

		manager.ResetKeys()
		//nolint
		manager.KeyDown(65513) // Alt
		//nolint
		manager.KeyDown(65473) // F4
		time.Sleep(10 * time.Millisecond)
		manager.ResetKeys()

		mu.Unlock()
	}
}

func (manager *DesktopManagerCtx) IsFileChooserDialogOpened() bool {
	mu.Lock()
	defer mu.Unlock()

	// TODO: Use native API.
	err := exec.Command(
		"xdotool",
		"search", "--name", FILE_CHOOSER_DIALOG_NAME,
	).Run()

	return err == nil
}
