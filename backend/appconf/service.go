package appconf

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type service struct {
	ctx        context.Context
	config     *AppConfig
	configPath string
	configMu   sync.RWMutex
}

// loadConfig loads config from file
func (s *service) loadConfig() error {
	s.configMu.Lock()
	defer s.configMu.Unlock()

	data, err := os.ReadFile(s.configPath)
	if err != nil {
		return fmt.Errorf("read config file failed: %w", err)
	}

	var config AppConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("parse config file failed: %w", err)
	}

	s.checkConfig(&config)

	s.config = &config
	return nil
}

// checkConfig checks and fixes invalid values
func (s *service) checkConfig(config *AppConfig) {
	if !isValidTheme(config.Theme) {
		config.Theme = ThemeSystem
	}
	if !isValidLanguage(config.Language) {
		config.Language = LangZhCN
	}
}

// saveConfig saves config to file
func (s *service) saveConfig() error {
	s.configMu.RLock()
	config := s.config
	s.configMu.RUnlock()

	if config == nil {
		return fmt.Errorf("config is nil")
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config failed: %w", err)
	}

	if err := os.WriteFile(s.configPath, data, 0644); err != nil {
		return fmt.Errorf("write config file failed: %w", err)
	}

	return nil
}

// ==================== frontend exposed interfaces ====================

// GetConfig gets full config
func (s *service) GetConfig() *AppConfig {
	s.configMu.RLock()
	defer s.configMu.RUnlock()

	if s.config == nil {
		return DefaultAppConfig()
	}

	// return a copy to avoid external modification
	return &AppConfig{
		Theme:    s.config.Theme,
		Language: s.config.Language,
	}
}

// SaveConfig saves full config
func (s *service) SaveConfig(config *AppConfig) error {
	if config == nil {
		return fmt.Errorf("config is nil")
	}

	// 验证
	if !isValidTheme(config.Theme) {
		return fmt.Errorf("invalid theme: %s", config.Theme)
	}
	if !isValidLanguage(config.Language) {
		return fmt.Errorf("invalid language: %s", config.Language)
	}

	s.configMu.Lock()
	s.config = &AppConfig{
		Theme:    config.Theme,
		Language: config.Language,
	}
	s.configMu.Unlock()

	return s.saveConfig()
}

// GetSupported gets supported
func (s *service) GetSupported() Supported {
	return GetSupported()
}
