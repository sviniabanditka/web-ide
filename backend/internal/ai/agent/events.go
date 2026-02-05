package agent

import "time"

type WSEvent struct {
	Type      string      `json:"type"`
	SessionID string      `json:"sessionId"`
	ProjectID string      `json:"projectId"`
	TS        time.Time   `json:"ts"`
	ID        string      `json:"id,omitempty"`
	Payload   interface{} `json:"payload,omitempty"`
}

const (
	EventToolCall             = "tool.call"
	EventToolApprovalRequired = "tool.approval_required"
	EventToolResult           = "tool.result"
	EventToolError            = "tool.error"
	EventCommandOutput        = "command.output"
	EventCommandDone          = "command.done"
	EventAgentDone            = "agent.done"
	EventAgentError           = "agent.error"
)

type ToolCallPayload struct {
	ToolCallID string                 `json:"id"`
	Name       string                 `json:"name"`
	Arguments  map[string]interface{} `json:"arguments"`
}

type ToolApprovalPayload struct {
	ToolCallID string                 `json:"id"`
	Name       string                 `json:"name"`
	Arguments  map[string]interface{} `json:"arguments"`
	Summary    string                 `json:"summary"`
	Policy     string                 `json:"policy"`
}

type ToolResultPayload struct {
	ToolCallID string      `json:"id"`
	Name       string      `json:"name"`
	OK         bool        `json:"ok"`
	Result     interface{} `json:"result,omitempty"`
	Error      *ToolError  `json:"error,omitempty"`
	DurationMs int64       `json:"duration_ms,omitempty"`
}

type ToolError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

type CommandOutputPayload struct {
	Handle string `json:"handle"`
	Stream string `json:"stream"`
	Text   string `json:"text"`
	TS     int64  `json:"ts"`
}

type CommandDonePayload struct {
	Handle   string `json:"handle"`
	ExitCode int    `json:"exit_code"`
}

type AgentDonePayload struct {
	Steps    int    `json:"steps"`
	FinalMsg string `json:"final_message"`
}

type AgentErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewToolCallEvent(sessionID, projectID string, payload ToolCallPayload) WSEvent {
	return WSEvent{
		Type:      EventToolCall,
		SessionID: sessionID,
		ProjectID: projectID,
		TS:        time.Now(),
		ID:        payload.ToolCallID,
		Payload:   payload,
	}
}

func NewToolApprovalRequiredEvent(sessionID, projectID string, payload ToolApprovalPayload) WSEvent {
	return WSEvent{
		Type:      EventToolApprovalRequired,
		SessionID: sessionID,
		ProjectID: projectID,
		TS:        time.Now(),
		ID:        payload.ToolCallID,
		Payload:   payload,
	}
}

func NewToolResultEvent(sessionID, projectID string, payload ToolResultPayload) WSEvent {
	return WSEvent{
		Type:      EventToolResult,
		SessionID: sessionID,
		ProjectID: projectID,
		TS:        time.Now(),
		ID:        payload.ToolCallID,
		Payload:   payload,
	}
}

func NewToolErrorEvent(sessionID, projectID string, toolCallID string, err ToolError) WSEvent {
	return WSEvent{
		Type:      EventToolError,
		SessionID: sessionID,
		ProjectID: projectID,
		TS:        time.Now(),
		ID:        toolCallID,
		Payload: map[string]interface{}{
			"error": err,
		},
	}
}

func NewCommandOutputEvent(sessionID, projectID string, payload CommandOutputPayload) WSEvent {
	return WSEvent{
		Type:      EventCommandOutput,
		SessionID: sessionID,
		ProjectID: projectID,
		TS:        time.Now(),
		Payload:   payload,
	}
}

func NewCommandDoneEvent(sessionID, projectID string, payload CommandDonePayload) WSEvent {
	return WSEvent{
		Type:      EventCommandDone,
		SessionID: sessionID,
		ProjectID: projectID,
		TS:        time.Now(),
		Payload:   payload,
	}
}
