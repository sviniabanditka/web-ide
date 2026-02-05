package agent

type MessageRole string

const (
	RoleSystem    MessageRole = "system"
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
	RoleTool      MessageRole = "tool"
)

type ModelMessage struct {
	ID         string      `json:"id"`
	Role       MessageRole `json:"role"`
	Content    string      `json:"content,omitempty"`
	ToolCalls  []ToolCall  `json:"tool_calls,omitempty"`
	ToolCallID string      `json:"tool_call_id,omitempty"`
	Name       string      `json:"name,omitempty"`
	Timestamp  int64       `json:"timestamp"`
}

type ToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function ToolCallFunction `json:"function"`
	Index    int              `json:"index"`
}

type ToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type ToolResultMessage struct {
	ToolCallID string      `json:"tool_call_id"`
	Role       MessageRole `json:"role"`
	Name       string      `json:"name"`
	Content    string      `json:"content"`
	Timestamp  int64       `json:"timestamp"`
}

func NewUserMessage(content string) ModelMessage {
	return ModelMessage{
		ID:        generateID(),
		Role:      RoleUser,
		Content:   content,
		Timestamp: now(),
	}
}

func NewSystemMessage(content string) ModelMessage {
	return ModelMessage{
		ID:        generateID(),
		Role:      RoleSystem,
		Content:   content,
		Timestamp: now(),
	}
}

func NewAssistantMessage(content string, toolCalls []ToolCall) ModelMessage {
	var tcJSON []map[string]interface{}
	for _, tc := range toolCalls {
		tcJSON = append(tcJSON, map[string]interface{}{
			"id":   tc.ID,
			"type": tc.Type,
			"function": map[string]interface{}{
				"name":      tc.Function.Name,
				"arguments": tc.Function.Arguments,
			},
		})
	}
	return ModelMessage{
		ID:        generateID(),
		Role:      RoleAssistant,
		Content:   content,
		ToolCalls: toolCalls,
		Timestamp: now(),
	}
}

func NewToolResultMessage(toolCallID, name, content string) ModelMessage {
	return ModelMessage{
		ToolCallID: toolCallID,
		Role:       RoleTool,
		Name:       name,
		Content:    content,
		Timestamp:  now(),
	}
}
