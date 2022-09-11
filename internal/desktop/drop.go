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

	drop.Emmiter.Once("create", func(payload ...any) {
		manager.Move(0, 0)
	})

	drop.Emmiter.Once("cursor-enter", func(payload ...any) {
		//nolint
		manager.ButtonDown(1)
	})

	drop.Emmiter.Once("button-press", func(payload ...any) {
		manager.Move(x, y)
	})

	drop.Emmiter.Once("begin", func(payload ...any) {
		for i := 0; i < dropMoveRepeat; i++ {
			manager.Move(x, y)
			time.Sleep(dropMoveDelay)
		}

		//nolint
		manager.ButtonUp(1)
	})

	finished := make(chan bool)
	drop.Emmiter.Once("finish", func(payload ...any) {
		b, ok := payload[0].(bool)
		// workaround until https://github.com/kataras/go-events/pull/8 is merged
		if !ok {
			b = (payload[0].([]any))[0].(bool)
		}
		finished <- b
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
