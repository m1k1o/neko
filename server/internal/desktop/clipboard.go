package desktop

import "m1k1o/neko/internal/desktop/clipboard"

func (manager *DesktopManagerCtx) ReadClipboard() string {
	return clipboard.Read()
}

func (manager *DesktopManagerCtx) WriteClipboard(data string) {
	clipboard.Write(data)
}
