package system

import (
	"context"
	goruntime "runtime"
)

var (
	AppVersion = "0.0.0-dev"
	BuildAt    = "unknown"
)

type service struct {
	ctx context.Context
}

type AppInfo struct {
	Version   string `json:"version"`
	BuildAt   string `json:"build_at"`
	BuildOS   string `json:"build_os"`
	BuildArch string `json:"build_arch"`
	GoVersion string `json:"go_version"`
}

func (s *service) GetAppInfo() AppInfo {
	return AppInfo{
		Version:   AppVersion,
		BuildAt:   BuildAt,
		BuildOS:   goruntime.GOOS,
		BuildArch: goruntime.GOARCH,
		GoVersion: goruntime.Version(),
	}
}
