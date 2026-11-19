package desktop

import (
	"time"

	"github.com/m1k1o/neko/server/pkg/drop"
)

const (
	// repeat move event multiple times
	dropMoveRepeat = 4

	// wait after each repeated move event
	dropMoveDelay = 100 * time.Millisecond

	// wait for drop to finish before giving up
	dropFinishTimeout = 1 * time.Second
)

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
		for range dropMoveRepeat {
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
	case <-time.After(dropFinishTimeout):
		drop.CloseWindow()
		return false
	}
}

func (manager *DesktopManagerCtx) IsUploadDropEnabled() bool {
	return manager.config.UploadDrop
}
