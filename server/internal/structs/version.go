package structs

import "fmt"

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
