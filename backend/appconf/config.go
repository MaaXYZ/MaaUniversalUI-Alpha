package appconf

type AppConfig struct {
	Theme    Theme    `json:"theme"`
	Language Language `json:"language"`
}

type Theme string

const (
	ThemeLight  Theme = "light"
	ThemeDark   Theme = "dark"
	ThemeSystem Theme = "system"
)

// GetSupportedThemes gets supported themes
func GetSupportedThemes() []string {
	return []string{
		string(ThemeLight),
		string(ThemeDark),
		string(ThemeSystem),
	}
}

func isValidTheme(t Theme) bool {
	switch t {
	case ThemeLight, ThemeDark, ThemeSystem:
		return true
	default:
		return false
	}
}

type Language string

const (
	LangZhCN Language = "zh-CN" // Simplified Chinese
	LangZhTW Language = "zh-TW" // Traditional Chinese
	LangEnUS Language = "en-US" // English
)

type SupportedLanguage struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// GetSupportedLanguages gets supported languages
func GetSupportedLanguages() []SupportedLanguage {
	return []SupportedLanguage{
		{Code: string(LangZhCN), Name: "简体中文"},
		{Code: string(LangZhTW), Name: "繁體中文"},
		{Code: string(LangEnUS), Name: "English"},
	}
}

type Supported struct {
	Themes    []string            `json:"themes"`
	Languages []SupportedLanguage `json:"languages"`
}

func GetSupported() Supported {
	return Supported{
		Themes:    GetSupportedThemes(),
		Languages: GetSupportedLanguages(),
	}
}

func isValidLanguage(l Language) bool {
	switch l {
	case LangZhCN, LangZhTW, LangEnUS:
		return true
	default:
		return false
	}
}

func DefaultAppConfig() *AppConfig {
	return &AppConfig{
		Theme:    ThemeSystem,
		Language: LangZhCN,
	}
}
