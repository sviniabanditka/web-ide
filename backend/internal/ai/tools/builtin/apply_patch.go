package builtin

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pmezard/go-difflib/difflib"
	"github.com/webide/ide/backend/internal/ai/tools"
)

func ApplyPatch() tools.Tool {
	return tools.Tool{
		Name:        "apply_patch",
		Description: "Apply a unified diff patch to modify files",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"patch": map[string]interface{}{
					"type": "string",
				},
				"dry_run": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
			},
			"required": []string{"patch"},
		},
		Policy: tools.PolicyConfirm,
		Execute: func(ctx context.Context, args map[string]interface{}, tc tools.ToolContext) (tools.ToolResult, error) {
			startTime := time.Now()

			patch, _ := args["patch"].(string)
			if patch == "" {
				return tools.NewErrorResult(tools.ErrCodeValidation, "patch is required", nil), nil
			}

			dryRun := false
			if dr, ok := args["dry_run"].(bool); ok {
				dryRun = dr
			}

			patchSet, err := parsePatch(patch)
			if err != nil {
				return tools.NewErrorResult(tools.ErrCodeValidation, "invalid patch format", err.Error()), nil
			}

			if len(patchSet) > tc.Limits.MaxPatchFiles {
				return tools.NewErrorResult(tools.ErrCodeSizeLimit, "too many files in patch", map[string]interface{}{
					"files": len(patchSet),
					"max":   tc.Limits.MaxPatchFiles,
				}), nil
			}

			var applied []FileChange
			var rejects []Reject

			for _, patch := range patchSet {
				absPath := filepath.Join(tc.ProjectRoot, patch.File)

				_, err := os.Stat(absPath)
				if err != nil && !os.IsNotExist(err) {
					return tools.NewErrorResult(tools.ErrCodeExecution, err.Error(), nil), nil
				}

				var originalContent string
				if os.IsNotExist(err) {
					originalContent = ""
				} else {
					content, err := os.ReadFile(absPath)
					if err != nil {
						return tools.NewErrorResult(tools.ErrCodeExecution, "cannot read file: "+absPath, nil), nil
					}
					originalContent = string(content)
				}

				shaBefore := computeSHA(originalContent)

				patchedContent, hunksApplied, err := applyPatchToContent(originalContent, patch.Hunks)
				if err != nil {
					rejects = append(rejects, Reject{
						Path:   patch.File,
						Reason: err.Error(),
						Hunk:   patch.OriginalHunks,
					})
					continue
				}

				summary := ""
				if dryRun {
					diff := difflib.UnifiedDiff{
						A:        difflib.SplitLines(originalContent),
						B:        difflib.SplitLines(patchedContent),
						FromFile: patch.File,
						ToFile:   patch.File,
						Context:  3,
					}
					summary, _ = difflib.GetUnifiedDiffString(diff)
				}

				if !dryRun {
					err = os.WriteFile(absPath, []byte(patchedContent), 0644)
					if err != nil {
						rejects = append(rejects, Reject{
							Path:   patch.File,
							Reason: err.Error(),
							Hunk:   patch.OriginalHunks,
						})
						continue
					}
				}

				shaAfter := computeSHA(patchedContent)
				applied = append(applied, FileChange{
					Path:      patch.File,
					SHABefore: shaBefore,
					SHAAfter:  shaAfter,
				})

				if summary == "" && !dryRun {
					summary = fmt.Sprintf("Applied %d hunks to %s", hunksApplied, patch.File)
				}
				if summary != "" && dryRun {
					_ = summary
				}
			}

			result := map[string]interface{}{
				"applied": applied,
				"rejects": rejects,
			}

			if dryRun {
				result["preview_summary"] = fmt.Sprintf("Would modify %d files, %d hunks", len(applied), len(applied))
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

type FileChange struct {
	Path      string `json:"path"`
	SHABefore string `json:"sha_before"`
	SHAAfter  string `json:"sha_after"`
}

type Reject struct {
	Path   string `json:"path"`
	Reason string `json:"reason"`
	Hunk   string `json:"hunk"`
}

type ParsedPatch struct {
	File          string
	OriginalHunks string
	Hunks         []PatchHunk
}

type PatchHunk struct {
	OrigStart, OrigCount int
	NewStart, NewCount   int
	Lines                []string
}

func parsePatch(patchStr string) ([]ParsedPatch, error) {
	var patches []ParsedPatch
	var currentPatch *ParsedPatch

	scanner := bufio.NewScanner(strings.NewReader(patchStr))
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "--- ") {
			if currentPatch != nil {
				patches = append(patches, *currentPatch)
			}
			currentPatch = &ParsedPatch{
				OriginalHunks: line + "\n",
			}
			continue
		}

		if strings.HasPrefix(line, "+++ ") {
			if currentPatch != nil {
				currentPatch.OriginalHunks += line + "\n"
				filePath := strings.TrimPrefix(line, "+++ ")
				filePath = strings.TrimSpace(filePath)
				if strings.HasPrefix(filePath, "b/") {
					filePath = strings.TrimPrefix(filePath, "b/")
				}
				currentPatch.File = filePath
			}
			continue
		}

		if strings.HasPrefix(line, "@@ -") && currentPatch != nil {
			currentPatch.OriginalHunks += line + "\n"
			hunk, err := parseHunk(line)
			if err != nil {
				return nil, err
			}
			currentPatch.Hunks = append(currentPatch.Hunks, hunk)
			continue
		}

		if currentPatch != nil {
			currentPatch.OriginalHunks += line + "\n"
		}
	}

	if currentPatch != nil {
		patches = append(patches, *currentPatch)
	}

	return patches, scanner.Err()
}

func parseHunk(line string) (PatchHunk, error) {
	var hunk PatchHunk

	_, err := fmt.Sscanf(line, "@@ -%d,%d +%d,%d @@",
		&hunk.OrigStart, &hunk.OrigCount,
		&hunk.NewStart, &hunk.NewCount)

	return hunk, err
}

func applyPatchToContent(content string, hunks []PatchHunk) (string, int, error) {
	lines := strings.Split(content, "\n")
	var result []string
	hunkIdx := 0
	origLine := 0
	newLine := 0
	hunksApplied := 0

	for hunkIdx < len(hunks) {
		hunk := hunks[hunkIdx]

		skipLines := hunk.OrigStart - origLine - 1
		for i := 0; i < skipLines && origLine < len(lines); i++ {
			result = append(result, lines[origLine])
			origLine++
			newLine++
		}

		for _, hunkLine := range hunk.Lines {
			if strings.HasPrefix(hunkLine, "-") {
				origLine++
			} else if strings.HasPrefix(hunkLine, "+") {
				result = append(result, strings.TrimPrefix(hunkLine, "+"))
				newLine++
			} else if strings.HasPrefix(hunkLine, "\\") {
			} else {
				result = append(result, lines[origLine])
				origLine++
				newLine++
			}
		}

		hunksApplied++
		hunkIdx++
	}

	for origLine < len(lines) {
		result = append(result, lines[origLine])
		origLine++
	}

	return strings.Join(result, "\n"), hunksApplied, nil
}

func computeSHA(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}
