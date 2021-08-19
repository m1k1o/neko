package desktop

import (
	"fmt"
	"sync"
	"time"

	"github.com/kataras/go-events"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"demodesk/neko/internal/config"
	"demodesk/neko/internal/desktop/xevent"
	"demodesk/neko/internal/desktop/xorg"
)

var mu = sync.Mutex{}

type DesktopManagerCtx struct {
	logger   zerolog.Logger
	shutdown chan bool
	emmiter  events.EventEmmiter
	config   *config.Desktop
}

func New(config *config.Desktop) *DesktopManagerCtx {
	return &DesktopManagerCtx{
		logger:   log.With().Str("module", "desktop").Logger(),
		shutdown: make(chan bool),
		emmiter:  events.New(),
		config:   config,
	}
}

func (manager *DesktopManagerCtx) Start() {
	if xorg.DisplayOpen(manager.config.Display) {
		manager.logger.Panic().Str("display", manager.config.Display).Msg("unable to open display")
	}

	xorg.GetScreenConfigurations()

	manager.logger.Info().
		Str("screen_size", fmt.Sprintf("%dx%d@%d", manager.config.ScreenWidth, manager.config.ScreenHeight, manager.config.ScreenRate)).
		Msgf("setting initial screen size")

	if err := xorg.ChangeScreenSize(manager.config.ScreenWidth, manager.config.ScreenHeight, manager.config.ScreenRate); err != nil {
		manager.logger.Warn().Err(err).Msg("unable to set initial screen size")
	}

	go xevent.EventLoop(manager.config.Display)

	// In case it was opened
	go manager.CloseFileChooserDialog()

	manager.OnEventError(func(error_code uint8, message string, request_code uint8, minor_code uint8) {
		manager.logger.Warn().
			Uint8("error_code", error_code).
			Str("message", message).
			Uint8("request_code", request_code).
			Uint8("minor_code", minor_code).
			Msg("X event error occured")
	})

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-manager.shutdown:
				xorg.DisplayClose()
				return
			case <-ticker.C:
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
	manager.logger.Info().Msgf("desktop shutting down")

	manager.shutdown <- true
	return nil
}
