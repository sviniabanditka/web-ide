package provider

import "context"

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Response struct {
	Content    string     `json:"content"`
	Usage      TokenUsage `json:"usage"`
	StopReason string     `json:"stop_reason"`
}

type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type Config struct {
	URL         string  `json:"url"`
	APIKey      string  `json:"api_key"`
	Model       string  `json:"model"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
}

type Chunk struct {
	Content string
	Done    bool
}

type ToolDefinition struct {
	Type     string                 `json:"type"`
	Function map[string]interface{} `json:"function"`
}

type ToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type StreamChunk struct {
	Content       string     `json:"content,omitempty"`
	ToolCalls     []ToolCall `json:"tool_calls,omitempty"`
	ToolCallIndex int        `json:"tool_call_index,omitempty"`
	Done          bool       `json:"done"`
}

type Provider interface {
	Complete(ctx context.Context, messages []Message, cfg Config) (*Response, error)
	Stream(ctx context.Context, messages []Message, cfg Config) (<-chan Chunk, error)
	StreamWithTools(ctx context.Context, messages []Message, cfg Config, tools []ToolDefinition, toolChoice string) (<-chan StreamChunk, error)
	Name() string
}
