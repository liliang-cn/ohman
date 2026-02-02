package version

import "fmt"

// Version variables, injected via ldflags
var (
	Version = "v0.4.8"
	Commit    = "unknown"
	BuildTime = "unknown"
)

// String returns a formatted version string
func String() string {
	return fmt.Sprintf("%s (commit: %s, built: %s)", Version, Commit, BuildTime)
}

// Short returns the short version
func Short() string {
	return Version
}
