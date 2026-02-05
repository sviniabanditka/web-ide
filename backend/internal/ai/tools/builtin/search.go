package builtin

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/webide/ide/backend/internal/ai/tools"
)

func SearchInFiles() tools.Tool {
	return tools.Tool{
		Name:        "search_in_files",
		Description: "Search for text patterns in files with optional glob filtering",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type": "string",
				},
				"globs": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "string",
					},
				},
				"max_results": map[string]interface{}{
					"type":    "integer",
					"default": 50,
					"minimum": 1,
					"maximum": 200,
				},
			},
			"required": []string{"query"},
		},
		Policy: tools.PolicyAllow,
		Execute: func(ctx context.Context, args map[string]interface{}, tc tools.ToolContext) (tools.ToolResult, error) {
			startTime := time.Now()

			query, _ := args["query"].(string)
			if query == "" {
				return tools.NewErrorResult(tools.ErrCodeValidation, "query is required", nil), nil
			}

			maxResults := 50
			if mr, ok := args["max_results"].(float64); ok {
				maxResults = int(mr)
			}
			if maxResults > tc.Limits.MaxSearchResults {
				maxResults = tc.Limits.MaxSearchResults
			}

			var globs []string
			if g, ok := args["globs"].([]interface{}); ok {
				for _, item := range g {
					if s, ok := item.(string); ok {
						globs = append(globs, s)
					}
				}
			}

			pattern, err := regexp.Compile("(?i)" + regexp.QuoteMeta(query))
			if err != nil {
				return tools.NewErrorResult(tools.ErrCodeValidation, "invalid regex pattern", nil), nil
			}

			matches, truncated, err := searchInDir(tc.ProjectRoot, pattern, globs, maxResults, tc.ProjectRoot)
			if err != nil {
				return tools.NewErrorResult(tools.ErrCodeExecution, err.Error(), nil), nil
			}

			return tools.ToolResult{
				OK: true,
				Data: map[string]interface{}{
					"query":     query,
					"matches":   matches,
					"truncated": truncated,
				},
				Meta: &tools.ResultMeta{
					DurationMs: time.Since(startTime).Milliseconds(),
				},
			}, nil
		},
	}
}

type SearchMatch struct {
	Path    string `json:"path"`
	Line    int    `json:"line"`
	Col     int    `json:"col"`
	Preview string `json:"preview"`
}

func searchInDir(root string, pattern *regexp.Regexp, globs []string, maxResults int, projectRoot string) ([]SearchMatch, bool, error) {
	var matches []SearchMatch
	var truncated bool

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		if len(globs) > 0 {
			matched := false
			for _, g := range globs {
				if matched, _ = filepath.Match(g, filepath.Base(path)); matched {
					break
				}
				if matched, _ = filepath.Match(g, path); matched {
					break
				}
			}
			if !matched {
				return nil
			}
		}

		fileMatches, err := searchInFile(path, pattern, projectRoot, maxResults-len(matches))
		if err != nil {
			return nil
		}

		for _, m := range fileMatches {
			if len(matches) >= maxResults {
				truncated = true
				return filepath.SkipDir
			}
			matches = append(matches, m)
		}

		return nil
	})

	return matches, truncated, err
}

func searchInFile(path string, pattern *regexp.Regexp, projectRoot string, limit int) ([]SearchMatch, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var matches []SearchMatch

	lineNum := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		indices := pattern.FindAllStringIndex(line, -1)
		for _, idx := range indices {
			if len(matches) >= limit {
				return matches, nil
			}

			preview := line
			if len(preview) > 200 {
				col := idx[0]
				start := 0
				if col > 100 {
					start = col - 100
				}
				preview = "..." + line[start:min(len(line), start+200)] + "..."
			}

			matches = append(matches, SearchMatch{
				Path:    getRelativePath(path, projectRoot),
				Line:    lineNum,
				Col:     idx[0] + 1,
				Preview: preview,
			})
		}
	}

	return matches, scanner.Err()
}

func getRelativePath(absPath, projectRoot string) string {
	rel, err := filepath.Rel(projectRoot, absPath)
	if err != nil {
		return absPath
	}
	return rel
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
