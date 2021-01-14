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
	for step := 1; step <= 6; step++ {
		time.Sleep(DELAY)

		switch step {
		case 1:
			manager.Move(0, 0)
		case 2:
			manager.ButtonDown(1)
		case 3:
			fallthrough
		case 4:
			fallthrough
		case 5:
			manager.Move(x, y)
		case 6:
			manager.ButtonUp(1)
		}
	}
}
