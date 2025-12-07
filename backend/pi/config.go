package pi

// ConfigController controller config
type ConfigController struct {
	Name string `json:"name"`
	Type string `json:"type"` // "Adb" | "Win32"
}

// ConfigAdb adb config
type ConfigAdb struct {
	AdbPath string                 `json:"adb_path,omitempty"`
	Address string                 `json:"address,omitempty"`
	Config  map[string]interface{} `json:"config,omitempty"`
}

// ConfigWin32 win32 config
type ConfigWin32 struct {
	// todo
}

// ConfigTaskOption task option
type ConfigTaskOption struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ConfigTask task config
type ConfigTask struct {
	ID      string             `json:"id"`
	Name    string             `json:"name"`
	Checked bool               `json:"checked"`
	Option  []ConfigTaskOption `json:"option,omitempty"`
}

// InterfaceConfig interface config
type InterfaceConfig struct {
	Controller ConfigController `json:"controller"`
	Adb        *ConfigAdb       `json:"adb,omitempty"`
	Win32      *ConfigWin32     `json:"win32,omitempty"`
	Resource   string           `json:"resource"`
	Task       []ConfigTask     `json:"task"`
}
