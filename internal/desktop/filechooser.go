package desktop

import (
	"time"
	"os/exec"
)

var (
	file_chooser_dialog_open = false
)

func (manager *DesktopManagerCtx) fileChooserDialogStart() {
	if manager.IsFileChooserDialogOpen() {
		manager.CloseFileChooserDialog()
	}

	manager.OnWindowCreated(func(window uint32, name string, role string) {
		if role != "GtkFileChooserDialog" {
			return
		}

		// TODO: Implement, call event.
		file_chooser_dialog_open = true
	
		manager.logger.Debug().
			Uint32("window", window).
			Msg("GtkFileChooserDialog has been opened")
	})

	manager.OnWindowConfigured(func(window uint32, name string, role string) {
		if role != "GtkFileChooserDialog" {
			return
		}

		go func(){
			// TOOD: Refactor. 
			manager.PutWindowBelow(window)

			// Because first dialog is not put properly to background
			time.Sleep(500 * time.Millisecond)
			manager.PutWindowBelow(window)
		}()
	
		manager.logger.Debug().
			Uint32("window", window).
			Msg("GtkFileChooserDialog has been put below main window")
	})
}

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

	// TODO: Implement, call event.
	file_chooser_dialog_open = false

	_, err := cmd.Output()
	return err
}

func (manager *DesktopManagerCtx) CloseFileChooserDialog() bool {
	for i := 0; i < 5; i++ {
		// TOOD: Use native API.
		mu.Lock()
		exec.Command(
			"xdotool",
				"search", "--name", "Open", "windowfocus",
				"sleep", "0.2",
				"key", "--clearmodifiers", "alt+F4",
		).Output()
		mu.Unlock()

		if !manager.IsFileChooserDialogOpen() {
			// TODO: Implement, call event.
			file_chooser_dialog_open = false
			return true
		}
	}

	return false
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
