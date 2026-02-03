package files

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/webide/ide/backend/internal/projects"
)

type FileNode struct {
	ID       string      `json:"id"`
	Name     string      `json:"name"`
	Path     string      `json:"path"`
	Type     string      `json:"type"` // "file" | "directory"
	Children []*FileNode `json:"children,omitempty"`
}

type FileInfo struct {
	Path      string `json:"path"`
	Content   string `json:"content,omitempty"`
	Etag      string `json:"etag"`
	Size      int64  `json:"size"`
	UpdatedAt string `json:"updated_at"`
}

const maxFileSize = 5 * 1024 * 1024 // 5MB
const maxTreeDepth = 20

func BuildFileTree(projectRoot, relativePath string, depth int) (*FileNode, error) {
	if depth > maxTreeDepth {
		return nil, fmt.Errorf("max depth exceeded")
	}

	fullPath := projects.PathToFull(projectRoot, relativePath)

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}

	cleanPath := "/" + strings.TrimLeft(relativePath, "/")
	node := &FileNode{
		ID:   uuid.New().String(),
		Name: filepath.Base(cleanPath),
		Path: cleanPath,
		Type: "directory",
	}

	var children []*FileNode
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}

		childPath := filepath.Join(relativePath, name)
		cleanChildPath := "/" + strings.TrimLeft(childPath, "/")
		if entry.IsDir() {
			child, err := BuildFileTree(projectRoot, childPath, depth+1)
			if err != nil {
				continue
			}
			child.Name = name
			child.Path = cleanChildPath
			child.Type = "directory"
			children = append(children, child)
		} else {
			children = append(children, &FileNode{
				ID:   uuid.New().String(),
				Name: name,
				Path: cleanChildPath,
				Type: "file",
			})
		}
	}

	if len(children) > 0 {
		node.Type = "directory"
		node.Children = children
	}

	return node, nil
}

func ReadFile(projectRoot, relativePath string) (*FileInfo, error) {
	fullPath := projects.PathToFull(projectRoot, relativePath)

	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, err
	}

	if info.IsDir() {
		return nil, fmt.Errorf("path is a directory")
	}

	if info.Size() > maxFileSize {
		return nil, fmt.Errorf("file too large")
	}

	file, err := os.Open(fullPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	etag := calculateEtag(content, info.ModTime())

	return &FileInfo{
		Path:      relativePath,
		Content:   string(content),
		Etag:      etag,
		Size:      info.Size(),
		UpdatedAt: info.ModTime().Format(time.RFC3339),
	}, nil
}

func WriteFile(projectRoot, relativePath, content, expectedEtag string) error {
	fullPath := projects.PathToFull(projectRoot, relativePath)

	info, err := os.Stat(fullPath)
	if err == nil && info.IsDir() {
		return fmt.Errorf("path is a directory")
	}

	if expectedEtag != "" && err == nil {
		existingContent, err := os.ReadFile(fullPath)
		if err != nil {
			return err
		}
		existingEtag := calculateEtag(existingContent, info.ModTime())
		if existingEtag != expectedEtag {
			return fmt.Errorf("etag mismatch")
		}
	}

	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return err
	}

	return nil
}

func Mkdir(projectRoot, relativePath string) error {
	fullPath := projects.PathToFull(projectRoot, relativePath)

	if strings.Contains(relativePath, "..") {
		return fmt.Errorf("invalid path")
	}

	return os.MkdirAll(fullPath, 0755)
}

func Rename(projectRoot, fromPath, toPath string) error {
	fromFull := projects.PathToFull(projectRoot, fromPath)
	toFull := projects.PathToFull(projectRoot, toPath)

	if strings.HasPrefix(toFull, fromFull) && toFull != fromFull {
		return fmt.Errorf("cannot rename to subdirectory of source")
	}

	return os.Rename(fromFull, toFull)
}

func Remove(projectRoot, relativePath string, recursive bool) error {
	fullPath := projects.PathToFull(projectRoot, relativePath)

	if recursive {
		return os.RemoveAll(fullPath)
	}
	return os.Remove(fullPath)
}

func calculateEtag(content []byte, modTime time.Time) string {
	hash := sha256.Sum256(content)
	return fmt.Sprintf("\"%s-%d\"", hex.EncodeToString(hash[:8]), modTime.UnixNano())
}
