package pi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
)

var (
	srvInst *service
	srvOnce sync.Once
)

func PI() *service {
	if srvInst == nil {
		srvOnce.Do(func() {
			exePath, err := os.Executable()
			if err != nil {
				exePath = "."
			}
			exeDir := filepath.Dir(exePath)

			srvInst = &service{
				version:    VersionUnknown,
				v2Loaded:   nil,
				config:     nil,
				configPath: filepath.Join(exeDir, "config", "interface_config.json"),
			}
		})
	}
	return srvInst
}

func Startup(ctx context.Context) {
	s := PI()

	s.ctx = ctx

	exePath, err := os.Executable()
	if err != nil {
		log.Printf("get executable path failed: %v", err)
		return
	}
	exeDir := filepath.Dir(exePath)

	ifacePath := filepath.Join(exeDir, "config", "interface.json")
	data, err := os.ReadFile(ifacePath)
	if err != nil {
		log.Printf("read interface file failed: %v", err)
		return
	}

	version, err := DetectVersion(data)
	if err != nil {
		log.Printf("detect version failed: %v", err)
		return
	}
	s.version = version

	if version == Version2 {
		v2Loaded, err := LoadV2FromFile(ifacePath)
		if err != nil {
			log.Printf("load v2 interface failed: %v", err)
			return
		}
		log.Printf("v2 interface loaded: %+v", v2Loaded)
		s.v2Loaded = v2Loaded
	} else {
		log.Printf("unknown version: %d", version)
		return
	}

	// 加载配置
	if err := s.loadConfig(); err != nil {
		log.Printf("load config failed, initializing default config: %v", err)
		s.initDefaultConfig()
		// 保存默认配置到文件
		if err := s.saveConfig(); err != nil {
			log.Printf("save default config failed: %v", err)
		}
	} else {
		// 同步配置中的 option（补全遗漏的，移除多余的）
		if s.syncConfigOptions() {
			// 如果有变更，保存配置
			if err := s.saveConfig(); err != nil {
				log.Printf("save synced config failed: %v", err)
			}
		}
	}
}

// Version represents the version of the PI protocol
type Version int

const (
	VersionUnknown Version = 0
	Version1       Version = 1
	Version2       Version = 2
)

// String returns the string representation of the version
func (v Version) String() string {
	switch v {
	case Version1:
		return "v1"
	case Version2:
		return "v2"
	default:
		return "unknown"
	}
}

// versionInfo represents the information of a version
type versionInfo struct {
	InterfaceVersion int `json:"interface_version"`
}

// DetectVersion detects the version of the interface
func DetectVersion(data []byte) (Version, error) {
	var info versionInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return VersionUnknown, fmt.Errorf("parse version information failed: %w", err)
	}

	switch info.InterfaceVersion {
	case 0, 1:
		return Version1, nil
	case 2:
		return Version2, nil
	default:
		return VersionUnknown, fmt.Errorf("unknown interface version: %d", info.InterfaceVersion)
	}
}

// DetectVersionFromFile detects the version of the interface from a file
func DetectVersionFromFile(path string) (Version, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return VersionUnknown, fmt.Errorf("read file failed: %w", err)
	}
	return DetectVersion(data)
}

// DetectVersionFromReader detects the version of the interface from a reader
func DetectVersionFromReader(r io.Reader) (Version, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return VersionUnknown, fmt.Errorf("read data failed: %w", err)
	}
	return DetectVersion(data)
}

// IsI18nString checks if the string is an internationalized string (starts with $)
func IsI18nString(s string) bool {
	return len(s) > 0 && s[0] == '$' && len(s) != 1
}

// GetI18nKey gets the key of the internationalized string
func GetI18nKey(s string) string {
	if IsI18nString(s) {
		return s[1:]
	}
	return s
}
