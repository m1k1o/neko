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
	logger                         zerolog.Logger
	wg                             sync.WaitGroup
	shutdown                       chan struct{}
	beforeScreenSizeChangeChannel  chan bool
	afterScreenSizeChangeChannel   chan int16
	config                         *config.Desktop
}

func New(config *config.Desktop) *DesktopManagerCtx {
	return &DesktopManagerCtx{
		logger:   log.With().Str("module", "desktop").Logger(),
		shutdown: make(chan struct{}),
		beforeScreenSizeChangeChannel: make (chan bool),
		afterScreenSizeChangeChannel: make (chan int16),
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

	go xevent.EventLoop(manager.config.Display)

	go func() {
		for {
			desktopErrorMessage := <- xevent.EventErrorChannel
			manager.logger.Warn().
			Uint8("error_code", desktopErrorMessage.Error_code).
			Str("message", desktopErrorMessage.Message).
			Uint8("request_code", desktopErrorMessage.Request_code).
			Uint8("minor_code", desktopErrorMessage.Minor_code).
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

func (manager *DesktopManagerCtx) GetBeforeScreenSizeChangeChannel() (chan bool) {
	return manager.beforeScreenSizeChangeChannel
}

func (manager *DesktopManagerCtx) GetAfterScreenSizeChangeChannel() (chan int16) {
	return manager.afterScreenSizeChangeChannel
}

func (manager *DesktopManagerCtx) Shutdown() error {
	manager.logger.Info().Msgf("desktop shutting down")

	close(manager.shutdown)
	manager.wg.Wait()

	xorg.DisplayClose()
	return nil
}
