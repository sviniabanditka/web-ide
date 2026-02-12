package handlers

import (
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/webide/ide/backend/internal/db"
	"github.com/webide/ide/backend/internal/models"
)

func GetSettings(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var settings models.UserSettings
	err := db.GetDB().QueryRow(`
		SELECT id, user_id, ai_provider, ai_base_url, ai_api_key, ai_model,
		       ui_theme_id, editor_theme_id, terminal_theme_id, custom_theme_json, created_at, updated_at
		FROM user_settings WHERE user_id = ?`, userID).Scan(
		&settings.ID, &settings.UserID, &settings.AIProvider, &settings.AIBaseURL,
		&settings.AIAPIKey, &settings.AIModel, &settings.UIThemeID, &settings.EditorThemeID,
		&settings.TerminalThemeID, &settings.CustomThemeJSON, &settings.CreatedAt, &settings.UpdatedAt)

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			defaultSettings := models.UserSettings{
				ID:              uuid.New(),
				UserID:          userID,
				AIProvider:      "anthropic",
				AIBaseURL:       "",
				AIAPIKey:        "",
				AIModel:         "claude-sonnet-4-20250514",
				UIThemeID:       "dark-plus",
				EditorThemeID:   "vs-dark",
				TerminalThemeID: "monokai",
				CustomThemeJSON: "{}",
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			}
			return c.JSON(defaultSettings)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get settings"})
	}

	return c.JSON(settings)
}

func SaveSettings(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var input struct {
		AIProvider      string `json:"ai_provider"`
		AIBaseURL       string `json:"ai_base_url"`
		AIAPIKey        string `json:"ai_api_key"`
		AIModel         string `json:"ai_model"`
		UIThemeID       string `json:"ui_theme_id"`
		EditorThemeID   string `json:"editor_theme_id"`
		TerminalThemeID string `json:"terminal_theme_id"`
		CustomThemeJSON string `json:"custom_theme_json"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	var settingsID uuid.UUID
	var existingSettings struct {
		ID              uuid.UUID
		AIProvider      string
		AIBaseURL       string
		AIAPIKey        string
		AIModel         string
		UIThemeID       string
		EditorThemeID   string
		TerminalThemeID string
		CustomThemeJSON string
	}

	err := db.GetDB().QueryRow(`
		SELECT id, ai_provider, ai_base_url, ai_api_key, ai_model, ui_theme_id, editor_theme_id, terminal_theme_id, custom_theme_json
		FROM user_settings WHERE user_id = ?`, userID).Scan(
		&existingSettings.ID, &existingSettings.AIProvider, &existingSettings.AIBaseURL,
		&existingSettings.AIAPIKey, &existingSettings.AIModel, &existingSettings.UIThemeID,
		&existingSettings.EditorThemeID, &existingSettings.TerminalThemeID, &existingSettings.CustomThemeJSON)

	if err != nil && err.Error() != "sql: no rows in result set" {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to check existing settings"})
	}

	if existingSettings.ID != uuid.Nil {
		settingsID = existingSettings.ID

		aiProvider := input.AIProvider
		if aiProvider == "" {
			aiProvider = existingSettings.AIProvider
		}
		aiBaseURL := input.AIBaseURL
		if aiBaseURL == "" {
			aiBaseURL = existingSettings.AIBaseURL
		}
		aiAPIKey := input.AIAPIKey
		if aiAPIKey == "" {
			aiAPIKey = existingSettings.AIAPIKey
		}
		aiModel := input.AIModel
		if aiModel == "" {
			aiModel = existingSettings.AIModel
		}
		uiThemeID := input.UIThemeID
		if uiThemeID == "" {
			uiThemeID = existingSettings.UIThemeID
		}
		editorThemeID := input.EditorThemeID
		if editorThemeID == "" {
			editorThemeID = existingSettings.EditorThemeID
		}
		terminalThemeID := input.TerminalThemeID
		if terminalThemeID == "" {
			terminalThemeID = existingSettings.TerminalThemeID
		}
		customThemeJSON := input.CustomThemeJSON
		if customThemeJSON == "" {
			customThemeJSON = existingSettings.CustomThemeJSON
		}

		_, err = db.GetDB().Exec(`
			UPDATE user_settings
			SET ai_provider = ?, ai_base_url = ?, ai_api_key = ?, ai_model = ?,
			    ui_theme_id = ?, editor_theme_id = ?, terminal_theme_id = ?, custom_theme_json = ?, updated_at = CURRENT_TIMESTAMP
			WHERE id = ?`,
			aiProvider, aiBaseURL, aiAPIKey, aiModel,
			uiThemeID, editorThemeID, terminalThemeID, customThemeJSON, settingsID)
	} else {
		settingsID = uuid.New()
		aiProvider := input.AIProvider
		if aiProvider == "" {
			aiProvider = "anthropic"
		}
		aiBaseURL := input.AIBaseURL
		if aiBaseURL == "" {
			aiBaseURL = ""
		}
		aiAPIKey := input.AIAPIKey
		if aiAPIKey == "" {
			aiAPIKey = ""
		}
		aiModel := input.AIModel
		if aiModel == "" {
			aiModel = "claude-sonnet-4-20250514"
		}
		uiThemeID := input.UIThemeID
		if uiThemeID == "" {
			uiThemeID = "dark-plus"
		}
		editorThemeID := input.EditorThemeID
		if editorThemeID == "" {
			editorThemeID = "vs-dark"
		}
		terminalThemeID := input.TerminalThemeID
		if terminalThemeID == "" {
			terminalThemeID = "monokai"
		}
		customThemeJSON := input.CustomThemeJSON
		if customThemeJSON == "" {
			customThemeJSON = "{}"
		}

		_, err = db.GetDB().Exec(`
			INSERT INTO user_settings (id, user_id, ai_provider, ai_base_url, ai_api_key, ai_model, ui_theme_id, editor_theme_id, terminal_theme_id, custom_theme_json)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			settingsID, userID, aiProvider, aiBaseURL, aiAPIKey,
			aiModel, uiThemeID, editorThemeID, terminalThemeID, customThemeJSON)
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save settings"})
	}

	return c.JSON(fiber.Map{"status": "saved"})
}

func GetThemes(c *fiber.Ctx) error {
	themeType := c.Query("type", "ui")
	themes := make([]map[string]interface{}, len(models.BuiltinThemes))
	for i, theme := range models.BuiltinThemes {
		var colors map[string]string
		json.Unmarshal([]byte("{}"), &colors)
		themes[i] = map[string]interface{}{
			"id":     theme.ID,
			"name":   theme.Name,
			"colors": theme.Colors,
		}
	}
	if themeType == "editor" {
		monacoThemes := []map[string]interface{}{
			{"id": "vs-dark", "name": "Dark (VS Code)"},
			{"id": "vs", "name": "Light (VS Code)"},
			{"id": "hc-black", "name": "High Contrast"},
		}
		return c.JSON(monacoThemes)
	}
	return c.JSON(themes)
}
