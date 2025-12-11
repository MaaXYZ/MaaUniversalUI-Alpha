package engine

import (
	"context"
	"muu-alpha/backend/pi"
	"os"
	"path/filepath"
	"sync"

	"github.com/MaaXYZ/maa-framework-go/v3"
)

var (
	srvInst *service
	srvOnce sync.Once
)

func Engine() *service {
	srvOnce.Do(func() {
		srvInst = &service{}
	})
	return srvInst
}

func Startup(ctx context.Context) {
	exePath, err := os.Executable()
	if err != nil {
		panic(err) // todo
	}
	exeDir := filepath.Dir(exePath)
	libDir := filepath.Join(exeDir, "lib")
	logDir := filepath.Join(exeDir, "log")

	maa.Init(
		maa.WithLibDir(libDir),
		maa.WithLogDir(logDir),
	)
	s := Engine()
	s.ctx = ctx
}

func GetPathsByResName(name string) []string {
	piSrv := pi.PI()
	v2Loaded := piSrv.V2Loaded()

	if v2Loaded == nil || v2Loaded.Interface == nil {
		return []string{}
	}

	iface := v2Loaded.Interface

	for _, res := range iface.Resource {
		if res.Name == name {
			return res.Path
		}
	}

	return []string{}
}
