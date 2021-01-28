package desktop

import (
	"fmt"
	"bytes"
	"os/exec"
	"strings"
)

func (manager *DesktopManagerCtx) ClipboardGetBinary(mime string) ([]byte, error) {
	cmd := exec.Command("xclip", "-selection", "clipboard", "-o", "-t", mime)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		msg := strings.TrimSpace(string(stderr.Bytes()))
		return nil, fmt.Errorf("%s", msg)
	}

	return stdout.Bytes(), nil
}

func (manager *DesktopManagerCtx) ClipboardSetBinary(mime string, data []byte) error {
	cmd := exec.Command("xclip", "-selection", "clipboard", "-i", "-t", mime)
	cmd.Stdin = bytes.NewReader(data)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		msg := strings.TrimSpace(string(stderr.Bytes()))
		return fmt.Errorf("%s", msg)
	}

	return nil
}

func (manager *DesktopManagerCtx) ClipboardGetTargets() ([]string, error) {
	cmd := exec.Command("xclip", "-selection", "clipboard", "-o", "-t", "TARGETS")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		msg := strings.TrimSpace(string(stderr.Bytes()))
		return nil, fmt.Errorf("%s", msg)
	}

	var response []string
	targets := strings.Split(string(stdout.Bytes()), "\n")
	for _, target := range targets {
		if target == "" {
			continue
		}

		if !strings.Contains(target, "/") {
			continue
		}
		
		response = append(response, target)
	}

	return response, nil
}

func (manager *DesktopManagerCtx) ClipboardGetPlainText() (string, error) {
	bytes, err := manager.ClipboardGetBinary("STRING")
	return string(bytes), err
}

func (manager *DesktopManagerCtx) ClipboardSetPlainText(data string) error {
	return manager.ClipboardSetBinary("STRING", []byte(data))
}

func (manager *DesktopManagerCtx) ClipboardGetRichText() (string, error) {
	bytes, err := manager.ClipboardGetBinary("text/html")
	return string(bytes), err
}

func (manager *DesktopManagerCtx) ClipboardSetRichText(data string) error {
	return manager.ClipboardSetBinary("text/html", []byte(data))
}
