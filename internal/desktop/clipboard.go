package desktop

import (
	"fmt"
	"bytes"
	"os/exec"
	"strings"
)

func (manager *DesktopManagerCtx) ReadClipboard() (string, error) {
	cmd := exec.Command("xclip", "-selection", "clipboard", "-o")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		msg := strings.TrimSpace(string(stderr.Bytes()))
		return "", fmt.Errorf("%s", msg)
	}

	return string(stdout.Bytes()), nil
}

func (manager *DesktopManagerCtx) WriteClipboard(data string) error {
	cmd := exec.Command("xclip", "-selection", "clipboard", "-i")
	cmd.Stdin = strings.NewReader(data)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		msg := strings.TrimSpace(string(stderr.Bytes()))
		return fmt.Errorf("%s", msg)
	}

	return nil
}
