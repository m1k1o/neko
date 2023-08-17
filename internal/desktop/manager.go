package desktop

import (
	"sync"
	"time"

	"github.com/kataras/go-events"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/demodesk/neko/internal/config"
	"github.com/demodesk/neko/pkg/types"
	"github.com/demodesk/neko/pkg/xevent"
	"github.com/demodesk/neko/pkg/xinput"
	"github.com/demodesk/neko/pkg/xorg"
)

var mu = sync.Mutex{}

type DesktopManagerCtx struct {
	logger     zerolog.Logger
	wg         sync.WaitGroup
	shutdown   chan struct{}
	emmiter    events.EventEmmiter
	config     *config.Desktop
	screenSize types.ScreenSize // cached screen size
	input      xinput.Driver
}

func New(config *config.Desktop) *DesktopManagerCtx {
	var input xinput.Driver
	if config.UseInputDriver {
		input = xinput.NewDriver(config.InputSocket)
	} else {
		input = xinput.NewDummy()
	}

	return &DesktopManagerCtx{
		logger:     log.With().Str("module", "desktop").Logger(),
		shutdown:   make(chan struct{}),
		emmiter:    events.New(),
		config:     config,
		screenSize: config.ScreenSize,
		input:      input,
	}
}

func (manager *DesktopManagerCtx) Start() {
	if xorg.DisplayOpen(manager.config.Display) {
		manager.logger.Panic().Str("display", manager.config.Display).Msg("unable to open display")
	}

	xorg.GetScreenConfigurations()

	screenSize, err := xorg.ChangeScreenSize(manager.config.ScreenSize)
	if err != nil {
		manager.logger.Err(err).
			Str("screen_size", screenSize.String()).
			Msgf("unable to set initial screen size")
	} else {
		// cache screen size
		manager.screenSize = screenSize
		manager.logger.Info().
			Str("screen_size", screenSize.String()).
			Msgf("setting initial screen size")
	}

	err = manager.input.Connect()
	if err != nil {
		// TODO: fail silently to dummy driver?
		manager.logger.Panic().Err(err).Msg("unable to connect to input driver")
	}

	xevent.Unminimize = manager.config.Unminimize
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

	manager.wg.Add(1)

	go func() {
		defer manager.wg.Done()

		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		const debounceDuration = 10 * time.Second

		for {
			select {
			case <-manager.shutdown:
				return
			case <-ticker.C:
				xorg.CheckKeys(debounceDuration)
				manager.input.Debounce(debounceDuration)
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
	manager.logger.Info().Msgf("shutdown")

	close(manager.shutdown)
	manager.wg.Wait()

	xorg.DisplayClose()
	return nil
}
