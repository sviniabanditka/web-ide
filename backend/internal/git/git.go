package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type GitStatus struct {
	Changed   []FileStatus `json:"changed"`
	Untracked []string     `json:"untracked"`
	Staged    []FileStatus `json:"staged"`
}

type FileStatus struct {
	Path   string `json:"path"`
	Status string `json:"status"` // "M" = modified, "A" = added, "D" = deleted, "?" = untracked
}

type GitResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func RunGit(repoDir string, timeout time.Duration, args ...string) (*GitResult, error) {
	if _, statErr := os.Stat(repoDir); os.IsNotExist(statErr) {
		return nil, statErr
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = repoDir
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()

	var runErr error
	select {
	case <-time.After(timeout):
		cmd.Process.Kill()
		return &GitResult{
			ExitCode: -1,
			Stderr:   "timeout",
		}, nil
	case runErr = <-done:
		_ = runErr
	}

	return &GitResult{
		Stdout:   strings.TrimSuffix(stdout.String(), "\n"),
		Stderr:   strings.TrimSuffix(stderr.String(), "\n"),
		ExitCode: 0,
	}, nil
}

func GetStatus(repoDir string) (*GitStatus, error) {
	result, runErr := RunGit(repoDir, 10*time.Second, "status", "--short")
	if runErr != nil {
		return nil, runErr
	}

	status := &GitStatus{
		Changed:   []FileStatus{},
		Untracked: []string{},
		Staged:    []FileStatus{},
	}

	lines := strings.Split(result.Stdout, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if len(line) < 3 {
			continue
		}

		staged := line[0:1]
		unstaged := line[1:2]
		path := strings.TrimSpace(line[2:])

		// On macOS, format is "XY path" without leading space
		// "M " or "M" (at pos 0) = modified in working tree (not staged)
		// "??" = untracked
		// "A ", "M " etc at pos 0 = staged

		// Untracked: ?? or ? at pos 0 and ?
		if staged == "?" && unstaged == "?" {
			status.Untracked = append(status.Untracked, path)
			continue
		}

		// Modified in working tree: "M " or just "M" (single M at start)
		if staged == "M" && unstaged == " " {
			status.Changed = append(status.Changed, FileStatus{
				Path:   path,
				Status: "M",
			})
			continue
		}

		// Staged changes (A, M, D at pos 0 with space after)
		if (staged == "A" || staged == "M" || staged == "D" || staged == "R") && unstaged == " " {
			status.Staged = append(status.Staged, FileStatus{
				Path:   path,
				Status: staged,
			})
			continue
		}

		// Both staged and modified: "MM" or "AM" etc
		if staged != " " && staged != "" && unstaged != " " && unstaged != "" {
			status.Staged = append(status.Staged, FileStatus{
				Path:   path,
				Status: staged,
			})
			status.Changed = append(status.Changed, FileStatus{
				Path:   path,
				Status: unstaged,
			})
			continue
		}

		// Untracked single ?
		if staged == "?" && unstaged != "?" {
			status.Untracked = append(status.Untracked, path)
			continue
		}
	}

	return status, nil
}

func GetDiff(repoDir string, cached bool, filePath string) (string, error) {
	args := []string{"diff", "--patch", "--no-color"}
	if cached {
		args = append(args, "--cached")
	}
	if filePath != "" {
		args = append(args, "--", filePath)
	}

	result, err := RunGit(repoDir, 10*time.Second, args...)
	if err != nil {
		return "", err
	}

	return result.Stdout, nil
}

func StageFiles(repoDir string, paths []string) error {
	args := append([]string{"add", "--"}, paths...)
	_, err := RunGit(repoDir, 10*time.Second, args...)
	return err
}

func UnstageFiles(repoDir string, paths []string) error {
	args := append([]string{"reset", "--"}, paths...)
	_, err := RunGit(repoDir, 10*time.Second, args...)
	return err
}

func Commit(repoDir, message string) error {
	result, err := RunGit(repoDir, 30*time.Second, "commit", "-m", message)
	if err != nil {
		return err
	}
	if result.ExitCode != 0 && result.Stderr != "" {
		return &GitError{Message: result.Stderr}
	}
	return nil
}

func Push(repoDir string, remote string, branch string) error {
	args := []string{"push"}
	if remote != "" {
		args = append(args, remote)
	}
	if branch != "" {
		args = append(args, branch)
	}

	result, err := RunGit(repoDir, 60*time.Second, args...)
	if err != nil {
		return err
	}
	if result.ExitCode != 0 && result.Stderr != "" {
		return &GitError{Message: result.Stderr}
	}
	return nil
}

func GetBranches(repoDir string) ([]string, error) {
	result, err := RunGit(repoDir, 10*time.Second, "branch", "-a", "--format=%(refname:short)")
	if err != nil {
		return nil, err
	}

	branches := strings.Split(result.Stdout, "\n")
	var validBranches []string
	for _, b := range branches {
		b = strings.TrimSpace(b)
		if b != "" {
			validBranches = append(validBranches, b)
		}
	}

	return validBranches, nil
}

func GetLog(repoDir string, limit int) ([]GitLogEntry, error) {
	result, err := RunGit(repoDir, 10*time.Second, "log", "--pretty=format:%H|%an|%ae|%ad|%s", "--date=iso", "-n", strconv.Itoa(limit))
	if err != nil {
		return nil, err
	}

	lines := strings.Split(result.Stdout, "\n")
	var entries []GitLogEntry
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 5)
		if len(parts) >= 5 {
			entries = append(entries, GitLogEntry{
				Hash:    parts[0],
				Author:  parts[1],
				Email:   parts[2],
				Date:    parts[3],
				Subject: parts[4],
			})
		}
	}

	return entries, nil
}

type GitLogEntry struct {
	Hash    string `json:"hash"`
	Author  string `json:"author"`
	Email   string `json:"email"`
	Date    string `json:"date"`
	Subject string `json:"subject"`
}

type GitError struct {
	Message string
}

func (e *GitError) Error() string {
	return e.Message
}

func IsGitRepo(repoDir string) bool {
	_, err := os.Stat(repoDir + "/.git")
	return err == nil
}

func ApplyPatch(repoDir, patch string) error {
	if patch == "" {
		return nil
	}

	tmpFile, err := os.CreateTemp("", "patch-*.patch")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(patch); err != nil {
		return fmt.Errorf("failed to write patch: %w", err)
	}
	tmpFile.Close()

	result, err := RunGit(repoDir, 30*time.Second, "apply", tmpFile.Name())
	if err != nil {
		return fmt.Errorf("failed to apply patch: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("patch failed: %s", result.Stderr)
	}

	return nil
}

func GetHeadCommit(repoDir string) (string, error) {
	result, err := RunGit(repoDir, 10*time.Second, "rev-parse", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(result.Stdout), nil
}
