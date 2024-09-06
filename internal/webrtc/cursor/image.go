package cursor

import (
	"reflect"
	"sync"

	"github.com/rs/zerolog"

	"github.com/demodesk/neko/pkg/types"
	"github.com/demodesk/neko/pkg/utils"
)

type ImageListener interface {
	SendCursorImage(cur *types.CursorImage, img []byte) error
}

type Image interface {
	Start()
	Shutdown()
	GetCurrent() (cur *types.CursorImage, img []byte, err error)
	AddListener(listener ImageListener)
	RemoveListener(listener ImageListener)
}

type imageEntry struct {
	*types.CursorImage
	ImagePNG []byte
}

type image struct {
	logger  zerolog.Logger
	desktop types.DesktopManager

	listeners   map[uintptr]ImageListener
	listenersMu sync.RWMutex

	cache     map[uint64]*imageEntry
	cacheMu   sync.RWMutex
	current   *imageEntry
	maxSerial uint64
}

func NewImage(logger zerolog.Logger, desktop types.DesktopManager) *image {
	return &image{
		logger:    logger.With().Str("submodule", "cursor-image").Logger(),
		desktop:   desktop,
		listeners: map[uintptr]ImageListener{},
		cache:     map[uint64]*imageEntry{},
		maxSerial: 300, // TODO: Cleanup?
	}
}

func (manager *image) Start() {
	manager.desktop.OnCursorChanged(func(serial uint64) {
		entry, err := manager.getCached(serial)
		if err != nil {
			manager.logger.Err(err).Msg("failed to get cursor image")
			return
		}

		manager.current = entry

		manager.listenersMu.RLock()
		for _, l := range manager.listeners {
			if err := l.SendCursorImage(entry.CursorImage, entry.ImagePNG); err != nil {
				manager.logger.Err(err).Msg("failed to set cursor image")
			}
		}
		manager.listenersMu.RUnlock()
	})

	manager.logger.Info().Msg("starting")
}

func (manager *image) Shutdown() {
	manager.logger.Info().Msg("shutdown")

	manager.listenersMu.Lock()
	for key := range manager.listeners {
		delete(manager.listeners, key)
	}
	manager.listenersMu.Unlock()
}

func (manager *image) getCached(serial uint64) (*imageEntry, error) {
	// zero means no serial available
	if serial == 0 || serial > manager.maxSerial {
		manager.logger.Debug().Uint64("serial", serial).Msg("cache bypass")
		return manager.fetchEntry()
	}

	manager.cacheMu.RLock()
	entry, ok := manager.cache[serial]
	manager.cacheMu.RUnlock()

	if ok {
		return entry, nil
	}

	manager.logger.Debug().Uint64("serial", serial).Msg("cache miss")

	entry, err := manager.fetchEntry()
	if err != nil {
		return nil, err
	}

	manager.cacheMu.Lock()
	manager.cache[entry.Serial] = entry
	manager.cacheMu.Unlock()

	if entry.Serial != serial {
		manager.logger.Warn().
			Uint64("expected-serial", serial).
			Uint64("received-serial", entry.Serial).
			Msg("serial mismatch")
	}

	return entry, nil
}

func (manager *image) GetCurrent() (cur *types.CursorImage, img []byte, err error) {
	if manager.current != nil {
		return manager.current.CursorImage, manager.current.ImagePNG, nil
	}

	entry, err := manager.fetchEntry()
	if err != nil {
		return nil, nil, err
	}

	manager.current = entry
	return entry.CursorImage, entry.ImagePNG, nil
}

func (manager *image) AddListener(listener ImageListener) {
	manager.listenersMu.Lock()
	defer manager.listenersMu.Unlock()

	if listener != nil {
		ptr := reflect.ValueOf(listener).Pointer()
		manager.listeners[ptr] = listener
	}
}

func (manager *image) RemoveListener(listener ImageListener) {
	manager.listenersMu.Lock()
	defer manager.listenersMu.Unlock()

	if listener != nil {
		ptr := reflect.ValueOf(listener).Pointer()
		delete(manager.listeners, ptr)
	}
}

func (manager *image) fetchEntry() (*imageEntry, error) {
	cur := manager.desktop.GetCursorImage()

	img, err := utils.CreatePNGImage(cur.Image)
	if err != nil {
		return nil, err
	}
	cur.Image = nil // free memory

	return &imageEntry{
		CursorImage: cur,
		ImagePNG:    img,
	}, nil
}
