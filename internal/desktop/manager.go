package desktop

import (
	"time"

	"github.com/kataras/go-events"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"demodesk/neko/internal/desktop/xorg"
)

type DesktopManagerCtx struct {
	logger    zerolog.Logger
	cleanup   *time.Ticker
	shutdown  chan bool
	display   string
}

func New(display string) *DesktopManagerCtx {
	return &DesktopManagerCtx{
		logger:    log.With().Str("module", "desktop").Logger(),
		cleanup:   time.NewTicker(1 * time.Second),
		shutdown:  make(chan bool),
		display:   display,
	}
}

func (manager *DesktopManagerCtx) Start() {
	if err := xorg.DisplayOpen(manager.display); err != nil {
		manager.logger.Warn().Err(err).Msg("unable to open dispaly")
	}

	xorg.GetScreenConfigurations()

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

func (manager *DesktopManagerCtx) Shutdown() error {
	manager.logger.Info().Msgf("remote shutting down")

	manager.cleanup.Stop()
	manager.shutdown <- true
	return nil
}
