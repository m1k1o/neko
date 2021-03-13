package neko

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"

	"demodesk/neko/internal/api"
	"demodesk/neko/internal/capture"
	"demodesk/neko/internal/config"
	"demodesk/neko/internal/desktop"
	"demodesk/neko/internal/http"
	"demodesk/neko/internal/session"
	"demodesk/neko/internal/webrtc"
	"demodesk/neko/internal/websocket"
	"demodesk/neko/modules"

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
&1&37  nurdism/m1k1o &33%s v%s&0
`

var (
	//
	buildDate = "dev"
	//
	gitCommit = "dev"
	//
	gitBranch = "dev"

	// Major version when you make incompatible API changes,
	major = "dev"
	// Minor version when you add functionality in a backwards-compatible manner, and
	minor = "dev"
	// Patch version when you make backwards-compatible bug fixeneko.
	patch = "dev"
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
		Configs: &Configs{
			Root:    &config.Root{},
			Desktop: &config.Desktop{},
			Capture: &config.Capture{},
			WebRTC:  &config.WebRTC{},
			Session: &config.Session{},
			Server:  &config.Server{},
		},
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

type Configs struct {
	Root    *config.Root
	Desktop *config.Desktop
	Capture *config.Capture
	WebRTC  *config.WebRTC
	Session *config.Session
	Server  *config.Server
}

type Neko struct {
	Version *Version
	Configs *Configs

	logger           zerolog.Logger
	desktopManager   *desktop.DesktopManagerCtx
	captureManager   *capture.CaptureManagerCtx
	webRTCManager    *webrtc.WebRTCManagerCtx
	sessionManager   *session.SessionManagerCtx
	webSocketManager *websocket.WebSocketManagerCtx
	apiManager       *api.ApiManagerCtx
	httpManager      *http.HttpManagerCtx
}

func (neko *Neko) Preflight() {
	neko.logger = log.With().Str("service", "neko").Logger()
}

func (neko *Neko) Start() {
	neko.sessionManager = session.New(
		neko.Configs.Session,
	)

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

func (neko *Neko) Shutdown() {
	if err := neko.desktopManager.Shutdown(); err != nil {
		neko.logger.Err(err).Msg("desktop manager shutdown with an error")
	} else {
		neko.logger.Debug().Msg("desktop manager shutdown")
	}

	if err := neko.captureManager.Shutdown(); err != nil {
		neko.logger.Err(err).Msg("capture manager shutdown with an error")
	} else {
		neko.logger.Debug().Msg("capture manager shutdown")
	}

	if err := neko.webRTCManager.Shutdown(); err != nil {
		neko.logger.Err(err).Msg("webrtc manager shutdown with an error")
	} else {
		neko.logger.Debug().Msg("webrtc manager shutdown")
	}

	if err := neko.webSocketManager.Shutdown(); err != nil {
		neko.logger.Err(err).Msg("websocket manager shutdown with an error")
	} else {
		neko.logger.Debug().Msg("websocket manager shutdown")
	}

	if err := modules.Shutdown(); err != nil {
		neko.logger.Err(err).Msg("modules shutdown with an error")
	} else {
		neko.logger.Debug().Msg("modules shutdown")
	}

	if err := neko.httpManager.Shutdown(); err != nil {
		neko.logger.Err(err).Msg("http manager shutdown with an error")
	} else {
		neko.logger.Debug().Msg("http manager shutdown")
	}
}

func (neko *Neko) ServeCommand(cmd *cobra.Command, args []string) {
	neko.logger.Info().Msg("starting neko server")
	neko.Start()
	neko.logger.Info().Msg("neko ready")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	sig := <-quit

	neko.logger.Warn().Msgf("received %s, attempting graceful shutdown: \n", sig)
	neko.Shutdown()
	neko.logger.Info().Msg("shutdown complete")
}
