package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/webide/ide/backend/internal/projects"
	"github.com/webide/ide/backend/internal/terminal"
)

type CreateTerminalRequest struct {
	Title string `json:"title,omitempty"`
	Cwd   string `json:"cwd,omitempty"`
	Shell string `json:"shell,omitempty"`
}

type TerminalResponse struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	Title     string `json:"title"`
	Cwd       string `json:"cwd"`
	Shell     string `json:"shell"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

type ResizeRequest struct {
	Cols int `json:"cols"`
	Rows int `json:"rows"`
}

func CreateTerminal(c *fiber.Ctx) error {
	projectIDStr := c.Params("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}

	project, err := projects.GetProject(projectID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "project not found"})
	}

	var req CreateTerminalRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	cwd := req.Cwd
	if cwd == "" {
		cwd = project.RootPath
	}

	session, err := terminal.CreateSession(projectID, cwd, req.Title, req.Shell)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create terminal"})
	}

	return c.JSON(TerminalResponse{
		ID:        session.ID.String(),
		ProjectID: session.ProjectID.String(),
		Title:     session.Title,
		Cwd:       session.Cwd,
		Shell:     session.Shell,
		Status:    session.Status,
		CreatedAt: session.CreatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

func ListTerminals(c *fiber.Ctx) error {
	projectIDStr := c.Params("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}

	sessions, err := terminal.GetProjectSessions(projectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to list terminals"})
	}

	var result []TerminalResponse
	for _, s := range sessions {
		result = append(result, TerminalResponse{
			ID:        s.ID.String(),
			ProjectID: s.ProjectID.String(),
			Title:     s.Title,
			Cwd:       s.Cwd,
			Shell:     s.Shell,
			Status:    s.Status,
			CreatedAt: s.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return c.JSON(result)
}

func GetTerminal(c *fiber.Ctx) error {
	termIDStr := c.Params("tid")
	termID, err := uuid.Parse(termIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid terminal id"})
	}

	session, err := terminal.GetSession(termID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "terminal not found"})
	}

	return c.JSON(TerminalResponse{
		ID:        session.ID.String(),
		ProjectID: session.ProjectID.String(),
		Title:     session.Title,
		Cwd:       session.Cwd,
		Shell:     session.Shell,
		Status:    session.Status,
		CreatedAt: session.CreatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

func ResizeTerminal(c *fiber.Ctx) error {
	termIDStr := c.Params("tid")
	termID, err := uuid.Parse(termIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid terminal id"})
	}

	var req ResizeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	session, err := terminal.GetSession(termID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "terminal not found"})
	}

	if err := session.Resize(req.Cols, req.Rows); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to resize"})
	}

	return c.JSON(fiber.Map{"message": "resized"})
}

func CloseTerminal(c *fiber.Ctx) error {
	termIDStr := c.Params("tid")
	termID, err := uuid.Parse(termIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid terminal id"})
	}

	session, err := terminal.GetSession(termID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "terminal not found"})
	}

	session.Close()
	return c.JSON(fiber.Map{"message": "closed"})
}
