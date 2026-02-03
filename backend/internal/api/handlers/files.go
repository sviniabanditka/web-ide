package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/webide/ide/backend/internal/files"
	"github.com/webide/ide/backend/internal/projects"
)

type WriteFileRequest struct {
	Content      string `json:"content"`
	ExpectedEtag string `json:"expectedEtag"`
}

type MkdirRequest struct {
	Path string `json:"path"`
}

type RenameRequest struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type RemoveRequest struct {
	Path      string `json:"path"`
	Recursive bool   `json:"recursive"`
}

func GetFileTree(c *fiber.Ctx) error {
	projectID := c.Params("id")
	id, err := uuid.Parse(projectID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}

	project, err := projects.GetProject(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "project not found"})
	}

	path := c.Query("path", "/")

	tree, err := files.BuildFileTree(project.RootPath, path, 0)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to build file tree"})
	}

	return c.JSON(tree)
}

func GetFile(c *fiber.Ctx) error {
	projectID := c.Params("id")
	id, err := uuid.Parse(projectID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}

	project, err := projects.GetProject(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "project not found"})
	}

	path := c.Query("path")
	if path == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "path required"})
	}

	normalizedPath, err := projects.NormalizePath(project.RootPath, path)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "invalid path"})
	}

	fileInfo, err := files.ReadFile(project.RootPath, normalizedPath)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "file not found"})
	}

	return c.JSON(fileInfo)
}

func PutFile(c *fiber.Ctx) error {
	projectID := c.Params("id")
	id, err := uuid.Parse(projectID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}

	project, err := projects.GetProject(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "project not found"})
	}

	path := c.Query("path")
	if path == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "path required"})
	}

	normalizedPath, err := projects.NormalizePath(project.RootPath, path)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "invalid path"})
	}

	var req WriteFileRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	if err := files.WriteFile(project.RootPath, normalizedPath, req.Content, req.ExpectedEtag); err != nil {
		if err.Error() == "etag mismatch" {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "etag mismatch"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "saved"})
}

func Mkdir(c *fiber.Ctx) error {
	projectID := c.Params("id")
	id, err := uuid.Parse(projectID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}

	project, err := projects.GetProject(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "project not found"})
	}

	var req MkdirRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	normalizedPath, err := projects.NormalizePath(project.RootPath, req.Path)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "invalid path"})
	}

	if err := files.Mkdir(project.RootPath, normalizedPath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "created"})
}

func RenameFile(c *fiber.Ctx) error {
	projectID := c.Params("id")
	id, err := uuid.Parse(projectID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}

	project, err := projects.GetProject(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "project not found"})
	}

	var req RenameRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	fromPath, err := projects.NormalizePath(project.RootPath, req.From)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "invalid from path"})
	}

	toPath, err := projects.NormalizePath(project.RootPath, req.To)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "invalid to path"})
	}

	if err := files.Rename(project.RootPath, fromPath, toPath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "renamed"})
}

func RemoveFile(c *fiber.Ctx) error {
	projectID := c.Params("id")
	id, err := uuid.Parse(projectID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project id"})
	}

	project, err := projects.GetProject(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "project not found"})
	}

	var req RemoveRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	normalizedPath, err := projects.NormalizePath(project.RootPath, req.Path)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "invalid path"})
	}

	if err := files.Remove(project.RootPath, normalizedPath, req.Recursive); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "deleted"})
}
