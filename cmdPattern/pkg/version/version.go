package version

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"time"
)

// These variables are set during build time via linker flags.
//
//nolint:gochecknoglobals // these globals are necessary as they're populated by ldflags during build
var (
	app       = "default"
	version   = "1.0.0"
	buildTime = ""
	hash      = ""
)

// BuildInfo describes the compile time information.
type BuildInfo struct {
	// App is the app name.
	App string `json:"app,omitempty"`
	// Version is the current semver.
	Version string `json:"version,omitempty"`
	// BuildTime is the time when the build was created.
	BuildTime string `json:"build_time,omitempty"`
	// Hash is the git commit sha1.
	Hash string `json:"hash,omitempty"`
	// GoVersion is the version of the Go compiler used.
	GoVersion string `json:"go_version,omitempty"`
}

// GetBuildInfo returns build info.
func GetBuildInfo() *BuildInfo {
	return &BuildInfo{
		App:       app,
		Version:   version,
		BuildTime: buildTime,
		Hash:      hash,
		GoVersion: runtime.Version(),
	}
}

// String returns build info in full string format.
func (bi *BuildInfo) String() string {
	return fmt.Sprintf(
		"Application Name: %s \nVersion: %s\nHash: %s\nBuild Time: %s\nGo version: %s",
		bi.App, bi.Version, bi.Hash, bi.BuildTime, bi.GoVersion,
	)
}

// Short returns build info in short string format.
func (bi *BuildInfo) Short(ctx context.Context) string {
	t := int64(0)
	h := bi.Hash

	// Standard Git short hash length is 7 characters
	const gitShortHashLength = 7
	if len(h) >= gitShortHashLength {
		h = bi.Hash[:gitShortHashLength]
	}
	if bi.BuildTime != "" {
		timestamp, err := time.Parse(time.RFC3339, bi.BuildTime)
		if err != nil {
			slog.WarnContext(ctx, "could not convert date to timestamp", "error", err)
		} else {
			t = timestamp.Unix()
		}
	}

	return fmt.Sprintf("%s+h.%s-t.%d", bi.Version, h, t)
}
