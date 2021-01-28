package desktop

import (
	"os/exec"
	"strings"
)

func (manager *DesktopManagerCtx) ReadClipboard() (string, error) {
	out, err := exec.Command("xclip", "-selection", "clipboard", "-o").Output()
	return string(out), err
}

func (manager *DesktopManagerCtx) WriteClipboard(data string) error {
	cmd := exec.Command("xclip", "-selection", "clipboard", "-i")
    cmd.Stdin = strings.NewReader(data)
	return cmd.Run()
}
