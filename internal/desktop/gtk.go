package desktop

import (
	"time"

	"demodesk/neko/internal/desktop/gtk"
)

const (
	DELAY = 100 * time.Millisecond
)

func (manager *DesktopManagerCtx) DropFiles(x int, y int, files []string) {
	mu.Lock()
	defer mu.Unlock()

	go gtk.DragWindow(files)

	// TODO: Find a bettter way.
	time.Sleep(DELAY)
	manager.Move(0, 0)
	manager.ButtonDown(1)
	manager.Move(x, y)
	time.Sleep(DELAY)
	manager.Move(x, y)
	time.Sleep(DELAY)
	manager.Move(x, y)
	time.Sleep(DELAY)
	manager.ButtonUp(1)
}
