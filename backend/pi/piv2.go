package pi

import "encoding/json"

// V2Interface represents the interface of the v2 version
type V2Interface struct {
	InterfaceVersion         int                 `json:"interface_version"`
	Languages                map[string]string   `json:"languages,omitempty"`
	Name                     string              `json:"name"`
	Label                    string              `json:"label,omitempty"`
	Title                    string              `json:"title,omitempty"`
	Icon                     string              `json:"icon,omitempty"`
	MirrorchyanRID           string              `json:"mirrorchyan_rid,omitempty"`
	MirrorchyanMultiplatform bool                `json:"mirrorchyan_multiplatform,omitempty"`
	Github                   string              `json:"github,omitempty"`
	Version                  string              `json:"version,omitempty"`
	Contact                  string              `json:"contact,omitempty"`
	License                  string              `json:"license,omitempty"`
	Welcome                  string              `json:"welcome,omitempty"`
	Description              string              `json:"description,omitempty"`
	Controller               []V2Controller      `json:"controller,omitempty"`
	Resource                 []V2Resource        `json:"resource,omitempty"`
	Agent                    *V2Agent            `json:"agent,omitempty"`
	Task                     []V2Task            `json:"task,omitempty"`
	Option                   map[string]V2Option `json:"option,omitempty"`
}

// V2Controller represents the controller of the v2 version
type V2Controller struct {
	Name             string         `json:"name"`
	Label            string         `json:"label,omitempty"`
	Description      string         `json:"description,omitempty"`
	Icon             string         `json:"icon,omitempty"`
	Type             string         `json:"type"`
	DisplayShortSide *int           `json:"display_short_side,omitempty"`
	DisplayLongSide  *int           `json:"display_long_side,omitempty"`
	DisplayRaw       bool           `json:"display_raw,omitempty"`
	Adb              *V2AdbConfig   `json:"adb,omitempty"`
	Win32            *V2Win32Config `json:"win32,omitempty"`
}

// V2AdbConfig represents the adb config of the v2 version
type V2AdbConfig struct{}

// V2Win32Config represents the win32 config of the v2 version
type V2Win32Config struct {
	ClassRegex  string `json:"class_regex,omitempty"`
	WindowRegex string `json:"window_regex,omitempty"`
	Mouse       string `json:"mouse,omitempty"`
	Keyboard    string `json:"keyboard,omitempty"`
	Screencap   string `json:"screencap,omitempty"`
}

// V2Resource represents the resource of the v2 version
type V2Resource struct {
	Name        string   `json:"name"`
	Label       string   `json:"label,omitempty"`
	Description string   `json:"description,omitempty"`
	Icon        string   `json:"icon,omitempty"`
	Path        []string `json:"path"`
	Controller  []string `json:"controller,omitempty"`
}

// V2Agent represents the agent of the v2 version
type V2Agent struct {
	ChildExec  string   `json:"child_exec"`
	ChildArgs  []string `json:"child_args,omitempty"`
	Identifier string   `json:"identifier,omitempty"`
}

// V2Task represents the task of the v2 version
type V2Task struct {
	Name             string          `json:"name"`
	Label            string          `json:"label,omitempty"`
	Entry            string          `json:"entry"`
	DefaultCheck     bool            `json:"default_check,omitempty"`
	Description      string          `json:"description,omitempty"`
	Icon             string          `json:"icon,omitempty"`
	Resource         []string        `json:"resource,omitempty"`
	PipelineOverride json.RawMessage `json:"pipeline_override,omitempty"`
	Option           []string        `json:"option,omitempty"`
}

// V2Option represents the option of the v2 version
type V2Option struct {
	Type             string          `json:"type,omitempty"`
	Label            string          `json:"label,omitempty"`
	Description      string          `json:"description,omitempty"`
	Icon             string          `json:"icon,omitempty"`
	Cases            []V2OptionCase  `json:"cases,omitempty"`
	Inputs           []V2OptionInput `json:"inputs,omitempty"`
	PipelineOverride json.RawMessage `json:"pipeline_override,omitempty"`
	DefaultCase      string          `json:"default_case,omitempty"`
}

// GetType returns the type of the option
func (o *V2Option) GetType() string {
	if o.Type == "" {
		return "select"
	}
	return o.Type
}

// V2OptionCase represents the option case of the v2 version
type V2OptionCase struct {
	Name             string          `json:"name"`
	Label            string          `json:"label,omitempty"`
	Description      string          `json:"description,omitempty"`
	Icon             string          `json:"icon,omitempty"`
	Option           []string        `json:"option,omitempty"`
	PipelineOverride json.RawMessage `json:"pipeline_override,omitempty"`
}

// V2OptionInput represents the option input of the v2 version
type V2OptionInput struct {
	Name         string `json:"name"`
	Label        string `json:"label,omitempty"`
	Description  string `json:"description,omitempty"`
	Default      string `json:"default,omitempty"`
	PipelineType string `json:"pipeline_type,omitempty"`
	Verify       string `json:"verify,omitempty"`
	PatternMsg   string `json:"pattern_msg,omitempty"`
}
