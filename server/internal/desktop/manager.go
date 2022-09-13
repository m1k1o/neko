package desktop

import (
	"fmt"
	"sync"
	"time"

	"m1k1o/neko/internal/config"
	"m1k1o/neko/internal/desktop/xorg"
	"m1k1o/neko/internal/types"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type DesktopManagerCtx struct {
	logger   zerolog.Logger
	wg       sync.WaitGroup
	shutdown chan struct{}
	config   *config.Desktop
}

func New(config *config.Desktop, broadcast types.BroadcastManager) *DesktopManagerCtx {
	return &DesktopManagerCtx{
		logger:   log.With().Str("module", "desktop").Logger(),
		shutdown: make(chan struct{}),
		config:   config,
	}
}

func (manager *DesktopManagerCtx) Start() {
	if xorg.DisplayOpen(manager.config.Display) {
		manager.logger.Panic().Str("display", manager.config.Display).Msg("unable to open display")
	}

	xorg.GetScreenConfigurations()

	err := xorg.ChangeScreenSize(manager.config.ScreenWidth, manager.config.ScreenHeight, manager.config.ScreenRate)
	manager.logger.Err(err).
		Str("screen_size", fmt.Sprintf("%dx%d@%d", manager.config.ScreenWidth, manager.config.ScreenHeight, manager.config.ScreenRate)).
		Msgf("setting initial screen size")

	manager.wg.Add(1)

	go func() {
		defer manager.wg.Done()

		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-manager.shutdown:
				return
			case <-ticker.C:
				xorg.CheckKeys(time.Second * 10)
			}
		}
	}()
}

func (manager *DesktopManagerCtx) Shutdown() error {
	manager.logger.Info().Msgf("desktop shutting down")

	close(manager.shutdown)
	manager.wg.Wait()

	xorg.DisplayClose()
	return nil
}
