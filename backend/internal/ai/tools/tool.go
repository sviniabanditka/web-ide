package tools

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ToolPolicy string

const (
	PolicyAllow   ToolPolicy = "allow"
	PolicyConfirm ToolPolicy = "confirm"
	PolicyDeny    ToolPolicy = "deny"
)

type ToolContext struct {
	SessionID   uuid.UUID
	ProjectID   uuid.UUID
	UserID      uuid.UUID
	ProjectRoot string
	Mode        string
	Limits      ToolLimits
}

type ToolLimits struct {
	MaxFileBytes     int64
	MaxOutputBytes   int64
	MaxSearchResults int
	MaxPatchFiles    int
	MaxToolTime      time.Duration
}

type Tool struct {
	Name        string
	Description string
	Parameters  map[string]interface{}
	Policy      ToolPolicy
	Execute     func(ctx context.Context, args map[string]interface{}, tc ToolContext) (ToolResult, error)
}

type ToolDefinition struct {
	Type     string                 `json:"type"`
	Function map[string]interface{} `json:"function"`
}

func MakeToolDefinition(name, description string, params map[string]interface{}, policy ToolPolicy) ToolDefinition {
	return ToolDefinition{
		Type: "function",
		Function: map[string]interface{}{
			"name":        name,
			"description": description,
			"parameters":  params,
		},
	}
}
