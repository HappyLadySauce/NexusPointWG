package environment

const (
	userAgent = "NexusPointWG"
	release   = "1.1.0"
	dev       = "1.1.0-dev"
)

var (
	Version      = dev             // Version of this binary
	gitVersion   = "v0.0.0-master" // nolint:unused
	gitCommit    = "unknown"       // nolint:unused // sha1 from git, output of $(git rev-parse HEAD)
	gitTreeState = "unknown"       // nolint:unused // state of git tree, either "clean" or "dirty"
	buildDate    = "unknown"       // nolint:unused // build date in ISO8601 format, output of $(date -u +'%Y-%m-%dT%H:%M:%SZ')
)

// IsDev returns true if the version is dev.
func IsDev() bool {
	return Version == dev
}