package desktop

import (
	"fmt"
	"time"

	"github.com/kataras/go-events"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"demodesk/neko/internal/config"
	"demodesk/neko/internal/desktop/xorg"
)

type DesktopManagerCtx struct {
	logger    zerolog.Logger
	cleanup   *time.Ticker
	shutdown  chan bool
	emmiter   events.EventEmmiter
	display   string
	config    *config.Desktop
}

func New(display string, config *config.Desktop) *DesktopManagerCtx {
	return &DesktopManagerCtx{
		logger:    log.With().Str("module", "desktop").Logger(),
		cleanup:   time.NewTicker(1 * time.Second),
		shutdown:  make(chan bool),
		emmiter:   events.New(),
		display:   display,
		config:    config,
	}
}

func (manager *DesktopManagerCtx) Start() {
	if err := xorg.DisplayOpen(manager.display); err != nil {
		manager.logger.Warn().Err(err).Msg("unable to open dispaly")
	}

	xorg.GetScreenConfigurations()

	manager.logger.Info().
		Str("screen_size", fmt.Sprintf("%dx%d@%d", manager.config.ScreenWidth, manager.config.ScreenHeight, manager.config.ScreenRate)).
		Msgf("Setting initial screen size...")

	if err := xorg.ChangeScreenSize(manager.config.ScreenWidth, manager.config.ScreenHeight, manager.config.ScreenRate); err != nil {
		manager.logger.Warn().Err(err).Msg("unable to set initial screen size")
	}

	go func() {
		defer func() {
			xorg.DisplayClose()
			manager.logger.Info().Msg("shutdown")
		}()

		for {
			select {
			case <-manager.shutdown:
				return
			case <-manager.cleanup.C:
				xorg.CheckKeys(time.Second * 10)
			}
		}
	}()
}

func (manager *DesktopManagerCtx) OnBeforeScreenSizeChange(listener func()) {
	manager.emmiter.On("before_screen_size_change", func(payload ...interface{}) {
		listener()
	})
}

func (manager *DesktopManagerCtx) OnAfterScreenSizeChange(listener func()) {
	manager.emmiter.On("after_screen_size_change", func(payload ...interface{}) {
		listener()
	})
}

func (manager *DesktopManagerCtx) Shutdown() error {
	manager.logger.Info().Msgf("remote shutting down")

	manager.cleanup.Stop()
	manager.shutdown <- true
	return nil
}
