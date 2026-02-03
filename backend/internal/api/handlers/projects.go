package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/webide/ide/backend/internal/projects"
)

type CreateProjectRequest struct {
	Name     string `json:"name"`
	RootPath string `json:"rootPath"`
}

type ProjectResponse struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	RootPath     string `json:"root_path"`
	IsGitRepo    bool   `json:"is_git_repo,omitempty"`
	CreatedAt    string `json:"created_at"`
	LastOpenedAt string `json:"last_opened_at"`
}

func ListProjects(c *fiber.Ctx, projectsDir string) error {
	dbProjects, err := projects.GetProjects()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get projects"})
	}

	if len(dbProjects) == 0 {
		scanned, err := projects.ScanProjects(projectsDir)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to scan projects"})
		}

		var resp []ProjectResponse
		for _, p := range scanned {
			registered, err := projects.RegisterProject(p.Name, p.RootPath)
			if err != nil {
				continue
			}
			resp = append(resp, ProjectResponse{
				ID:        registered.ID.String(),
				Name:      registered.Name,
				RootPath:  registered.RootPath,
				IsGitRepo: p.IsGitRepo,
				CreatedAt: registered.CreatedAt.Format("2006-01-02T15:04:05Z"),
			})
		}
		return c.JSON(resp)
	}

	var resp []ProjectResponse
	for _, p := range dbProjects {
		resp = append(resp, ProjectResponse{
			ID:           p.ID.String(),
			Name:         p.Name,
			RootPath:     p.RootPath,
			CreatedAt:    p.CreatedAt.Format("2006-01-02T15:04:05Z"),
			LastOpenedAt: p.LastOpenedAt.Format("2006-01-02T15:04:05Z"),
		})
	}
	return c.JSON(resp)
}

func GetProject(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}

	project, err := projects.GetProject(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "project not found"})
	}

	projects.UpdateLastOpened(id)

	return c.JSON(ProjectResponse{
		ID:           project.ID.String(),
		Name:         project.Name,
		RootPath:     project.RootPath,
		CreatedAt:    project.CreatedAt.Format("2006-01-02T15:04:05Z"),
		LastOpenedAt: project.LastOpenedAt.Format("2006-01-02T15:04:05Z"),
	})
}

func CreateProject(c *fiber.Ctx) error {
	var req CreateProjectRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	if req.Name == "" || req.RootPath == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name and rootPath required"})
	}

	project, err := projects.RegisterProject(req.Name, req.RootPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create project"})
	}

	return c.JSON(ProjectResponse{
		ID:        project.ID.String(),
		Name:      project.Name,
		RootPath:  project.RootPath,
		CreatedAt: project.CreatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

func ScanProjects(c *fiber.Ctx, projectsDir string) error {
	scanned, err := projects.ScanProjects(projectsDir)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to scan projects"})
	}

	var registered []ProjectResponse
	for _, p := range scanned {
		project, err := projects.RegisterProject(p.Name, p.RootPath)
		if err != nil {
			continue
		}
		registered = append(registered, ProjectResponse{
			ID:        project.ID.String(),
			Name:      project.Name,
			RootPath:  project.RootPath,
			IsGitRepo: p.IsGitRepo,
			CreatedAt: project.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return c.JSON(registered)
}
