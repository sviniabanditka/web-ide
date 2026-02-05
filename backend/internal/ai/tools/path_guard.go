package tools

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type PathGuard struct {
	ProjectRoot  string
	MaxFileBytes int64
	Limits       ToolLimits
}

func NewPathGuard(projectRoot string, limits ToolLimits) *PathGuard {
	return &PathGuard{
		ProjectRoot:  projectRoot,
		MaxFileBytes: limits.MaxFileBytes,
		Limits:       limits,
	}
}

var (
	ErrPathOutsideProject = errors.New("path is outside project directory")
	ErrPathForbidden      = errors.New("access to this path is forbidden")
	ErrFileTooLarge       = errors.New("file exceeds size limit")
	ErrSymlinkEscape      = errors.New("symbolic link escapes project directory")
	ErrPathTraversal      = errors.New("path traversal attempt detected")
	ErrInvalidPath        = errors.New("invalid path")
)

func (g *PathGuard) ResolveProjectPath(userPath string) (string, error) {
	if userPath == "" {
		return "", ErrInvalidPath
	}

	userPath = filepath.Clean(userPath)

	if strings.Contains(userPath, "..") {
		return "", ErrPathTraversal
	}

	forbiddenPatterns := []string{
		"/proc/", "/proc", "/sys/", "/sys", "/dev/", "/dev",
		"/etc/passwd", "/etc/shadow", "/etc/sudoers",
	}
	for _, pattern := range forbiddenPatterns {
		if strings.HasPrefix(userPath, pattern) {
			return "", ErrPathForbidden
		}
	}

	evalProjectRoot, err := filepath.EvalSymlinks(g.ProjectRoot)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	if err == nil {
		evalProjectRoot = filepath.Clean(evalProjectRoot)
	} else {
		evalProjectRoot, _ = filepath.Abs(g.ProjectRoot)
	}

	var resolvedPath string
	if filepath.IsAbs(userPath) {
		resolvedPath = userPath
	} else {
		resolvedPath = filepath.Join(g.ProjectRoot, userPath)
	}

	resolvedPath = filepath.Clean(resolvedPath)

	evalPath, err := filepath.EvalSymlinks(resolvedPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	if err == nil {
		evalPath = filepath.Clean(evalPath)
	} else {
		evalPath = resolvedPath
	}

	if !strings.HasPrefix(evalPath, evalProjectRoot) {
		return "", ErrPathOutsideProject
	}

	return resolvedPath, nil
}

func (g *PathGuard) ValidateFileAccess(absPath string) error {
	info, err := os.Stat(absPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ErrPathOutsideProject
		}
		return err
	}

	if info.IsDir() {
		return ErrPathForbidden
	}

	if info.Size() > g.MaxFileBytes {
		return ErrFileTooLarge
	}

	return nil
}

func (g *PathGuard) ValidateDirAccess(absPath string) error {
	info, err := os.Stat(absPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ErrPathOutsideProject
		}
		return err
	}

	if !info.IsDir() {
		return ErrPathForbidden
	}

	return nil
}

func (g *PathGuard) CanReadFile(absPath string) bool {
	err := g.ValidateFileAccess(absPath)
	return err == nil
}

func (g *PathGuard) CanListDir(absPath string) bool {
	err := g.ValidateDirAccess(absPath)
	return err == nil
}

func (g *PathGuard) CanWriteFile(absPath string) bool {
	dir := filepath.Dir(absPath)
	return g.CanListDir(dir)
}

func (g *PathGuard) GetFileSize(absPath string) (int64, error) {
	info, err := os.Stat(absPath)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}
