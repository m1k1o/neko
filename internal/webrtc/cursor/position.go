package cursor

import (
	"reflect"
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"demodesk/neko/internal/types"
)

func NewPosition(desktop types.DesktopManager) *PositionCtx {
	return &PositionCtx{
		logger:    log.With().Str("module", "webrtc").Str("submodule", "cursor-position").Logger(),
		desktop:   desktop,
		listeners: map[uintptr]*func(x, y int){},
	}
}

type PositionCtx struct {
	logger    zerolog.Logger
	desktop   types.DesktopManager
	emitMu    sync.Mutex
	listeners map[uintptr]*func(x, y int)
}

func (manager *PositionCtx) Shutdown() {
	manager.logger.Info().Msg("shutdown")

	manager.emitMu.Lock()
	for key := range manager.listeners {
		delete(manager.listeners, key)
	}
	manager.emitMu.Unlock()
}

func (manager *PositionCtx) Get() (x, y int) {
	return manager.desktop.GetCursorPosition()
}

func (manager *PositionCtx) Set(x, y int) {
	for _, emit := range manager.listeners {
		(*emit)(x, y)
	}
}

func (manager *PositionCtx) AddListener(listener *func(x, y int)) {
	manager.emitMu.Lock()
	defer manager.emitMu.Unlock()

	if listener != nil {
		ptr := reflect.ValueOf(listener).Pointer()
		manager.listeners[ptr] = listener
	}
}

func (manager *PositionCtx) RemoveListener(listener *func(x, y int)) {
	manager.emitMu.Lock()
	defer manager.emitMu.Unlock()

	if listener != nil {
		ptr := reflect.ValueOf(listener).Pointer()
		delete(manager.listeners, ptr)
	}
}
