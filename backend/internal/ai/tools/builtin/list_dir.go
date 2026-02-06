package builtin

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/webide/ide/backend/internal/ai/tools"
)

func ListDir() tools.Tool {
	return tools.Tool{
		Name:        "list_dir",
		Description: "List directory contents with optional depth and hidden files",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":    "string",
					"default": ".",
				},
				"depth": map[string]interface{}{
					"type":    "integer",
					"default": 1,
					"minimum": 1,
					"maximum": 5,
				},
				"include_hidden": map[string]interface{}{
					"type":    "boolean",
					"default": false,
				},
			},
			"required": []string{"path"},
		},
		Policy: tools.PolicyAllow,
		Execute: func(ctx context.Context, args map[string]interface{}, tc tools.ToolContext) (tools.ToolResult, error) {
			startTime := time.Now()

			path, _ := args["path"].(string)
			if path == "" {
				path = "."
			}

			depth := 1
			if d, ok := args["depth"].(float64); ok {
				depth = int(d)
			}
			if depth < 1 {
				depth = 1
			}
			if depth > 5 {
				depth = 5
			}

			includeHidden := false
			if h, ok := args["include_hidden"].(bool); ok {
				includeHidden = h
			}

			path, err := normalizePath(path, tc.ProjectRoot)
			if err != nil {
				return tools.NewErrorResult(tools.ErrCodeInvalidPath, err.Error(), nil), nil
			}

			entries, err := listDirRecursive(path, depth, includeHidden, tc.ProjectRoot)
			if err != nil {
				return tools.NewErrorResult(tools.ErrCodeNotFound, err.Error(), nil), nil
			}

			relPath, _ := filepath.Rel(tc.ProjectRoot, path)

			return tools.ToolResult{
				OK: true,
				Data: map[string]interface{}{
					"path":    relPath,
					"entries": entries,
				},
				Meta: &tools.ResultMeta{
					DurationMs: time.Since(startTime).Milliseconds(),
				},
			}, nil
		},
	}
}

type DirEntry struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Size  int64  `json:"size"`
	MTime int64  `json:"mtime"`
}

func listDirRecursive(path string, depth int, includeHidden bool, projectRoot string) ([]DirEntry, error) {
	var entries []DirEntry

	if depth <= 0 {
		return entries, nil
	}

	dirEntries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, entry := range dirEntries {
		name := entry.Name()
		if !includeHidden && strings.HasPrefix(name, ".") {
			continue
		}

		entryPath := filepath.Join(path, name)
		info, err := entry.Info()
		if err != nil {
			continue
		}

		entryType := "file"
		if entry.IsDir() {
			entryType = "dir"
		} else if entry.Type()&os.ModeSymlink != 0 {
			entryType = "symlink"
		}

		entries = append(entries, DirEntry{
			Name:  name,
			Type:  entryType,
			Size:  info.Size(),
			MTime: info.ModTime().Unix(),
		})

		if entry.IsDir() && depth > 1 {
			subEntries, err := listDirRecursive(entryPath, depth-1, includeHidden, projectRoot)
			if err == nil {
				for _, se := range subEntries {
					entries = append(entries, DirEntry{
						Name:  filepath.Join(name, se.Name),
						Type:  se.Type,
						Size:  se.Size,
						MTime: se.MTime,
					})
				}
			}
		}
	}

	return entries, nil
}

func normalizePath(userPath, projectRoot string) (string, error) {
	userPath = strings.TrimSpace(userPath)
	if userPath == "" || userPath == "." {
		return projectRoot, nil
	}

	if strings.Contains(userPath, "..") {
		return "", os.ErrInvalid
	}

	cleanPath := filepath.Clean(userPath)
	if filepath.IsAbs(cleanPath) {
		return cleanPath, nil
	}

	joined := filepath.Join(projectRoot, cleanPath)
	if strings.HasPrefix(cleanPath, projectRoot) || strings.HasPrefix(joined, projectRoot) {
		return cleanPath, nil
	}

	return joined, nil
}
