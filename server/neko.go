package neko

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"

	"m1k1o/neko/internal/capture"
	"m1k1o/neko/internal/config"
	"m1k1o/neko/internal/desktop"
	"m1k1o/neko/internal/http"
	"m1k1o/neko/internal/session"
	"m1k1o/neko/internal/webrtc"
	"m1k1o/neko/internal/websocket"

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
&1&37 nurdism/m1k1o &33%s v%s&0
`

var (
	//
	buildDate = "dev"
	//
	gitCommit = "dev"
	//
	gitBranch = "dev"

	// Major version when you make incompatible API changes,
	major = "2"
	// Minor version when you add functionality in a backwards-compatible manner, and
	minor = "8"
	// Patch version when you make backwards-compatible bug fixes.
	patch = "0"
)

var Service *Neko

func init() {
	Service = &Neko{
		Version: &Version{
			Major:     major,
			Minor:     minor,
			Patch:     patch,
			GitCommit: gitCommit,
			GitBranch: gitBranch,
			BuildDate: buildDate,
			GoVersion: runtime.Version(),
			Compiler:  runtime.Compiler,
			Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		},
		Root:      &config.Root{},
		Server:    &config.Server{},
		Capture:   &config.Capture{},
		Desktop:   &config.Desktop{},
		WebRTC:    &config.WebRTC{},
		WebSocket: &config.WebSocket{},
	}
}

type Version struct {
	Major     string
	Minor     string
	Patch     string
	GitCommit string
	GitBranch string
	BuildDate string
	GoVersion string
	Compiler  string
	Platform  string
}

func (i *Version) String() string {
	return fmt.Sprintf("%s.%s.%s %s", i.Major, i.Minor, i.Patch, i.GitCommit)
}

func (i *Version) Details() string {
	return fmt.Sprintf(
		"%s\n%s\n%s\n%s\n%s\n%s\n%s\n",
		fmt.Sprintf("Version %s.%s.%s", i.Major, i.Minor, i.Patch),
		fmt.Sprintf("GitCommit %s", i.GitCommit),
		fmt.Sprintf("GitBranch %s", i.GitBranch),
		fmt.Sprintf("BuildDate %s", i.BuildDate),
		fmt.Sprintf("GoVersion %s", i.GoVersion),
		fmt.Sprintf("Compiler %s", i.Compiler),
		fmt.Sprintf("Platform %s", i.Platform),
	)
}

type Neko struct {
	Version   *Version
	Root      *config.Root
	Capture   *config.Capture
	Desktop   *config.Desktop
	Server    *config.Server
	WebRTC    *config.WebRTC
	WebSocket *config.WebSocket

	logger           zerolog.Logger
	server           *http.Server
	sessionManager   *session.SessionManager
	captureManager   *capture.CaptureManagerCtx
	desktopManager   *desktop.DesktopManagerCtx
	webRTCManager    *webrtc.WebRTCManager
	webSocketHandler *websocket.WebSocketHandler
}

func (neko *Neko) Preflight() {
	neko.logger = log.With().Str("service", "neko").Logger()
}

func (neko *Neko) Start() {
	desktopManager := desktop.New(neko.Desktop)
	desktopManager.Start()

	captureManager := capture.New(desktopManager, neko.Capture)
	captureManager.Start()

	sessionManager := session.New(captureManager)

	webRTCManager := webrtc.New(sessionManager, captureManager, desktopManager, neko.WebRTC)
	webRTCManager.Start()

	webSocketHandler := websocket.New(sessionManager, desktopManager, captureManager, webRTCManager, neko.WebSocket)
	webSocketHandler.Start()

	server := http.New(neko.Server, webSocketHandler, desktopManager)
	server.Start()

	neko.sessionManager = sessionManager
	neko.captureManager = captureManager
	neko.desktopManager = desktopManager
	neko.webRTCManager = webRTCManager
	neko.webSocketHandler = webSocketHandler
	neko.server = server
}

func (neko *Neko) Shutdown() {
	var err error

	err = neko.server.Shutdown()
	neko.logger.Err(err).Msg("server shutdown")

	err = neko.webSocketHandler.Shutdown()
	neko.logger.Err(err).Msg("websocket handler shutdown")

	err = neko.webRTCManager.Shutdown()
	neko.logger.Err(err).Msg("webrtc manager shutdown")

	err = neko.captureManager.Shutdown()
	neko.logger.Err(err).Msg("capture manager shutdown")

	err = neko.desktopManager.Shutdown()
	neko.logger.Err(err).Msg("desktop manager shutdown")
}

func (neko *Neko) ServeCommand(cmd *cobra.Command, args []string) {
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
