package desktop

import (
	"fmt"
	"sync"
	"time"

	"m1k1o/neko/internal/config"
	"m1k1o/neko/internal/desktop/xevent"
	"m1k1o/neko/internal/desktop/xorg"
	"m1k1o/neko/internal/types"

	"github.com/kataras/go-events"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var mu = sync.Mutex{}

type DesktopManagerCtx struct {
	logger   zerolog.Logger
	wg       sync.WaitGroup
	shutdown chan struct{}
	emmiter  events.EventEmmiter
	config   *config.Desktop
}

func New(config *config.Desktop, broadcast types.BroadcastManager) *DesktopManagerCtx {
	return &DesktopManagerCtx{
		logger:   log.With().Str("module", "desktop").Logger(),
		shutdown: make(chan struct{}),
		emmiter:  events.New(),
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

	manager.OnEventError(func(error_code uint8, message string, request_code uint8, minor_code uint8) {
		manager.logger.Warn().
			Uint8("error_code", error_code).
			Str("message", message).
			Uint8("request_code", request_code).
			Uint8("minor_code", minor_code).
			Msg("X event error occurred")
	})

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

func (manager *DesktopManagerCtx) OnBeforeScreenSizeChange(listener func()) {
	manager.emmiter.On("before_screen_size_change", func(payload ...any) {
		listener()
	})
}

func (manager *DesktopManagerCtx) OnAfterScreenSizeChange(listener func()) {
	manager.emmiter.On("after_screen_size_change", func(payload ...any) {
		listener()
	})
}

func (manager *DesktopManagerCtx) Shutdown() error {
	manager.logger.Info().Msgf("desktop shutting down")

	close(manager.shutdown)
	manager.wg.Wait()

	xorg.DisplayClose()
	return nil
}
