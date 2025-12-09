package pi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type service struct {
	ctx        context.Context
	version    Version
	v2Loaded   *V2Loaded
	config     *InterfaceConfig
	configPath string
	configMu   sync.RWMutex
}

func (s *service) GetVersion() int {
	return int(s.version)
}

func (s *service) V2Loaded() *V2Loaded {
	return s.v2Loaded
}

// initDefaultConfig initializes default config from PI data
func (s *service) initDefaultConfig() {
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
		// Initialize options with default values
		options := s.initTaskOptions(task.Option, iface.Option)

		config.Task = append(config.Task, ConfigTask{
			ID:      uuid.New().String(),
			Name:    task.Name,
			Checked: task.DefaultCheck,
			Option:  options,
		})
	}

	s.config = config
}

// initTaskOptions initializes option values for a task
func (s *service) initTaskOptions(optionNames []string, optionDefs map[string]V2Option) []ConfigTaskOption {
	if len(optionNames) == 0 || optionDefs == nil {
		return []ConfigTaskOption{}
	}

	options := []ConfigTaskOption{}
	s.collectOptionDefaults(&options, optionNames, optionDefs)
	return options
}

// collectOptionDefaults recursively collects default values for options
func (s *service) collectOptionDefaults(options *[]ConfigTaskOption, optionNames []string, optionDefs map[string]V2Option) {
	for _, optName := range optionNames {
		optDef, exists := optionDefs[optName]
		if !exists {
			continue
		}

		optType := optDef.GetType()

		switch optType {
		case "select":
			// For select type, use default_case or first case
			var selectedValue string
			var selectedCase *V2OptionCase

			if optDef.DefaultCase != "" {
				selectedValue = optDef.DefaultCase
				for i := range optDef.Cases {
					if optDef.Cases[i].Name == optDef.DefaultCase {
						selectedCase = &optDef.Cases[i]
						break
					}
				}
			} else if len(optDef.Cases) > 0 {
				selectedValue = optDef.Cases[0].Name
				selectedCase = &optDef.Cases[0]
			}

			if selectedValue != "" {
				*options = append(*options, ConfigTaskOption{
					Name:  optName,
					Value: selectedValue,
				})
			}

			// Recursively collect nested options
			if selectedCase != nil && len(selectedCase.Option) > 0 {
				s.collectOptionDefaults(options, selectedCase.Option, optionDefs)
			}

		case "switch":
			// For switch type, default to No case
			var selectedValue string
			var selectedCase *V2OptionCase

			for i := range optDef.Cases {
				caseName := optDef.Cases[i].Name
				if caseName != "Yes" && caseName != "yes" && caseName != "Y" && caseName != "y" {
					selectedValue = caseName
					selectedCase = &optDef.Cases[i]
					break
				}
			}

			if selectedValue == "" && len(optDef.Cases) > 0 {
				selectedValue = "No"
			}

			if selectedValue != "" {
				*options = append(*options, ConfigTaskOption{
					Name:  optName,
					Value: selectedValue,
				})
			}

			// Recursively collect nested options
			if selectedCase != nil && len(selectedCase.Option) > 0 {
				s.collectOptionDefaults(options, selectedCase.Option, optionDefs)
			}

		case "input":
			// For input type, set default values for all inputs
			for _, input := range optDef.Inputs {
				value := input.Default
				if value == "" {
					// Use type-based default
					switch input.PipelineType {
					case "bool":
						value = "false"
					case "int":
						value = "0"
					default:
						value = ""
					}
				}

				*options = append(*options, ConfigTaskOption{
					Name:  optName + "." + input.Name,
					Value: value,
				})
			}
		}
	}
}

