package appconf

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"sync"
)

var (
	appcSrvInst *service
	appcSrvOnce sync.Once
)

func AppConf() *service {
	appcSrvOnce.Do(func() {
		exePath, err := os.Executable()
		if err != nil {
			exePath = "."
		}
		exeDir := filepath.Dir(exePath)
		configDir := filepath.Join(exeDir, "config")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			log.Printf("create config directory failed: %v", err)
		}

		appcSrvInst = &service{
			config:     nil,
			configPath: filepath.Join(configDir, "app_config.json"),
		}
	})
	return appcSrvInst
}

func Startup(ctx context.Context) {
	s := AppConf()
	s.ctx = ctx

	if err := s.loadConfig(); err != nil {
		log.Printf("load app config failed, initializing default config: %v", err)
		s.config = DefaultAppConfig()

		if err := s.saveConfig(); err != nil {
			log.Printf("save default app config failed: %v", err)
		}
	}

	log.Printf("app config loaded: theme=%s, language=%s", s.config.Theme, s.config.Language)
}
