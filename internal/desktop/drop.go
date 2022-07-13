package desktop

import (
	"time"

	"github.com/demodesk/neko/pkg/drop"
)

// repeat move event multiple times
const dropMoveRepeat = 4

// wait after each repeated move event
const dropMoveDelay = 100 * time.Millisecond

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
		for i := 0; i < dropMoveRepeat; i++ {
			manager.Move(x, y)
			time.Sleep(dropMoveDelay)
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
