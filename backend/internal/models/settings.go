package models

import (
	"time"

	"github.com/google/uuid"
)

type UserSettings struct {
	ID              uuid.UUID `json:"id" db:"id"`
	UserID          uuid.UUID `json:"user_id" db:"user_id"`
	AIProvider      string    `json:"ai_provider" db:"ai_provider"`
	AIBaseURL       string    `json:"ai_base_url" db:"ai_base_url"`
	AIAPIKey        string    `json:"ai_api_key" db:"ai_api_key"`
	AIModel         string    `json:"ai_model" db:"ai_model"`
	UIThemeID       string    `json:"ui_theme_id" db:"ui_theme_id"`
	EditorThemeID   string    `json:"editor_theme_id" db:"editor_theme_id"`
	TerminalThemeID string    `json:"terminal_theme_id" db:"terminal_theme_id"`
	CustomThemeJSON string    `json:"custom_theme_json" db:"custom_theme_json"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

type Theme struct {
	ID     string            `json:"id"`
	Name   string            `json:"name"`
	Colors map[string]string `json:"colors"`
}

var BuiltinThemes = []Theme{
	{
		ID:   "dark-plus",
		Name: "Dark+",
		Colors: map[string]string{
			"background":             "#1e1e1e",
			"foreground":             "#d4d4d4",
			"muted":                  "#2d2d30",
			"muted-foreground":       "#858585",
			"border":                 "#454545",
			"card":                   "#252526",
			"card-foreground":        "#cccccc",
			"primary":                "#007acc",
			"primary-foreground":     "#ffffff",
			"secondary":              "#3c3c3c",
			"secondary-foreground":   "#cccccc",
			"accent":                 "#3c3c3c",
			"accent-foreground":      "#cccccc",
			"destructive":            "#d13438",
			"destructive-foreground": "#ffffff",
		},
	},
	{
		ID:   "light-plus",
		Name: "Light+",
		Colors: map[string]string{
			"background":             "#ffffff",
			"foreground":             "#333333",
			"muted":                  "#f3f3f3",
			"muted-foreground":       "#666666",
			"border":                 "#e0e0e0",
			"card":                   "#ffffff",
			"card-foreground":        "#333333",
			"primary":                "#007acc",
			"primary-foreground":     "#ffffff",
			"secondary":              "#f3f3f3",
			"secondary-foreground":   "#333333",
			"accent":                 "#f3f3f3",
			"accent-foreground":      "#333333",
			"destructive":            "#d13438",
			"destructive-foreground": "#ffffff",
		},
	},
	{
		ID:   "monokai",
		Name: "Monokai",
		Colors: map[string]string{
			"background":             "#272822",
			"foreground":             "#f8f8f2",
			"muted":                  "#3e3d32",
			"muted-foreground":       "#a6a6a6",
			"border":                 "#49483e",
			"card":                   "#2d2d2a",
			"card-foreground":        "#f8f8f2",
			"primary":                "#a6e22e",
			"primary-foreground":     "#272822",
			"secondary":              "#3e3d32",
			"secondary-foreground":   "#f8f8f2",
			"accent":                 "#3e3d32",
			"accent-foreground":      "#f8f8f2",
			"destructive":            "#f92672",
			"destructive-foreground": "#f8f8f2",
		},
	},
	{
		ID:   "nord",
		Name: "Nord",
		Colors: map[string]string{
			"background":             "#2e3440",
			"foreground":             "#d8dee9",
			"muted":                  "#3b4252",
			"muted-foreground":       "#81a1c1",
			"border":                 "#4c566a",
			"card":                   "#3b4252",
			"card-foreground":        "#d8dee9",
			"primary":                "#88c0d0",
			"primary-foreground":     "#2e3440",
			"secondary":              "#434c5e",
			"secondary-foreground":   "#d8dee9",
			"accent":                 "#434c5e",
			"accent-foreground":      "#d8dee9",
			"destructive":            "#bf616a",
			"destructive-foreground": "#2e3440",
		},
	},
	{
		ID:   "dracula",
		Name: "Dracula",
		Colors: map[string]string{
			"background":             "#282a36",
			"foreground":             "#f8f8f2",
			"muted":                  "#44475a",
			"muted-foreground":       "#6272a4",
			"border":                 "#44475a",
			"card":                   "#21222c",
			"card-foreground":        "#f8f8f2",
			"primary":                "#bd93f9",
			"primary-foreground":     "#282a36",
			"secondary":              "#44475a",
			"secondary-foreground":   "#f8f8f2",
			"accent":                 "#44475a",
			"accent-foreground":      "#f8f8f2",
			"destructive":            "#ff5555",
			"destructive-foreground": "#f8f8f2",
		},
	},
}

func GetBuiltinThemeByID(id string) *Theme {
	for i := range BuiltinThemes {
		if BuiltinThemes[i].ID == id {
			return &BuiltinThemes[i]
		}
	}
	return &BuiltinThemes[0]
}
