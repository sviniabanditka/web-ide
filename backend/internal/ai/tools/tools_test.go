package tools_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/webide/ide/backend/internal/ai/tools"
)

func TestPathGuard_ResolveProjectPath(t *testing.T) {
	projectDir := t.TempDir()

	guard := tools.NewPathGuard(projectDir, tools.ToolLimits{
		MaxFileBytes:     1024 * 1024,
		MaxOutputBytes:   1024 * 1024,
		MaxSearchResults: 100,
		MaxPatchFiles:    10,
		MaxToolTime:      5 * time.Minute,
	})

	tests := []struct {
		name        string
		userPath    string
		expectError bool
	}{
		{
			name:        "empty path",
			userPath:    "",
			expectError: true,
		},
		{
			name:        "path traversal attempt",
			userPath:    "../etc/passwd",
			expectError: true,
		},
		{
			name:        "path traversal in middle",
			userPath:    "foo/../../etc/passwd",
			expectError: true,
		},
		{
			name:        "absolute path outside",
			userPath:    "/etc/passwd",
			expectError: true,
		},
		{
			name:        "proc access forbidden",
			userPath:    "/proc/self/status",
			expectError: true,
		},
		{
			name:        "sys access forbidden",
			userPath:    "/sys/kernel",
			expectError: true,
		},
		{
			name:        "dev access forbidden",
			userPath:    "/dev/null",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := guard.ResolveProjectPath(tt.userPath)
			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestPathGuard_SymlinkEscape(t *testing.T) {
	projectDir := t.TempDir()
	secretDir := filepath.Join(projectDir, "secret")
	os.MkdirAll(secretDir, 0755)
	os.WriteFile(filepath.Join(secretDir, "password.txt"), []byte("secret"), 0644)

	guard := tools.NewPathGuard(projectDir, tools.ToolLimits{
		MaxFileBytes: 1024 * 1024,
	})

	result, err := guard.ResolveProjectPath("secret/password.txt")
	if err != nil {
		t.Errorf("symlink should resolve within project: %v", err)
	}

	if !filepath.HasPrefix(result, projectDir) {
		t.Error("symlink escaped project directory")
	}
}

func TestPathGuard_ValidateFileAccess(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("test content"), 0644)

	guard := tools.NewPathGuard(tmpDir, tools.ToolLimits{
		MaxFileBytes: 100,
	})

	err := guard.ValidateFileAccess(testFile)
	if err != nil {
		t.Errorf("file should be accessible: %v", err)
	}

	largeFile := filepath.Join(tmpDir, "large.txt")
	os.WriteFile(largeFile, make([]byte, 200), 0644)

	err = guard.ValidateFileAccess(largeFile)
	if err != tools.ErrFileTooLarge {
		t.Errorf("expected file too large error, got: %v", err)
	}
}

func TestPathGuard_ValidateDirAccess(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	os.MkdirAll(subDir, 0755)

	guard := tools.NewPathGuard(tmpDir, tools.ToolLimits{})

	err := guard.ValidateDirAccess(subDir)
	if err != nil {
		t.Errorf("directory should be accessible: %v", err)
	}
}

func TestToolRegistry_Register(t *testing.T) {
	registry := tools.NewRegistry()

	tool := tools.Tool{
		Name:        "test_tool",
		Description: "A test tool",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"arg1": map[string]interface{}{
					"type": "string",
				},
			},
			"required": []string{"arg1"},
		},
		Policy: tools.PolicyAllow,
		Execute: func(ctx context.Context, args map[string]interface{}, tc tools.ToolContext) (tools.ToolResult, error) {
			return tools.ToolResult{OK: true}, nil
		},
	}

	err := registry.Register(tool)
	if err != nil {
		t.Errorf("failed to register tool: %v", err)
	}

	retrieved, ok := registry.Get("test_tool")
	if !ok {
		t.Error("tool should be retrievable")
	}
	if retrieved.Name != "test_tool" {
		t.Error("retrieved tool has wrong name")
	}
}

func TestToolRegistry_ValidateArgs(t *testing.T) {
	registry := tools.NewRegistry()

	tool := tools.Tool{
		Name:        "validate_test",
		Description: "Test validation",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type": "string",
				},
				"age": map[string]interface{}{
					"type":    "integer",
					"minimum": 0,
				},
			},
			"required": []string{"name"},
		},
		Policy: tools.PolicyAllow,
		Execute: func(ctx context.Context, args map[string]interface{}, tc tools.ToolContext) (tools.ToolResult, error) {
			return tools.ToolResult{OK: true}, nil
		},
	}

	registry.Register(tool)

	err := registry.ValidateArgs("validate_test", map[string]interface{}{
		"name": "John",
		"age":  25,
	})
	if err != nil {
		t.Errorf("valid args should not error: %v", err)
	}

	err = registry.ValidateArgs("validate_test", map[string]interface{}{
		"age": -5,
	})
	if err == nil {
		t.Error("invalid age should error")
	}

	err = registry.ValidateArgs("validate_test", map[string]interface{}{
		"age": "not a number",
	})
	if err == nil {
		t.Error("wrong type should error")
	}
}

func TestToolRegistry_ListForModel(t *testing.T) {
	registry := tools.NewRegistry()

	registry.Register(tools.Tool{
		Name:        "tool1",
		Description: "Tool 1",
		Parameters:  map[string]interface{}{},
		Policy:      tools.PolicyAllow,
		Execute: func(ctx context.Context, args map[string]interface{}, tc tools.ToolContext) (tools.ToolResult, error) {
			return tools.ToolResult{OK: true}, nil
		},
	})

	registry.Register(tools.Tool{
		Name:        "tool2",
		Description: "Tool 2",
		Parameters:  map[string]interface{}{},
		Policy:      tools.PolicyConfirm,
		Execute: func(ctx context.Context, args map[string]interface{}, tc tools.ToolContext) (tools.ToolResult, error) {
			return tools.ToolResult{OK: true}, nil
		},
	})

	defs := registry.ListForModel()
	if len(defs) != 2 {
		t.Errorf("expected 2 tool definitions, got %d", len(defs))
	}
}
