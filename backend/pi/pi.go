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

	"github.com/google/uuid"
)

type piService struct {
	ctx        context.Context
	version    Version
	v2Loaded   *V2Loaded
	config     *InterfaceConfig
	configPath string
	configMu   sync.RWMutex
}

var (
	piSrvInst *piService
	piSrvOnce sync.Once
)

func PI() *piService {
	if piSrvInst == nil {
		piSrvOnce.Do(func() {
			exePath, err := os.Executable()
			if err != nil {
				exePath = "."
			}
			exeDir := filepath.Dir(exePath)

			piSrvInst = &piService{
				version:    VersionUnknown,
				v2Loaded:   nil,
				config:     nil,
				configPath: filepath.Join(exeDir, "interface_config.json"),
			}
		})
	}
	return piSrvInst
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

	ifacePath := filepath.Join(exeDir, "interface.json")
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
		v2Loaded, err := LoadV2FromFile(filepath.Join(exeDir, "interface.json"))
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
	}
}

func (s *piService) GetVersion() int {
	return int(s.version)
}

func (s *piService) V2Loaded() *V2Loaded {
	return s.v2Loaded
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

// initDefaultConfig initializes default config from PI data
func (s *piService) initDefaultConfig() {
	s.configMu.Lock()
	defer s.configMu.Unlock()

	config := &InterfaceConfig{
		Controller: ConfigController{},
		Resource:   "",
		Task:       []ConfigTask{},
	}

	// if v2 data is not loaded, use empty config
	if s.v2Loaded == nil || s.v2Loaded.Interface == nil {
		s.config = config
		return
	}

	iface := s.v2Loaded.Interface

	// use the first controller
	if len(iface.Controller) > 0 {
		firstCtrl := iface.Controller[0]
		config.Controller = ConfigController{
			Name: firstCtrl.Name,
			Type: firstCtrl.Type,
		}
	}

	// use the first resource
	if len(iface.Resource) > 0 {
		config.Resource = iface.Resource[0].Name
	}

	// add all tasks, set DefaultCheck to checked
	for _, task := range iface.Task {
		config.Task = append(config.Task, ConfigTask{
			ID:      uuid.New().String(),
			Name:    task.Name,
			Checked: task.DefaultCheck,
			Option:  []ConfigTaskOption{},
		})
	}

	s.config = config
}

// loadConfig loads config from file
func (s *piService) loadConfig() error {
	s.configMu.Lock()
	defer s.configMu.Unlock()

	data, err := os.ReadFile(s.configPath)
	if err != nil {
		return fmt.Errorf("read config file failed: %w", err)
	}

	var config InterfaceConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("parse config file failed: %w", err)
	}

	s.config = &config
	return nil
}

// saveConfig saves config to file
func (s *piService) saveConfig() error {
	if s.config == nil {
		return fmt.Errorf("config is nil")
	}

	data, err := json.MarshalIndent(s.config, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config failed: %w", err)
	}

	if err := os.WriteFile(s.configPath, data, 0644); err != nil {
		return fmt.Errorf("write config file failed: %w", err)
	}

	return nil
}

// GetConfig gets the full config
func (s *piService) GetConfig() *InterfaceConfig {
	s.configMu.RLock()
	defer s.configMu.RUnlock()

	if s.config == nil {
		return &InterfaceConfig{
			Controller: ConfigController{},
			Task:       []ConfigTask{},
		}
	}
	return s.config
}

// SaveConfig saves the full config
func (s *piService) SaveConfig(config *InterfaceConfig) error {
	s.configMu.Lock()
	s.config = config
	s.configMu.Unlock()

	return s.saveConfig()
}
