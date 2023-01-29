package desktop

import (
	"fmt"
	"sync"
	"time"

	"m1k1o/neko/internal/config"
	"m1k1o/neko/internal/desktop/xevent"
	"m1k1o/neko/internal/desktop/xorg"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var mu = sync.Mutex{}

type DesktopManagerCtx struct {
	logger   zerolog.Logger
	wg       sync.WaitGroup
	shutdown chan struct{}
	config   *config.Desktop

	screenSizeChangeChannel chan bool
}

func New(config *config.Desktop) *DesktopManagerCtx {
	return &DesktopManagerCtx{
		logger:   log.With().Str("module", "desktop").Logger(),
		shutdown: make(chan struct{}),
		config:   config,

		screenSizeChangeChannel: make(chan bool),
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

	go xevent.EventLoop(manager.config.Display)

	go func() {
		for {
			msg, ok := <-xevent.EventErrorChannel
			if !ok {
				manager.logger.Info().Msg("xevent error channel was closed")
				return
			}

			manager.logger.Warn().
				Uint8("error_code", msg.Error_code).
				Str("message", msg.Message).
				Uint8("request_code", msg.Request_code).
				Uint8("minor_code", msg.Minor_code).
				Msg("X event error occurred")
		}
	}()

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

func (manager *DesktopManagerCtx) GetScreenSizeChangeChannel() chan bool {
	return manager.screenSizeChangeChannel
}

func (manager *DesktopManagerCtx) Shutdown() error {
	manager.logger.Info().Msgf("desktop shutting down")

	close(manager.shutdown)
	close(manager.screenSizeChangeChannel)
	manager.wg.Wait()

	xorg.DisplayClose()
	return nil
}
