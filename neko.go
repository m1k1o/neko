package neko

import (
	"fmt"
	"runtime"

	"demodesk/neko/internal/config"
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

	// Major version when you make incompatible API changes.
	major = "dev"
	// Minor version when you add functionality in a backwards-compatible manner.
	minor = "dev"
	// Patch version when you make backwards-compatible bug fixes.
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
			Member:  &config.Member{},
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
	Member  *config.Member
	Session *config.Session
	Server  *config.Server
}

type Neko struct {
	Version *Version
	Configs *Configs
}
