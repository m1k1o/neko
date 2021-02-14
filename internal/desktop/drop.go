package desktop

import (
	"time"

	"demodesk/neko/internal/desktop/drop"
)

const (
	DROP_MOVE_REPEAT = 4
	DROP_DELAY       = 100 * time.Millisecond
)

func (manager *DesktopManagerCtx) DropFiles(x int, y int, files []string) bool {
	mu.Lock()
	defer mu.Unlock()

	drop.Emmiter.Clear()

	drop.Emmiter.Once("create", func(payload ...interface{}) {
		manager.Move(0, 0)
	})

	drop.Emmiter.Once("cursor-enter", func(payload ...interface{}) {
		//nolint
		manager.ButtonDown(1)
	})

	drop.Emmiter.Once("button-press", func(payload ...interface{}) {
		manager.Move(x, y)
	})

	drop.Emmiter.Once("begin", func(payload ...interface{}) {
		for i := 0; i < DROP_MOVE_REPEAT; i++ {
			manager.Move(x, y)
			time.Sleep(DROP_DELAY)
		}

		//nolint
		manager.ButtonUp(1)
	})

	finished := make(chan bool)
	drop.Emmiter.Once("finish", func(payload ...interface{}) {
		finished <- payload[0].(bool)
	})

	manager.ResetKeys()
	go drop.OpenWindow(files)

	select {
	case succeeded := <-finished:
		return succeeded
	case <-time.After(1 * time.Second):
		drop.CloseWindow()
		return false
	}
}
