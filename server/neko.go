package neko

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"

	"n.eko.moe/neko/internal/config"
	"n.eko.moe/neko/internal/http"
	"n.eko.moe/neko/internal/session"
	"n.eko.moe/neko/internal/webrtc"
	"n.eko.moe/neko/internal/websocket"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

const Header = `&34
    _   __     __
   / | / /__  / /______   \    /\
  /  |/ / _ \/ //_/ __ \   )  ( ')
 / /|  /  __/ ,< / /_/ /  (  /  )
/_/ |_/\___/_/|_|\____/    \(__)|
&1&37   nurdism/neko &33%s v%s&0
`

var (
	//
	buildDate = ""
	//
	gitCommit = ""
	//
	gitVersion = ""
	//
	gitState = ""
	// Major version when you make incompatible API changes,
	major = "0"
	// Minor version when you add functionality in a backwards-compatible manner, and
	minor = "0"
	// Patch version when you make backwards-compatible bug fixeneko.
	patch = "0"
)

var Service *Neko

func init() {
	Service = &Neko{
		Version: &Version{
			Major:        major,
			Minor:        minor,
			Patch:        patch,
			GitVersion:   gitVersion,
			GitCommit:    gitCommit,
			GitTreeState: gitState,
			BuildDate:    buildDate,
			GoVersion:    runtime.Version(),
			Compiler:     runtime.Compiler,
			Platform:     fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		},
		Root:      &config.Root{},
		Server:    &config.Server{},
		WebRTC:    &config.WebRTC{},
		WebSocket: &config.WebSocket{},
	}
}

type Version struct {
	Major        string
	Minor        string
	Patch        string
	Version      string
	GitVersion   string
	GitCommit    string
	GitTreeState string
	BuildDate    string
	GoVersion    string
	Compiler     string
	Platform     string
}

func (i *Version) String() string {
	return fmt.Sprintf("%s.%s.%s", i.Major, i.Minor, i.Patch)
}

type Neko struct {
	Version   *Version
	Root      *config.Root
	Server    *config.Server
	WebRTC    *config.WebRTC
	WebSocket *config.WebSocket

	logger           zerolog.Logger
	server           *http.Server
	sessions         *session.SessionManager
	webRTCManager    *webrtc.WebRTCManager
	webSocketHandler *websocket.WebSocketHandler
}

func (neko *Neko) Preflight() {
	neko.logger = log.With().Str("service", "neko").Logger()
}

func (neko *Neko) Start() {
	sessions := session.New()

	webRTCManager := webrtc.New(sessions, neko.WebRTC)
	webRTCManager.Start()

	webSocketHandler := websocket.New(sessions, webRTCManager, neko.WebSocket)
	webSocketHandler.Start()

	server := http.New(neko.Server, webSocketHandler)
	server.Start()

	neko.sessions = sessions
	neko.webRTCManager = webRTCManager
	neko.webSocketHandler = webSocketHandler
	neko.server = server
}

func (neko *Neko) Shutdown() {
	if err := neko.webRTCManager.Shutdown(); err != nil {
		neko.logger.Err(err).Msg("webrtc manager shutdown with an error")
	} else {
		neko.logger.Debug().Msg("webrtc manager shutdown")
	}

	if err := neko.webSocketHandler.Shutdown(); err != nil {
		neko.logger.Err(err).Msg("websocket handler shutdown with an error")
	} else {
		neko.logger.Debug().Msg("websocket handler shutdown")
	}

	if err := neko.server.Shutdown(); err != nil {
		neko.logger.Err(err).Msg("server shutdown with an error")
	} else {
		neko.logger.Debug().Msg("server shutdown")
	}
}

func (neko *Neko) ServeCommand(cmd *cobra.Command, args []string) {
	neko.logger.Info().Msg("starting neko server")
	neko.Start()
	neko.logger.Info().Msg("neko ready")

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	sig := <-quit

	neko.logger.Warn().Msgf("received %s, attempting graceful shutdown: \n", sig)
	neko.Shutdown()
	neko.logger.Info().Msg("shutdown complete")
}
