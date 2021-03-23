package version

import (
	"fmt"
	"runtime"
)

var (
	GitCommit   string
	Built       string
	GoVersion   = runtime.Version()
	DiceVersion string
)

func String() string {
	if len(GitCommit) > 12 {
		GitCommit = GitCommit[0:12]
	}
	return fmt.Sprintf("DiceVersion: %s, GitCommit: %s, Built: %s, GoVersion: %s",
		DiceVersion, GitCommit, Built, GoVersion)
}
