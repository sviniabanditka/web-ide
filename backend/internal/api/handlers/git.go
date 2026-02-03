package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/webide/ide/backend/internal/git"
	"github.com/webide/ide/backend/internal/projects"
)

type StageRequest struct {
	Paths []string `json:"paths"`
}

type UnstageRequest struct {
	Paths []string `json:"paths"`
}

type CommitRequest struct {
	Message string `json:"message"`
}

type PushRequest struct {
	Remote string `json:"remote"`
	Branch string `json:"branch"`
}

type GitStatusResponse struct {
	IsGitRepo bool           `json:"is_git_repo"`
	Status    *git.GitStatus `json:"status"`
	Branches  []string       `json:"branches,omitempty"`
	Current   string         `json:"current_branch,omitempty"`
}

func GetGitStatus(c *fiber.Ctx) error {
	projectID := c.Params("id")
	id, err := uuid.Parse(projectID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}

	project, err := projects.GetProject(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "project not found"})
	}

	isGitRepo := git.IsGitRepo(project.RootPath)

	if !isGitRepo {
		return c.JSON(GitStatusResponse{
			IsGitRepo: false,
		})
	}

	status, err := git.GetStatus(project.RootPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get git status"})
	}

	branches, _ := git.GetBranches(project.RootPath)

	return c.JSON(GitStatusResponse{
		IsGitRepo: true,
		Status:    status,
		Branches:  branches,
	})
}

func GetGitDiff(c *fiber.Ctx) error {
	projectID := c.Params("id")
	id, err := uuid.Parse(projectID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}

	project, err := projects.GetProject(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "project not found"})
	}

	cached := c.QueryInt("cached", 0) == 1
	filePath := c.Query("path", "")

	diff, err := git.GetDiff(project.RootPath, cached, filePath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get diff"})
	}

	return c.SendString(diff)
}

func StageFiles(c *fiber.Ctx) error {
	projectID := c.Params("id")
	id, err := uuid.Parse(projectID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}

	project, err := projects.GetProject(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "project not found"})
	}

	var req StageRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	if len(req.Paths) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "paths required"})
	}

	if err := git.StageFiles(project.RootPath, req.Paths); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to stage files"})
	}

	return c.JSON(fiber.Map{"message": "staged"})
}

func UnstageFiles(c *fiber.Ctx) error {
	projectID := c.Params("id")
	id, err := uuid.Parse(projectID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}

	project, err := projects.GetProject(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "project not found"})
	}

	var req UnstageRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	if len(req.Paths) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "paths required"})
	}

	if err := git.UnstageFiles(project.RootPath, req.Paths); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to unstage files"})
	}

	return c.JSON(fiber.Map{"message": "unstaged"})
}

func GitCommit(c *fiber.Ctx) error {
	projectID := c.Params("id")
	id, err := uuid.Parse(projectID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}

	project, err := projects.GetProject(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "project not found"})
	}

	var req CommitRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	if req.Message == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "message required"})
	}

	if err := git.Commit(project.RootPath, req.Message); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "committed"})
}

func GitPush(c *fiber.Ctx) error {
	projectID := c.Params("id")
	id, err := uuid.Parse(projectID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}

	project, err := projects.GetProject(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "project not found"})
	}

	var req PushRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	if err := git.Push(project.RootPath, req.Remote, req.Branch); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "pushed"})
}

func GetGitBranches(c *fiber.Ctx) error {
	projectID := c.Params("id")
	id, err := uuid.Parse(projectID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}

	project, err := projects.GetProject(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "project not found"})
	}

	branches, err := git.GetBranches(project.RootPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get branches"})
	}

	return c.JSON(fiber.Map{"branches": branches})
}

func GetGitLog(c *fiber.Ctx) error {
	projectID := c.Params("id")
	id, err := uuid.Parse(projectID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}

	project, err := projects.GetProject(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "project not found"})
	}

	limit := c.QueryInt("limit", 50)

	entries, err := git.GetLog(project.RootPath, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get log"})
	}

	return c.JSON(fiber.Map{"log": entries})
}
