package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

type ToolRegistry struct {
	mu      sync.RWMutex
	tools   map[string]Tool
	schemas map[string]*jsonschema.Schema
}

func NewRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools:   make(map[string]Tool),
		schemas: make(map[string]*jsonschema.Schema),
	}
}

func (r *ToolRegistry) Register(tool Tool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tools[tool.Name]; exists {
		return fmt.Errorf("tool already registered: %s", tool.Name)
	}

	r.tools[tool.Name] = tool

	if tool.Parameters != nil {
		schemaJSON, _ := json.Marshal(tool.Parameters)
		schema, err := jsonschema.CompileString("tool-schema.json", string(schemaJSON))
		if err != nil {
			return fmt.Errorf("invalid JSON schema for tool %s: %w", tool.Name, err)
		}
		r.schemas[tool.Name] = schema
	}

	return nil
}

func (r *ToolRegistry) Get(name string) (Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tool, ok := r.tools[name]
	return tool, ok
}

func (r *ToolRegistry) List() []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		result = append(result, tool)
	}
	return result
}

func (r *ToolRegistry) ListForModel() []ToolDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]ToolDefinition, 0, len(r.tools))
	for _, tool := range r.tools {
		result = append(result, MakeToolDefinition(
			tool.Name,
			tool.Description,
			tool.Parameters,
			tool.Policy,
		))
	}
	return result
}

func (r *ToolRegistry) ValidateArgs(name string, args map[string]interface{}) error {
	r.mu.RLock()
	schema, ok := r.schemas[name]
	r.mu.RUnlock()

	if !ok {
		return nil
	}

	err := schema.Validate(args)
	if err == nil {
		return nil
	}

	return fmt.Errorf("validation failed: %s", err.Error())
}

func (r *ToolRegistry) Remove(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.tools, name)
	delete(r.schemas, name)
}

var GlobalRegistry *ToolRegistry

func init() {
	GlobalRegistry = NewRegistry()
}

func GlobalExecute(ctx context.Context, toolName string, args map[string]interface{}, tc ToolContext) ToolResult {
	tool, ok := GlobalRegistry.Get(toolName)
	if !ok {
		return ToolResult{
			OK: false,
			Error: &ToolError{
				Code:    ErrCodeNotFound,
				Message: "Tool not found: " + toolName,
			},
		}
	}

	result, err := tool.Execute(ctx, args, tc)
	if err != nil {
		return ToolResult{
			OK: false,
			Error: &ToolError{
				Code:    ErrCodeExecution,
				Message: err.Error(),
			},
		}
	}

	return result
}
