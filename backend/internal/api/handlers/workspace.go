package handlers

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/webide/ide/backend/internal/db"
)

type WorkspaceState struct {
	OpenFiles     []string `json:"open_files"`
	ExpandedDirs  []string `json:"expanded_dirs"`
	ActiveFile    string   `json:"active_file,omitempty"`
	ActiveTab     string   `json:"active_tab"`
	OpenTerminals []string `json:"open_terminals"`
}

type SaveWorkspaceRequest struct {
	OpenFiles     []string `json:"open_files"`
	ExpandedDirs  []string `json:"expanded_dirs"`
	ActiveFile    string   `json:"active_file,omitempty"`
	ActiveTab     string   `json:"active_tab"`
	OpenTerminals []string `json:"open_terminals"`
}

func GetWorkspace(c *fiber.Ctx) error {
	projectID := c.Params("id")
	userIDVal := c.Locals("user_id")
	if userIDVal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	var id, openFilesJSON, expandedDirsJSON, activeFile, activeTab, openTerminalsJSON string
	var updatedAt time.Time

	err := db.GetDB().QueryRow(`
		SELECT id, open_files_json, expanded_dirs_json, active_file, active_tab, open_terminals_json, updated_at
		FROM workspace_state
		WHERE user_id = ? AND project_id = ?`,
		userID.String(), projectID,
	).Scan(&id, &openFilesJSON, &expandedDirsJSON, &activeFile, &activeTab, &openTerminalsJSON, &updatedAt)

	if err == sql.ErrNoRows {
		return c.JSON(WorkspaceState{
			ActiveTab: "terminal",
		})
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get workspace"})
	}

	openFiles := []string{}
	expandedDirs := []string{}
	openTerminals := []string{}

	json.Unmarshal([]byte(openFilesJSON), &openFiles)
	json.Unmarshal([]byte(expandedDirsJSON), &expandedDirs)
	json.Unmarshal([]byte(openTerminalsJSON), &openTerminals)

	return c.JSON(WorkspaceState{
		OpenFiles:     openFiles,
		ExpandedDirs:  expandedDirs,
		ActiveFile:    activeFile,
		ActiveTab:     activeTab,
		OpenTerminals: openTerminals,
	})
}

func SaveWorkspace(c *fiber.Ctx) error {
	projectID := c.Params("id")
	userIDVal := c.Locals("user_id")
	if userIDVal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	var req SaveWorkspaceRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	openFilesJSON, _ := json.Marshal(req.OpenFiles)
	expandedDirsJSON, _ := json.Marshal(req.ExpandedDirs)
	openTerminalsJSON, _ := json.Marshal(req.OpenTerminals)

	if activeTab := req.ActiveTab; activeTab == "" {
		req.ActiveTab = "terminal"
	}

	_, err := db.GetDB().Exec(`
		INSERT INTO workspace_state (id, user_id, project_id, open_files_json, expanded_dirs_json, active_file, active_tab, open_terminals_json, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(user_id, project_id) DO UPDATE SET
			open_files_json = excluded.open_files_json,
			expanded_dirs_json = excluded.expanded_dirs_json,
			active_file = excluded.active_file,
			active_tab = excluded.active_tab,
			open_terminals_json = excluded.open_terminals_json,
			updated_at = CURRENT_TIMESTAMP`,
		uuid.New().String(), userID.String(), projectID,
		string(openFilesJSON), string(expandedDirsJSON), req.ActiveFile, req.ActiveTab, string(openTerminalsJSON),
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to save workspace"})
	}

	return c.JSON(fiber.Map{"status": "ok"})
}
