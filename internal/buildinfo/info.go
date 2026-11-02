package buildinfo

type Info struct {
	Version   string
	Commit    string
	BuildTime string
}

var (
	Version   = "dev"
	Commit    = "unknown"
	BuildTime = "unknown"
)

func Current() Info {
	return Info{Version: Version, Commit: Commit, BuildTime: BuildTime}
}
