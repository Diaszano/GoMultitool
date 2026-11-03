// Package buildinfo holds version metadata for the multitool executable.
package buildinfo

// Info describes the version and build provenance of the executable.
type Info struct {
	Version   string
	Commit    string
	BuildTime string
}

var (
	// Version is set by release builds.
	Version = "dev"
	// Commit is the source revision used for the build.
	Commit = "unknown"
	// BuildTime is the build timestamp.
	BuildTime = "unknown"
)

// Current returns the active build metadata.
func Current() Info {
	return Info{Version: Version, Commit: Commit, BuildTime: BuildTime}
}
