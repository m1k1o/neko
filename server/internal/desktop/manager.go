package desktop

import (
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
	xorg.Display(manager.config.Display)

	if !xorg.ValidScreenSize(manager.config.ScreenWidth, manager.config.ScreenHeight, manager.config.ScreenRate) {
		manager.logger.Warn().Msgf("invalid screen option %dx%d@%d", manager.config.ScreenWidth, manager.config.ScreenHeight, manager.config.ScreenRate)
	} else if err := xorg.ChangeScreenSize(manager.config.ScreenWidth, manager.config.ScreenHeight, manager.config.ScreenRate); err != nil {
		manager.logger.Warn().Err(err).Msg("unable to change screen size")
	}

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

	return nil
}
