package pi

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

// ParseV2 parses the data into a V2Interface
func ParseV2(data []byte) (*V2Interface, error) {
	var iface V2Interface
	if err := json.Unmarshal(data, &iface); err != nil {
		return nil, fmt.Errorf("parse JSON failed: %w", err)
	}

	if err := validateV2(&iface); err != nil {
		return nil, err
	}

	return &iface, nil
}

// ParseV2File parses the file into a V2Interface
func ParseV2File(path string) (*V2Interface, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file failed: %w", err)
	}
	return ParseV2(data)
}

// validateV2 validates the V2Interface
func validateV2(iface *V2Interface) error {
	if iface.InterfaceVersion != 2 {
		return fmt.Errorf("version mismatch: expected 2, got %d", iface.InterfaceVersion)
	}

	if iface.Name == "" {
		return fmt.Errorf("missing required field: name")
	}

	// validate controllers
	controllerNames := make(map[string]bool)
	for i, ctrl := range iface.Controller {
		if ctrl.Name == "" {
			return fmt.Errorf("controller[%d]: missing name", i)
		}
		if controllerNames[ctrl.Name] {
			return fmt.Errorf("controller[%d]: duplicate name: %s", i, ctrl.Name)
		}
		controllerNames[ctrl.Name] = true

		if ctrl.Type != "Adb" && ctrl.Type != "Win32" {
			return fmt.Errorf("controller[%d]: invalid type: %s", i, ctrl.Type)
		}

		// validate exclusive fields
		count := 0
		if ctrl.DisplayShortSide != nil {
			count++
		}
		if ctrl.DisplayLongSide != nil {
			count++
		}
		if ctrl.DisplayRaw {
			count++
		}
		if count > 1 {
			return fmt.Errorf("controller[%d]: display options are exclusive", i)
		}
	}

	// validate resources
	resourceNames := make(map[string]bool)
	for i, res := range iface.Resource {
		if res.Name == "" {
			return fmt.Errorf("resource[%d]: missing name", i)
		}
		if resourceNames[res.Name] {
			return fmt.Errorf("resource[%d]: duplicate name: %s", i, res.Name)
		}
		resourceNames[res.Name] = true

		if len(res.Path) == 0 {
			return fmt.Errorf("resource[%d]: missing path", i)
		}

		for _, ctrlName := range res.Controller {
			if !controllerNames[ctrlName] {
				return fmt.Errorf("resource[%d]: reference to non-existent controller: %s", i, ctrlName)
			}
		}
	}

	// validate agent
	if iface.Agent != nil && iface.Agent.ChildExec == "" {
		return fmt.Errorf("agent: missing child_exec")
	}

	// validate tasks
	for i, task := range iface.Task {
		if task.Name == "" {
			return fmt.Errorf("task[%d]: missing name", i)
		}
		if task.Entry == "" {
			return fmt.Errorf("task[%d]: missing entry", i)
		}

		for _, resName := range task.Resource {
			if !resourceNames[resName] {
				return fmt.Errorf("task[%d]: reference to non-existent resource: %s", i, resName)
			}
		}

		for _, optName := range task.Option {
			if _, ok := iface.Option[optName]; !ok {
				return fmt.Errorf("task[%d]: reference to non-existent option: %s", i, optName)
			}
		}
	}

	// validate options
	for name, opt := range iface.Option {
		optType := opt.GetType()

		switch optType {
		case "select", "switch":
			if len(opt.Cases) == 0 {
				return fmt.Errorf("option[%s]: missing cases", name)
			}
			if optType == "switch" && len(opt.Cases) != 2 {
				return fmt.Errorf("option[%s]: switch must have 2 cases", name)
			}

			caseNames := make(map[string]bool)
			for j, c := range opt.Cases {
				if c.Name == "" {
					return fmt.Errorf("option[%s].cases[%d]: missing name", name, j)
				}
				caseNames[c.Name] = true

				for _, subOpt := range c.Option {
					if _, ok := iface.Option[subOpt]; !ok {
						return fmt.Errorf("option[%s].cases[%d]: reference to non-existent option: %s", name, j, subOpt)
					}
				}
			}

			if opt.DefaultCase != "" && !caseNames[opt.DefaultCase] {
				return fmt.Errorf("option[%s]: default_case does not exist: %s", name, opt.DefaultCase)
			}

		case "input":
			if len(opt.Inputs) == 0 {
				return fmt.Errorf("option[%s]: missing inputs", name)
			}

			for j, input := range opt.Inputs {
				if input.Name == "" {
					return fmt.Errorf("option[%s].inputs[%d]: missing name", name, j)
				}
				if input.Verify != "" {
					if _, err := regexp.Compile(input.Verify); err != nil {
						return fmt.Errorf("option[%s].inputs[%d]: invalid regex: %w", name, j, err)
					}
				}
			}

		default:
			return fmt.Errorf("option[%s]: invalid type: %s", name, optType)
		}
	}

	return nil
}

// V2I18nResolver represents the internationalization resolver
type V2I18nResolver struct {
	translations map[string]string
}

// NewV2I18nResolver creates a new internationalization resolver
func NewV2I18nResolver(translationFile string) (*V2I18nResolver, error) {
	data, err := os.ReadFile(translationFile)
	if err != nil {
		return nil, fmt.Errorf("read translation file failed: %w", err)
	}

	var translations map[string]string
	if err := json.Unmarshal(data, &translations); err != nil {
		return nil, fmt.Errorf("parse translation file failed: %w", err)
	}

	return &V2I18nResolver{translations: translations}, nil
}

// Resolve resolves the internationalization string
func (r *V2I18nResolver) Resolve(s string) string {
	if !IsI18nString(s) {
		return s
	}
	key := GetI18nKey(s)
	if val, ok := r.translations[key]; ok {
		return val
	}
	return s
}

// V2Loaded represents the loaded V2Interface with i18n support
type V2Loaded struct {
	Interface *V2Interface
	Resolvers map[string]*V2I18nResolver
	BasePath  string
}

// LoadV2FromFile loads the V2Interface from a file and associates the translation files
func LoadV2FromFile(path string) (*V2Loaded, error) {
	iface, err := ParseV2File(path)
	if err != nil {
		return nil, err
	}

	basePath := filepath.Dir(path)
	resolvers := make(map[string]*V2I18nResolver)

	for lang, transPath := range iface.Languages {
		fullPath := filepath.Join(basePath, transPath)
		resolver, err := NewV2I18nResolver(fullPath)
		if err != nil {
			continue
		}
		resolvers[lang] = resolver
	}

	return &V2Loaded{
		Interface: iface,
		Resolvers: resolvers,
		BasePath:  basePath,
	}, nil
}

// GetLanguages returns the list of supported languages
func (l *V2Loaded) GetLanguages() []string {
	langs := make([]string, 0, len(l.Interface.Languages))
	for lang := range l.Interface.Languages {
		langs = append(langs, lang)
	}
	return langs
}

// ResolveString resolves the internationalization string using the specified language
func (l *V2Loaded) ResolveString(s string, lang string) string {
	if !IsI18nString(s) {
		return s
	}
	if resolver, ok := l.Resolvers[lang]; ok {
		return resolver.Resolve(s)
	}
	return GetI18nKey(s)
}
