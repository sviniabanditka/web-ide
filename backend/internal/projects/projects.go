package projects

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/webide/ide/backend/internal/db"
	"github.com/webide/ide/backend/internal/models"
)

type ProjectInfo struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	RootPath     string    `json:"root_path"`
	IsGitRepo    bool      `json:"is_git_repo"`
	CreatedAt    string    `json:"created_at"`
	LastOpenedAt string    `json:"last_opened_at"`
}

func ScanProjects(projectsDir string) ([]ProjectInfo, error) {
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return nil, err
	}

	var projects []ProjectInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		rootPath := filepath.Join(projectsDir, name)

		isGitRepo := false
		if _, err := os.Stat(filepath.Join(rootPath, ".git")); err == nil {
			isGitRepo = true
		}

		projects = append(projects, ProjectInfo{
			Name:      name,
			RootPath:  rootPath,
			IsGitRepo: isGitRepo,
		})
	}

	return projects, nil
}

func RegisterProject(name, rootPath string) (*models.Project, error) {
	id := uuid.New()
	_, err := db.GetDB().Exec(`
		INSERT INTO projects (id, name, root_path, created_at, last_opened_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		id, name, rootPath)
	if err != nil {
		return nil, err
	}

	return &models.Project{
		ID:       id,
		Name:     name,
		RootPath: rootPath,
	}, nil
}

func GetProject(id uuid.UUID) (*models.Project, error) {
	var project models.Project
	err := db.GetDB().QueryRow(`
		SELECT id, name, root_path, created_at, last_opened_at FROM projects WHERE id = ?`,
		id).Scan(&project.ID, &project.Name, &project.RootPath, &project.CreatedAt, &project.LastOpenedAt)
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func GetProjects() ([]models.Project, error) {
	rows, err := db.GetDB().Query(`
		SELECT id, name, root_path, created_at, last_opened_at FROM projects ORDER BY last_opened_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		var p models.Project
		if err := rows.Scan(&p.ID, &p.Name, &p.RootPath, &p.CreatedAt, &p.LastOpenedAt); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, nil
}

func UpdateLastOpened(id uuid.UUID) error {
	_, err := db.GetDB().Exec("UPDATE projects SET last_opened_at = CURRENT_TIMESTAMP WHERE id = ?", id)
	return err
}

func NormalizePath(projectRoot, relativePath string) (string, error) {
	cleaned := filepath.Clean(filepath.Join("/", relativePath))

	if strings.HasPrefix(cleaned, "..") {
		return "", os.ErrPermission
	}

	fullPath := filepath.Join(projectRoot, cleaned)

	if !strings.HasPrefix(fullPath, projectRoot) {
		return "", os.ErrPermission
	}

	return cleaned, nil
}

func PathToFull(projectRoot, relativePath string) string {
	return filepath.Join(projectRoot, relativePath)
}
