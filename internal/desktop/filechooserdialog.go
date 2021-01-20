package desktop

import (
	"time"
	"os/exec"
)

const (
	FILE_CHOOSER_DIALOG_NAME  = "Open File"
	FILE_CHOOSER_DIALOG_SLEEP = "0.2"
)

func (manager *DesktopManagerCtx) HandleFileChooserDialog(uri string) error {
	mu.Lock()
	defer mu.Unlock()

	// TODO: Use native API.
	err := exec.Command(
		"xdotool",
			"search", "--name", FILE_CHOOSER_DIALOG_NAME, "windowfocus",
			"sleep", FILE_CHOOSER_DIALOG_SLEEP,
			"key", "--clearmodifiers", "ctrl+l",
			"type", "--args", "1", uri + "//",
			"key", "Return",
			"sleep", FILE_CHOOSER_DIALOG_SLEEP,
	).Run()

	// TODO: Use native API.
	exec.Command(
		"xdotool",
			"search", "--name", FILE_CHOOSER_DIALOG_NAME, "windowfocus",
			"sleep", FILE_CHOOSER_DIALOG_SLEEP,
			"key", "Down",
			"key", "--clearmodifiers", "ctrl+a",
			"key", "Return",
			"sleep", FILE_CHOOSER_DIALOG_SLEEP,
	).Run()

	return err
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
		manager.KeyDown(65513) // Alt
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
