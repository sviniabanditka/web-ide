package builtin

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/webide/ide/backend/internal/ai/tools"
)

func ReadFile() tools.Tool {
	return tools.Tool{
		Name:        "read_file",
		Description: "Read file contents with optional line range and size limits",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type": "string",
				},
				"max_bytes": map[string]interface{}{
					"type":    "integer",
					"default": 65536,
					"minimum": 1,
					"maximum": 262144,
				},
				"start_line": map[string]interface{}{
					"type": "integer",
				},
				"end_line": map[string]interface{}{
					"type": "integer",
				},
			},
			"required": []string{"path"},
		},
		Policy: tools.PolicyAllow,
		Execute: func(ctx context.Context, args map[string]interface{}, tc tools.ToolContext) (tools.ToolResult, error) {
			startTime := time.Now()

			path, _ := args["path"].(string)
			if path == "" {
				return tools.NewErrorResult(tools.ErrCodeValidation, "path is required", nil), nil
			}

			path, err := normalizePath(path, tc.ProjectRoot)
			if err != nil {
				return tools.NewErrorResult(tools.ErrCodeInvalidPath, err.Error(), nil), nil
			}

			if !strings.HasPrefix(path, tc.ProjectRoot) {
				return tools.NewErrorResult(tools.ErrCodePermission, "path outside project", nil), nil
			}

			maxBytes := 65536
			if mb, ok := args["max_bytes"].(float64); ok {
				maxBytes = int(mb)
			}
			if maxBytes > 262144 {
				maxBytes = 262144
			}

			file, err := os.Open(path)
			if err != nil {
				return tools.NewErrorResult(tools.ErrCodeNotFound, "file not found: "+path, nil), nil
			}
			defer file.Close()

			info, err := file.Stat()
			if err != nil {
				return tools.NewErrorResult(tools.ErrCodeExecution, err.Error(), nil), nil
			}

			if info.Size() > int64(tc.Limits.MaxFileBytes) {
				return tools.NewErrorResult(tools.ErrCodeSizeLimit, "file too large", map[string]interface{}{
					"size":     info.Size(),
					"max_size": tc.Limits.MaxFileBytes,
				}), nil
			}

			var startLine, endLine *int
			if sl, ok := args["start_line"].(float64); ok {
				l := int(sl)
				startLine = &l
			}
			if el, ok := args["end_line"].(float64); ok {
				l := int(el)
				endLine = &l
			}

			var content string
			var truncated bool
			var lineStart, lineEnd int

			if startLine != nil || endLine != nil {
				content, truncated, lineStart, lineEnd, err = readFileLines(file, *startLine, *endLine, maxBytes)
			} else {
				content, truncated, err = readFileContent(file, maxBytes)
			}

			if err != nil {
				return tools.NewErrorResult(tools.ErrCodeExecution, err.Error(), nil), nil
			}

			file.Seek(0, 0)
			hash := sha256.New()
			io.Copy(hash, file)
			sha := hex.EncodeToString(hash.Sum(nil))

			relPath, _ := filepath.Rel(tc.ProjectRoot, path)

			result := map[string]interface{}{
				"path":      relPath,
				"sha":       sha,
				"content":   content,
				"truncated": truncated,
			}

			if startLine != nil || endLine != nil {
				result["line_start"] = lineStart
				result["line_end"] = lineEnd
			}

			return tools.ToolResult{
				OK:   true,
				Data: result,
				Meta: &tools.ResultMeta{
					DurationMs: time.Since(startTime).Milliseconds(),
				},
			}, nil
		},
	}
}

func readFileContent(file *os.File, maxBytes int) (string, bool, error) {
	content, err := io.ReadAll(io.LimitReader(file, int64(maxBytes)))
	if err != nil {
		return "", false, err
	}

	truncated := false
	if fileInfo, err := file.Stat(); err == nil && fileInfo.Size() > int64(len(content)) {
		truncated = true
	}

	return string(content), truncated, nil
}

func readFileContentFull(file *os.File) (string, error) {
	content, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func readFileLines(file *os.File, startLine, endLine, maxBytes int) (string, bool, int, int, error) {
	fullContent, err := readFileContentFull(file)
	if err != nil {
		return "", false, 0, 0, err
	}

	lines := strings.Split(fullContent, "\n")

	if startLine <= 0 {
		startLine = 1
	}
	if endLine <= 0 || endLine > len(lines) {
		endLine = len(lines)
	}

	if startLine > endLine {
		startLine = endLine
	}

	totalLines := len(lines)
	selectedLines := lines[startLine-1 : endLine]

	var content bytes.Buffer
	var totalLen int
	for i, line := range selectedLines {
		if totalLen > maxBytes {
			selectedLines = selectedLines[:i]
			break
		}
		content.WriteString(line)
		if i < len(selectedLines)-1 {
			content.WriteString("\n")
		}
		totalLen += len(line) + 1
	}

	truncated := false
	if endLine < totalLines || len(content.Bytes()) > maxBytes {
		truncated = true
	}

	return content.String(), truncated, startLine, endLine, nil
}
