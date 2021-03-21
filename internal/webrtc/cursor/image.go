package cursor

import (
	"reflect"
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"demodesk/neko/internal/types"
	"demodesk/neko/internal/utils"
)

func NewImage(desktop types.DesktopManager) *ImageCtx {
	return &ImageCtx{
		logger:    log.With().Str("module", "cursor-image").Logger(),
		desktop:   desktop,
		listeners: map[uintptr]*func(entry *ImageEntry){},
		cache:     map[uint64]*ImageEntry{},
		maxSerial: 200, // TODO: Cleanup?
	}
}

type ImageCtx struct {
	logger    zerolog.Logger
	desktop   types.DesktopManager
	emitMu    sync.Mutex
	listeners map[uintptr]*func(entry *ImageEntry)
	cacheMu   sync.Mutex
	cache     map[uint64]*ImageEntry
	current   *ImageEntry
	maxSerial uint64
}

type ImageEntry struct {
	Cursor *types.CursorImage
	Image  []byte
}

func (manager *ImageCtx) Start() {
	manager.desktop.OnCursorChanged(func(serial uint64) {
		entry, err := manager.GetCached(serial)
		if err != nil {
			manager.logger.Warn().Err(err).Msg("failed to get cursor image")
			return
		}

		manager.current = entry
		for _, emit := range manager.listeners {
			(*emit)(entry)
		}
	})
}

func (manager *ImageCtx) Shutdown() {
	manager.logger.Info().Msgf("shutting down")

	manager.emitMu.Lock()
	for key := range manager.listeners {
		delete(manager.listeners, key)
	}
	manager.emitMu.Unlock()
}

func (manager *ImageCtx) GetCached(serial uint64) (*ImageEntry, error) {
	// zero means no serial available
	if serial == 0 || serial > manager.maxSerial {
		return manager.fetchEntry()
	}

	manager.cacheMu.Lock()
	entry, ok := manager.cache[serial]
	manager.cacheMu.Unlock()

	if ok {
		return entry, nil
	}

	entry, err := manager.fetchEntry()
	if err != nil {
		return nil, err
	}

	manager.cacheMu.Lock()
	manager.cache[serial] = entry
	manager.cacheMu.Unlock()

	return entry, nil
}

func (manager *ImageCtx) Get() (*ImageEntry, error) {
	if manager.current != nil {
		return manager.current, nil
	}

	return manager.fetchEntry()
}

func (manager *ImageCtx) AddListener(listener *func(entry *ImageEntry)) {
	manager.emitMu.Lock()
	defer manager.emitMu.Unlock()

	if listener != nil {
		ptr := reflect.ValueOf(listener).Pointer()
		manager.listeners[ptr] = listener
	}
}

func (manager *ImageCtx) RemoveListener(listener *func(entry *ImageEntry)) {
	manager.emitMu.Lock()
	defer manager.emitMu.Unlock()

	if listener != nil {
		ptr := reflect.ValueOf(listener).Pointer()
		delete(manager.listeners, ptr)
	}
}

func (manager *ImageCtx) fetchEntry() (*ImageEntry, error) {
	cur := manager.desktop.GetCursorImage()

	img, err := utils.CreatePNGImage(cur.Image)
	if err != nil {
		return nil, err
	}

	entry := &ImageEntry{
		Cursor: cur,
		Image:  img,
	}

	return entry, nil
}