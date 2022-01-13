package cmd

import (
	"os"
	"os/signal"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"demodesk/neko"
	"demodesk/neko/internal/api"
	"demodesk/neko/internal/capture"
	"demodesk/neko/internal/desktop"
	"demodesk/neko/internal/http"
	"demodesk/neko/internal/member"
	"demodesk/neko/internal/session"
	"demodesk/neko/internal/webrtc"
	"demodesk/neko/internal/websocket"
	"demodesk/neko/modules"
)

func init() {
	service := Serve{
		Version: neko.Service.Version,
		Configs: neko.Service.Configs,
	}

	command := &cobra.Command{
		Use:   "serve",
		Short: "serve neko streaming server",
		Long:  `serve neko streaming server`,
		Run:   service.ServeCommand,
	}

	cobra.OnInitialize(func() {
		service.Configs.Set()
		service.Preflight()
	})

	if err := service.Configs.Init(command); err != nil {
		log.Panic().Err(err).Msg("unable to run serve command")
	}

	root.AddCommand(command)
}

type Serve struct {
	Version *neko.Version
	Configs *neko.Configs

	logger           zerolog.Logger
	desktopManager   *desktop.DesktopManagerCtx
	captureManager   *capture.CaptureManagerCtx
	webRTCManager    *webrtc.WebRTCManagerCtx
	memberManager    *member.MemberManagerCtx
	sessionManager   *session.SessionManagerCtx
	webSocketManager *websocket.WebSocketManagerCtx
	apiManager       *api.ApiManagerCtx
	httpManager      *http.HttpManagerCtx
}

func (neko *Serve) Preflight() {
	neko.logger = log.With().Str("service", "neko").Logger()
}

func (neko *Serve) Start() {
	neko.sessionManager = session.New(
		neko.Configs.Session,
	)

	neko.memberManager = member.New(
		neko.sessionManager,
		neko.Configs.Member,
	)

	if err := neko.memberManager.Connect(); err != nil {
		neko.logger.Panic().Err(err).Msg("unable to connect to member manager")
	}

	neko.desktopManager = desktop.New(
		neko.Configs.Desktop,
	)
	neko.desktopManager.Start()

	neko.captureManager = capture.New(
		neko.desktopManager,
		neko.Configs.Capture,
	)
	neko.captureManager.Start()

	neko.webRTCManager = webrtc.New(
		neko.desktopManager,
		neko.captureManager,
		neko.Configs.WebRTC,
	)
	neko.webRTCManager.Start()

	neko.webSocketManager = websocket.New(
		neko.sessionManager,
		neko.desktopManager,
		neko.captureManager,
		neko.webRTCManager,
	)
	neko.webSocketManager.Start()

	neko.apiManager = api.New(
		neko.sessionManager,
		neko.memberManager,
		neko.desktopManager,
		neko.captureManager,
		neko.Configs.Server,
	)

	modules.Start(
		neko.sessionManager,
		neko.webSocketManager,
		neko.apiManager,
	)

	neko.httpManager = http.New(
		neko.webSocketManager,
		neko.apiManager,
		neko.Configs.Server,
	)
	neko.httpManager.Start()
}

func (neko *Serve) Shutdown() {
	var err error

	err = neko.memberManager.Disconnect()
	neko.logger.Err(err).Msg("member manager disconnect")

	err = neko.desktopManager.Shutdown()
	neko.logger.Err(err).Msg("desktop manager shutdown")

	err = neko.captureManager.Shutdown()
	neko.logger.Err(err).Msg("capture manager shutdown")

	err = neko.webRTCManager.Shutdown()
	neko.logger.Err(err).Msg("webrtc manager shutdown")

	err = neko.webSocketManager.Shutdown()
	neko.logger.Err(err).Msg("websocket manager shutdown")

	err = modules.Shutdown()
	neko.logger.Err(err).Msg("modules shutdown")

	err = neko.httpManager.Shutdown()
	neko.logger.Err(err).Msg("http manager shutdown")
}

func (neko *Serve) ServeCommand(cmd *cobra.Command, args []string) {
	neko.logger.Info().Msg("starting neko server")
	neko.Start()
	neko.logger.Info().Msg("neko ready")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	sig := <-quit

	neko.logger.Warn().Msgf("received %s, attempting graceful shutdown", sig)
	neko.Shutdown()
	neko.logger.Info().Msg("shutdown complete")
}
