package cursor

import (
	"reflect"
	"sync"

	"github.com/rs/zerolog"
)

type PositionListener interface {
	SendCursorPosition(x, y int) error
}

type Position interface {
	Shutdown()
	Set(x, y int)
	AddListener(listener PositionListener)
	RemoveListener(listener PositionListener)
}

type position struct {
	logger zerolog.Logger

	listeners   map[uintptr]PositionListener
	listenersMu sync.RWMutex
}

func NewPosition(logger zerolog.Logger) *position {
	return &position{
		logger:    logger.With().Str("submodule", "cursor-position").Logger(),
		listeners: map[uintptr]PositionListener{},
	}
}

func (manager *position) Shutdown() {
	manager.logger.Info().Msg("shutdown")

	manager.listenersMu.Lock()
	for key := range manager.listeners {
		delete(manager.listeners, key)
	}
	manager.listenersMu.Unlock()
}

func (manager *position) Set(x, y int) {
	manager.listenersMu.RLock()
	defer manager.listenersMu.RUnlock()

	for _, l := range manager.listeners {
		if err := l.SendCursorPosition(x, y); err != nil {
			manager.logger.Err(err).Msg("failed to set cursor position")
		}
	}
}

func (manager *position) AddListener(listener PositionListener) {
	manager.listenersMu.Lock()
	defer manager.listenersMu.Unlock()

	if listener != nil {
		ptr := reflect.ValueOf(listener).Pointer()
		manager.listeners[ptr] = listener
	}
}

func (manager *position) RemoveListener(listener PositionListener) {
	manager.listenersMu.Lock()
	defer manager.listenersMu.Unlock()

	if listener != nil {
		ptr := reflect.ValueOf(listener).Pointer()
		delete(manager.listeners, ptr)
	}
}
