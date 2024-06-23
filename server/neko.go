package neko

import (
	"fmt"
	"runtime"
	"strings"
)

const Header = `&34
    _   __     __
   / | / /__  / /______   \    /\
  /  |/ / _ \/ //_/ __ \   )  ( ')
 / /|  /  __/ ,< / /_/ /  (  /  )
/_/ |_/\___/_/|_|\____/    \(__)|
&1&37  nurdism/m1k1o &33%s %s&0
`

var (
	//
	buildDate = "dev"
	//
	gitCommit = "dev"
	//
	gitBranch = "dev"
	//
	gitTag = "dev"
)

var Version = &version{
	GitCommit: gitCommit,
	GitBranch: gitBranch,
	GitTag:    gitTag,
	BuildDate: buildDate,
	GoVersion: runtime.Version(),
	Compiler:  runtime.Compiler,
	Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
}

type version struct {
	GitCommit string
	GitBranch string
	GitTag    string
	BuildDate string
	GoVersion string
	Compiler  string
	Platform  string
}

func (i *version) String() string {
	version := i.GitTag
	if version == "" || version == "dev" {
		version = i.GitBranch
	}

	return fmt.Sprintf("%s@%s", version, i.GitCommit)
}

func (i *version) Details() string {
	return "\n" + strings.Join([]string{
		fmt.Sprintf("Version %s", i.String()),
		fmt.Sprintf("GitCommit %s", i.GitCommit),
		fmt.Sprintf("GitBranch %s", i.GitBranch),
		fmt.Sprintf("GitTag %s", i.GitTag),
		fmt.Sprintf("BuildDate %s", i.BuildDate),
		fmt.Sprintf("GoVersion %s", i.GoVersion),
		fmt.Sprintf("Compiler %s", i.Compiler),
		fmt.Sprintf("Platform %s", i.Platform),
	}, "\n") + "\n"
}
