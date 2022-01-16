package cmd

import (
	"os"
	"os/signal"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"demodesk/neko/internal/api"
	"demodesk/neko/internal/capture"
	"demodesk/neko/internal/config"
	"demodesk/neko/internal/desktop"
	"demodesk/neko/internal/http"
	"demodesk/neko/internal/member"
	"demodesk/neko/internal/session"
	"demodesk/neko/internal/webrtc"
	"demodesk/neko/internal/websocket"
	"demodesk/neko/modules"
)

func init() {
	service := serve{}

	command := &cobra.Command{
		Use:   "serve",
		Short: "serve neko streaming server",
		Long:  `serve neko streaming server`,
		Run:   service.Command,
	}

	cobra.OnInitialize(service.Preflight)

	if err := service.Init(command); err != nil {
		log.Panic().Err(err).Msg("unable to initialize configuration")
	}

	root.AddCommand(command)
}

type serve struct {
	logger zerolog.Logger

	configs struct {
		Root    config.Root
		Desktop config.Desktop
		Capture config.Capture
		WebRTC  config.WebRTC
		Member  config.Member
		Session config.Session
		Server  config.Server
	}

	managers struct {
		desktop   *desktop.DesktopManagerCtx
		capture   *capture.CaptureManagerCtx
		webRTC    *webrtc.WebRTCManagerCtx
		member    *member.MemberManagerCtx
		session   *session.SessionManagerCtx
		webSocket *websocket.WebSocketManagerCtx
		api       *api.ApiManagerCtx
		http      *http.HttpManagerCtx
	}
}

func (c *serve) Init(cmd *cobra.Command) error {
	if err := c.configs.Desktop.Init(cmd); err != nil {
		return err
	}
	if err := c.configs.Capture.Init(cmd); err != nil {
		return err
	}
	if err := c.configs.WebRTC.Init(cmd); err != nil {
		return err
	}
	if err := c.configs.Member.Init(cmd); err != nil {
		return err
	}
	if err := c.configs.Session.Init(cmd); err != nil {
		return err
	}
	if err := c.configs.Server.Init(cmd); err != nil {
		return err
	}

	for _, cfg := range modules.Configs() {
		if err := cfg.Init(root); err != nil {
			log.Panic().Err(err).Msg("unable to initialize configuration")
		}
	}

	return nil
}

func (c *serve) Preflight() {
	c.logger = log.With().Str("service", "neko").Logger()

	c.configs.Desktop.Set()
	c.configs.Capture.Set()
	c.configs.WebRTC.Set()
	c.configs.Member.Set()
	c.configs.Session.Set()
	c.configs.Server.Set()

	for _, cfg := range modules.Configs() {
		cfg.Set()
	}
}

func (c *serve) Start() {
	c.managers.session = session.New(
		&c.configs.Session,
	)

	c.managers.member = member.New(
		c.managers.session,
		&c.configs.Member,
	)

	if err := c.managers.member.Connect(); err != nil {
		c.logger.Panic().Err(err).Msg("unable to connect to member manager")
	}

	c.managers.desktop = desktop.New(
		&c.configs.Desktop,
	)
	c.managers.desktop.Start()

	c.managers.capture = capture.New(
		c.managers.desktop,
		&c.configs.Capture,
	)
	c.managers.capture.Start()

	c.managers.webRTC = webrtc.New(
		c.managers.desktop,
		c.managers.capture,
		&c.configs.WebRTC,
	)
	c.managers.webRTC.Start()

	c.managers.webSocket = websocket.New(
		c.managers.session,
		c.managers.desktop,
		c.managers.capture,
		c.managers.webRTC,
	)
	c.managers.webSocket.Start()

	c.managers.api = api.New(
		c.managers.session,
		c.managers.member,
		c.managers.desktop,
		c.managers.capture,
		&c.configs.Server,
	)

	modules.Start(
		c.managers.session,
		c.managers.webSocket,
		c.managers.api,
	)

	c.managers.http = http.New(
		c.managers.webSocket,
		c.managers.api,
		&c.configs.Server,
	)
	c.managers.http.Start()
}

func (c *serve) Shutdown() {
	var err error

	err = c.managers.member.Disconnect()
	c.logger.Err(err).Msg("member manager disconnect")

	err = c.managers.desktop.Shutdown()
	c.logger.Err(err).Msg("desktop manager shutdown")

	err = c.managers.capture.Shutdown()
	c.logger.Err(err).Msg("capture manager shutdown")

	err = c.managers.webRTC.Shutdown()
	c.logger.Err(err).Msg("webrtc manager shutdown")

	err = c.managers.webSocket.Shutdown()
	c.logger.Err(err).Msg("websocket manager shutdown")

	err = modules.Shutdown()
	c.logger.Err(err).Msg("modules shutdown")

	err = c.managers.http.Shutdown()
	c.logger.Err(err).Msg("http manager shutdown")
}

func (c *serve) Command(cmd *cobra.Command, args []string) {
	c.logger.Info().Msg("starting neko server")
	c.Start()
	c.logger.Info().Msg("neko ready")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	sig := <-quit

	c.logger.Warn().Msgf("received %s, attempting graceful shutdown", sig)
	c.Shutdown()
	c.logger.Info().Msg("shutdown complete")
}