// loadConfig loads config from file
func (s *service) loadConfig() error {
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

// syncConfigOptions syncs options in config with PI definitions
// Returns true if any changes were made
func (s *service) syncConfigOptions() bool {
	if s.config == nil || s.v2Loaded == nil || s.v2Loaded.Interface == nil {
		return false
	}

	s.configMu.Lock()
	defer s.configMu.Unlock()

	iface := s.v2Loaded.Interface
	changed := false

	for i := range s.config.Task {
		task := &s.config.Task[i]

		// Find corresponding PI task
		var piTask *V2Task
		for j := range iface.Task {
			if iface.Task[j].Name == task.Name {
				piTask = &iface.Task[j]
				break
			}
		}

		if piTask == nil {
			continue
		}

		// Sync options for this task
		if s.syncTaskConfigOptions(task, piTask.Option, iface.Option) {
			changed = true
		}
	}

	return changed
}

// syncTaskConfigOptions syncs options for a single task
// Returns true if any changes were made
func (s *service) syncTaskConfigOptions(task *ConfigTask, optionNames []string, optionDefs map[string]V2Option) bool {
	if optionDefs == nil {
		return false
	}

	// Build map of current options for quick lookup
	currentOptions := make(map[string]string)
	for _, opt := range task.Option {
		currentOptions[opt.Name] = opt.Value
	}

	// Get expected options based on current selections
	expectedOptions := s.getExpectedOptions(optionNames, currentOptions, optionDefs)

	// Check for missing options
	missingOptions := []ConfigTaskOption{}
	for _, expected := range expectedOptions {
		if _, exists := currentOptions[expected.Name]; !exists {
			missingOptions = append(missingOptions, expected)
		}
	}

	// Check for extra options
	expectedNames := make(map[string]bool)
	for _, expected := range expectedOptions {
		expectedNames[expected.Name] = true
	}

	extraOptions := []string{}
	for name := range currentOptions {
		if !expectedNames[name] {
			extraOptions = append(extraOptions, name)
		}
	}

	// No changes needed
	if len(missingOptions) == 0 && len(extraOptions) == 0 {
		return false
	}

	// Build new options list
	newOptions := []ConfigTaskOption{}

	// Keep existing options that are expected
	for _, opt := range task.Option {
		if expectedNames[opt.Name] {
			newOptions = append(newOptions, opt)
		}
	}

	// Add missing options
	newOptions = append(newOptions, missingOptions...)

	task.Option = newOptions
	return true
}

// getExpectedOptions returns expected options based on current selections
func (s *service) getExpectedOptions(optionNames []string, currentValues map[string]string, optionDefs map[string]V2Option) []ConfigTaskOption {
	expected := []ConfigTaskOption{}
	s.collectExpectedOptions(&expected, optionNames, currentValues, optionDefs)
	return expected
}

// collectExpectedOptions recursively collects expected options
func (s *service) collectExpectedOptions(expected *[]ConfigTaskOption, optionNames []string, currentValues map[string]string, optionDefs map[string]V2Option) {
	for _, optName := range optionNames {
		optDef, exists := optionDefs[optName]
		if !exists {
			continue
		}

		optType := optDef.GetType()

		switch optType {
		case "select":
			// Get current or default value
			selectedValue := currentValues[optName]
			var selectedCase *V2OptionCase

			if selectedValue == "" {
				// Use default
				if optDef.DefaultCase != "" {
					selectedValue = optDef.DefaultCase
				} else if len(optDef.Cases) > 0 {
					selectedValue = optDef.Cases[0].Name
				}
			}

			// Find selected case
			for i := range optDef.Cases {
				if optDef.Cases[i].Name == selectedValue {
					selectedCase = &optDef.Cases[i]
					break
				}
			}

			if selectedValue != "" {
				*expected = append(*expected, ConfigTaskOption{
					Name:  optName,
					Value: selectedValue,
				})
			}

			// Recursively collect nested options
			if selectedCase != nil && len(selectedCase.Option) > 0 {
				s.collectExpectedOptions(expected, selectedCase.Option, currentValues, optionDefs)
			}

		case "switch":
			// Get current or default value
			selectedValue := currentValues[optName]
			var selectedCase *V2OptionCase

			if selectedValue == "" {
				// Default to No case
				for i := range optDef.Cases {
					caseName := optDef.Cases[i].Name
					if caseName != "Yes" && caseName != "yes" && caseName != "Y" && caseName != "y" {
						selectedValue = caseName
						selectedCase = &optDef.Cases[i]
						break
					}
				}
				if selectedValue == "" {
					selectedValue = "No"
				}
			} else {
				// Find selected case
				for i := range optDef.Cases {
					if optDef.Cases[i].Name == selectedValue {
						selectedCase = &optDef.Cases[i]
						break
					}
				}
			}

			if selectedValue != "" {
				*expected = append(*expected, ConfigTaskOption{
					Name:  optName,
					Value: selectedValue,
				})
			}

			// Recursively collect nested options
			if selectedCase != nil && len(selectedCase.Option) > 0 {
				s.collectExpectedOptions(expected, selectedCase.Option, currentValues, optionDefs)
			}

		case "input":
			// For input type, add all input fields
			for _, input := range optDef.Inputs {
				key := optName + "." + input.Name
				value := currentValues[key]

				if value == "" {
					// Use default
					value = input.Default
					if value == "" {
						switch input.PipelineType {
						case "bool":
							value = "false"
						case "int":
							value = "0"
						default:
							value = ""
						}
					}
				}

				*expected = append(*expected, ConfigTaskOption{
					Name:  key,
					Value: value,
				})
			}
		}
	}
}

// saveConfig saves config to file
func (s *service) saveConfig() error {
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
func (s *service) GetConfig() *InterfaceConfig {
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
func (s *service) SaveConfig(config *InterfaceConfig) error {
	s.configMu.Lock()
	s.config = config
	s.configMu.Unlock()

	return s.saveConfig()
}

// ReadContent reads content from a file path, URL, or returns direct text
// Supports:
// - File path (relative to interface.json directory)
// - URL (http:// or https://)
// - Direct text (returned as-is)
func (s *service) ReadContent(content string) string {
	if content == "" {
		return ""
	}

	// Check if it's a URL
	if strings.HasPrefix(content, "http://") || strings.HasPrefix(content, "https://") {
		return s.readFromURL(content)
	}

	// Check if it's a file path (relative to base path)
	if s.v2Loaded != nil && s.v2Loaded.BasePath != "" {
		filePath := filepath.Join(s.v2Loaded.BasePath, content)
		if data, err := os.ReadFile(filePath); err == nil {
			return string(data)
		}
	}

	// Return as direct text
	return content
}

// readFromURL reads content from a URL
func (s *service) readFromURL(url string) string {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		log.Printf("failed to fetch URL %s: %v", url, err)
		return url // Return URL as fallback
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("failed to fetch URL %s: status %d", url, resp.StatusCode)
		return url
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("failed to read response body from %s: %v", url, err)
		return url
	}

	return string(data)
}
