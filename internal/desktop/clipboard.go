package desktop

import (
	"demodesk/neko/internal/desktop/clipboard"
)

func (manager *DesktopManagerCtx) ReadClipboard() string {
	return clipboard.ReadClipboard()
}

func (manager *DesktopManagerCtx) WriteClipboard(data string) {
	clipboard.WriteClipboard(data)
}
