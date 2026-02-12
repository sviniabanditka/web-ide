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
	if themeType == "terminal" {
		terminalThemes := []map[string]interface{}{
			{"id": "monokai", "name": "Monokai"},
			{"id": "nord", "name": "Nord"},
			{"id": "dracula", "name": "Dracula"},
			{"id": "github-dark", "name": "GitHub Dark"},
			{"id": "github-light", "name": "GitHub Light"},
			{"id": "one-dark", "name": "One Dark"},
		}
		return c.JSON(terminalThemes)
	}
	return c.JSON(themes)
}

func GetCustomThemes(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	themeType := c.Query("type", "")

	var query string
	var args []interface{}

	if themeType != "" {
		query = "SELECT id, type, name, colors_json, created_at, updated_at FROM custom_themes WHERE user_id = ? AND type = ?"
		args = []interface{}{userID, themeType}
	} else {
		query = "SELECT id, type, name, colors_json, created_at, updated_at FROM custom_themes WHERE user_id = ?"
		args = []interface{}{userID}
	}

	rows, err := db.GetDB().Query(query, args...)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get custom themes"})
	}
	defer rows.Close()

	var themes []map[string]interface{}
	for rows.Next() {
		var id uuid.UUID
		var themeType string
		var name string
		var colorsJSON string
		var createdAt, updatedAt time.Time

		if err := rows.Scan(&id, &themeType, &name, &colorsJSON, &createdAt, &updatedAt); err != nil {
			continue
		}

		var colors map[string]interface{}
		json.Unmarshal([]byte(colorsJSON), &colors)

		themes = append(themes, map[string]interface{}{
			"id":         id.String(),
			"type":       themeType,
			"name":       name,
			"colors":     colors,
			"created_at": createdAt,
			"updated_at": updatedAt,
		})
	}

	return c.JSON(themes)
}

func CreateCustomTheme(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var input struct {
		Type   string            `json:"type"`
		Name   string            `json:"name"`
		Colors map[string]string `json:"colors"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if input.Type == "" || input.Name == "" || len(input.Colors) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Type, name and colors are required"})
	}

	if input.Type != "ui" && input.Type != "editor" && input.Type != "terminal" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid theme type"})
	}

	colorsJSON, err := json.Marshal(input.Colors)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to marshal colors"})
	}

	themeID := uuid.New()
	_, err = db.GetDB().Exec(`
		INSERT INTO custom_themes (id, user_id, type, name, colors_json)
		VALUES (?, ?, ?, ?, ?)`,
		themeID, userID, input.Type, input.Name, string(colorsJSON))

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create custom theme"})
	}

	return c.JSON(map[string]interface{}{
		"id":         themeID.String(),
		"type":       input.Type,
		"name":       input.Name,
		"colors":     input.Colors,
		"created_at": time.Now(),
		"updated_at": time.Now(),
	})
}

func UpdateCustomTheme(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	themeIDStr := c.Params("id")
	themeID, err := uuid.Parse(themeIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid theme ID"})
	}

	var input struct {
		Name   string            `json:"name"`
		Colors map[string]string `json:"colors"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	var existingTheme struct {
		ID     uuid.UUID
		UserID uuid.UUID
	}

	err = db.GetDB().QueryRow("SELECT id, user_id FROM custom_themes WHERE id = ?", themeID).Scan(&existingTheme.ID, &existingTheme.UserID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Theme not found"})
	}

	if existingTheme.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Not authorized to update this theme"})
	}

	if input.Name == "" && len(input.Colors) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Name or colors are required"})
	}

	var setClauses []string
	var args []interface{}

	if input.Name != "" {
		setClauses = append(setClauses, "name = ?")
		args = append(args, input.Name)
	}

	if len(input.Colors) > 0 {
		colorsJSON, err := json.Marshal(input.Colors)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to marshal colors"})
		}
		setClauses = append(setClauses, "colors_json = ?")
		args = append(args, string(colorsJSON))
	}

	setClauses = append(setClauses, "updated_at = CURRENT_TIMESTAMP")
	args = append(args, themeID)

	query := "UPDATE custom_themes SET " + joinStrings(setClauses, ", ") + " WHERE id = ?"
	_, err = db.GetDB().Exec(query, args...)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update theme"})
	}

	return c.JSON(map[string]interface{}{"status": "saved"})
}

func DeleteCustomTheme(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	themeIDStr := c.Params("id")
	themeID, err := uuid.Parse(themeIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid theme ID"})
	}

	var existingTheme struct {
		ID     uuid.UUID
		UserID uuid.UUID
	}

	err = db.GetDB().QueryRow("SELECT id, user_id FROM custom_themes WHERE id = ?", themeID).Scan(&existingTheme.ID, &existingTheme.UserID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Theme not found"})
	}

	if existingTheme.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Not authorized to delete this theme"})
	}

	_, err = db.GetDB().Exec("DELETE FROM custom_themes WHERE id = ?", themeID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete theme"})
	}

	return c.JSON(map[string]interface{}{"status": "deleted"})
}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
