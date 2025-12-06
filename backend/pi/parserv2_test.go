package pi

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseV2(t *testing.T) {
	validV2 := `{
		"interface_version": 2,
		"name": "TestProject",
		"version": "1.0.0",
		"controller": [
			{
				"name": "Android",
				"type": "Adb"
			}
		],
		"resource": [
			{
				"name": "Default",
				"path": ["resource"]
			}
		],
		"task": [
			{
				"name": "StartUp",
				"entry": "StartUp"
			}
		]
	}`

	t.Run("valid v2 interface", func(t *testing.T) {
		iface, err := ParseV2([]byte(validV2))
		require.NoError(t, err)
		require.Equal(t, "TestProject", iface.Name)
		require.Equal(t, "1.0.0", iface.Version)
		require.Equal(t, 1, len(iface.Controller))
		require.Equal(t, 1, len(iface.Resource))
		require.Equal(t, 1, len(iface.Task))
	})

	t.Run("missing name", func(t *testing.T) {
		data := `{"interface_version": 2}`
		_, err := ParseV2([]byte(data))
		require.Error(t, err)
	})

	t.Run("wrong version", func(t *testing.T) {
		data := `{"interface_version": 1, "name": "Test"}`
		_, err := ParseV2([]byte(data))
		require.Error(t, err)
	})

	t.Run("invalid controller type", func(t *testing.T) {
		data := `{
			"interface_version": 2,
			"name": "Test",
			"controller": [{"name": "Test", "type": "Invalid"}]
		}`
		_, err := ParseV2([]byte(data))
		require.Error(t, err)
	})

	t.Run("duplicate controller name", func(t *testing.T) {
		data := `{
			"interface_version": 2,
			"name": "Test",
			"controller": [
				{"name": "Dup", "type": "Adb"},
				{"name": "Dup", "type": "Win32"}
			]
		}`
		_, err := ParseV2([]byte(data))
		require.Error(t, err)
	})

	t.Run("exclusive display options", func(t *testing.T) {
		data := `{
			"interface_version": 2,
			"name": "Test",
			"controller": [{
				"name": "Test",
				"type": "Adb",
				"display_short_side": 720,
				"display_long_side": 1280
			}]
		}`
		_, err := ParseV2([]byte(data))
		require.Error(t, err)
	})

	t.Run("resource missing path", func(t *testing.T) {
		data := `{
			"interface_version": 2,
			"name": "Test",
			"resource": [{"name": "Res"}]
		}`
		_, err := ParseV2([]byte(data))
		require.Error(t, err)
	})

	t.Run("task missing entry", func(t *testing.T) {
		data := `{
			"interface_version": 2,
			"name": "Test",
			"task": [{"name": "Task"}]
		}`
		_, err := ParseV2([]byte(data))
		require.Error(t, err)
	})

	t.Run("task references non-existent option", func(t *testing.T) {
		data := `{
			"interface_version": 2,
			"name": "Test",
			"task": [{
				"name": "Task",
				"entry": "Entry",
				"option": ["NonExistent"]
			}]
		}`
		_, err := ParseV2([]byte(data))
		require.Error(t, err)
	})

	t.Run("option select without cases", func(t *testing.T) {
		data := `{
			"interface_version": 2,
			"name": "Test",
			"option": {
				"MyOption": {
					"type": "select"
				}
			}
		}`
		_, err := ParseV2([]byte(data))
		require.Error(t, err)
	})

	t.Run("option switch must have 2 cases", func(t *testing.T) {
		data := `{
			"interface_version": 2,
			"name": "Test",
			"option": {
				"MyOption": {
					"type": "switch",
					"cases": [{"name": "Yes"}]
				}
			}
		}`
		_, err := ParseV2([]byte(data))
		require.Error(t, err)
	})

	t.Run("option input without inputs", func(t *testing.T) {
		data := `{
			"interface_version": 2,
			"name": "Test",
			"option": {
				"MyOption": {
					"type": "input"
				}
			}
		}`
		_, err := ParseV2([]byte(data))
		require.Error(t, err)
	})

	t.Run("option input with invalid regex", func(t *testing.T) {
		data := `{
			"interface_version": 2,
			"name": "Test",
			"option": {
				"MyOption": {
					"type": "input",
					"inputs": [{
						"name": "Field",
						"verify": "[invalid"
					}]
				}
			}
		}`
		_, err := ParseV2([]byte(data))
		require.Error(t, err)
	})
}

func TestV2OptionGetType(t *testing.T) {
	type Case struct {
		opt      V2Option
		expected string
	}

	testCases := []Case{
		{V2Option{Type: ""}, "select"},
		{V2Option{Type: "select"}, "select"},
		{V2Option{Type: "input"}, "input"},
		{V2Option{Type: "switch"}, "switch"},
	}

	for _, tc := range testCases {
		got := tc.opt.GetType()
		require.Equal(t, tc.expected, got)
	}
}

func TestV2I18nResolver(t *testing.T) {
	// Create temp translation file
	tmpDir := t.TempDir()
	transFile := filepath.Join(tmpDir, "i18n.json")
	transContent := `{
		"hello": "你好",
		"world": "世界"
	}`
	err := os.WriteFile(transFile, []byte(transContent), 0644)
	require.NoError(t, err)

	resolver, err := NewV2I18nResolver(transFile)
	require.NoError(t, err)

	type Case struct {
		input    string
		expected string
	}

	testCases := []Case{
		{"$hello", "你好"},
		{"$world", "世界"},
		{"$unknown", "$unknown"},
		{"plain text", "plain text"},
		{"", ""},
	}

	for _, tc := range testCases {
		got := resolver.Resolve(tc.input)
		require.Equal(t, tc.expected, got)
	}
}

func TestLoadV2FromFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create interface.json
	interfaceFile := filepath.Join(tmpDir, "interface.json")
	interfaceContent := `{
		"interface_version": 2,
		"name": "TestProject",
		"label": "$project_name",
		"languages": {
			"zh_cn": "i18n_zh.json"
		}
	}`
	err := os.WriteFile(interfaceFile, []byte(interfaceContent), 0644)
	require.NoError(t, err)

	// Create translation file
	transFile := filepath.Join(tmpDir, "i18n_zh.json")
	transContent := `{
		"project_name": "测试项目"
	}`
	err = os.WriteFile(transFile, []byte(transContent), 0644)
	require.NoError(t, err)

	loaded, err := LoadV2FromFile(interfaceFile)
	require.NoError(t, err)

	// Check interface
	require.Equal(t, "TestProject", loaded.Interface.Name)

	// Check languages
	langs := loaded.GetLanguages()
	require.Equal(t, 1, len(langs))
	require.Equal(t, "zh_cn", langs[0])

	// Check resolve
	resolved := loaded.ResolveString("$project_name", "zh_cn")
	require.Equal(t, "测试项目", resolved)

	// Check resolve with non-existent language
	resolved = loaded.ResolveString("$project_name", "en_us")
	require.Equal(t, "project_name", resolved)

	// Check resolve non-i18n string
	resolved = loaded.ResolveString("plain", "zh_cn")
	require.Equal(t, "plain", resolved)
}

func TestParseV2File(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "interface.json")

	content := `{
		"interface_version": 2,
		"name": "FileTest"
	}`
	err := os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err)

	iface, err := ParseV2File(filePath)
	require.NoError(t, err)

	require.Equal(t, "FileTest", iface.Name)
}

func TestDetectVersionFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "interface.json")

	content := `{"interface_version": 2}`
	err := os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err)

	version, err := DetectVersionFromFile(filePath)
	require.NoError(t, err)

	require.Equal(t, Version2, version)
}
