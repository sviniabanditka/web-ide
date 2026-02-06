package provider

import (
	"context"
	"log"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

type Anthropic struct {
	client  anthropic.Client
	baseURL string
}

func NewAnthropic(apiKey, baseURL string) *Anthropic {
	client := anthropic.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(baseURL),
	)
	return &Anthropic{client: client, baseURL: baseURL}
}

func (a *Anthropic) Complete(ctx context.Context, messages []Message, cfg Config) (*Response, error) {
	apiMessages := convertMessages(messages)

	msg, err := a.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(cfg.Model),
		Messages:  apiMessages,
		MaxTokens: int64(cfg.MaxTokens),
	})
	if err != nil {
		return nil, err
	}

	content := ""
	for _, block := range msg.Content {
		if tb, ok := block.AsAny().(anthropic.TextBlock); ok {
			content += tb.Text
		}
	}

	return &Response{
		Content: content,
		Usage: TokenUsage{
			PromptTokens:     int(msg.Usage.InputTokens),
			CompletionTokens: int(msg.Usage.OutputTokens),
			TotalTokens:      int(msg.Usage.InputTokens + msg.Usage.OutputTokens),
		},
		StopReason: string(msg.StopReason),
	}, nil
}

func (a *Anthropic) Stream(ctx context.Context, messages []Message, cfg Config) (<-chan Chunk, error) {
	apiMessages := convertMessages(messages)

	stream := a.client.Messages.NewStreaming(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(cfg.Model),
		Messages:  apiMessages,
		MaxTokens: int64(cfg.MaxTokens),
	})

	ch := make(chan Chunk, 100)

	go func() {
		defer close(ch)

		for stream.Next() {
			event := stream.Current()

			switch eventVariant := event.AsAny().(type) {
			case anthropic.ContentBlockDeltaEvent:
				switch deltaVariant := eventVariant.Delta.AsAny().(type) {
				case anthropic.TextDelta:
					ch <- Chunk{Content: deltaVariant.Text, Done: false}
				}
			case anthropic.MessageStopEvent:
				ch <- Chunk{Done: true}
				return
			}
		}

		if stream.Err() != nil {
			log.Printf("[Anthropic] Stream error: %v", stream.Err())
		}
	}()

	return ch, nil
}

func (a *Anthropic) StreamWithTools(ctx context.Context, messages []Message, cfg Config, tools []ToolDefinition, toolChoice string) (<-chan StreamChunk, error) {
	apiMessages := convertMessages(messages)
	apiTools := convertTools(tools)

	log.Printf("[Anthropic] StreamWithTools: messages=%d, tools=%d, baseURL=%s", len(messages), len(tools), a.baseURL)
	for i, msg := range messages {
		log.Printf("[Anthropic] Message %d: role=%s, ToolCallID=%s", i, msg.Role, msg.ToolCallID)
	}

	ch := make(chan StreamChunk, 100)

	go func() {
		defer close(ch)

		stream := a.client.Messages.NewStreaming(ctx, anthropic.MessageNewParams{
			Model:       anthropic.Model(cfg.Model),
			Messages:    apiMessages,
			MaxTokens:   int64(cfg.MaxTokens),
			Tools:       apiTools,
			Temperature: anthropic.Float(cfg.Temperature),
		})

		pendingToolCalls := make(map[int]ToolCall)
		var thinking strings.Builder

		for stream.Next() {
			event := stream.Current()

			switch ev := event.AsAny().(type) {
			case anthropic.ContentBlockStartEvent:
				switch block := ev.ContentBlock.AsAny().(type) {
				case anthropic.TextBlock:
					if block.Text != "" {
						ch <- StreamChunk{Content: block.Text, Done: false}
					}
				case anthropic.ToolUseBlock:
					log.Printf("[Anthropic] ToolUseBlock: ID=%s, Name=%s", block.ID, block.Name)
					tc := ToolCall{
						ID:   block.ID,
						Type: "function",
						Function: struct {
							Name      string `json:"name"`
							Arguments string `json:"arguments"`
						}{
							Name:      block.Name,
							Arguments: "",
						},
					}
					pendingToolCalls[int(ev.Index)] = tc
					log.Printf("[Anthropic] Stored tool call ID: %s", block.ID)
				case anthropic.ThinkingBlock:
					if block.Thinking != "" {
						thinking.WriteString(block.Thinking)
						ch <- StreamChunk{Thinking: block.Thinking, Done: false}
					}
				}

			case anthropic.ContentBlockDeltaEvent:
				switch delta := ev.Delta.AsAny().(type) {
				case anthropic.TextDelta:
					if delta.Text != "" {
						ch <- StreamChunk{Content: delta.Text, Done: false}
					}
				case anthropic.ThinkingDelta:
					if delta.Thinking != "" {
						thinking.WriteString(delta.Thinking)
						ch <- StreamChunk{Thinking: delta.Thinking, Done: false}
					}
				case anthropic.InputJSONDelta:
					if delta.PartialJSON != "" {
						chunkStr := parseASCIIArray([]byte(delta.PartialJSON))
						if tc, ok := pendingToolCalls[int(ev.Index)]; ok {
							tc.Function.Arguments += chunkStr
							pendingToolCalls[int(ev.Index)] = tc
						}
					}
				default:
					log.Printf("[Anthropic] Unknown delta type: %T", ev.Delta.AsAny())
				}
			}
		}

		if err := stream.Err(); err != nil {
			log.Printf("[Anthropic] Stream error: %v", err)
		}

		for i, tc := range pendingToolCalls {
			ch <- StreamChunk{
				ToolCalls:     []ToolCall{tc},
				ToolCallIndex: i,
				Done:          false,
			}
		}

		ch <- StreamChunk{Done: true}
	}()

	return ch, nil
}

func (a *Anthropic) Name() string {
	return "anthropic"
}

func convertMessages(messages []Message) []anthropic.MessageParam {
	result := make([]anthropic.MessageParam, 0, len(messages))
	i := 0
	for i < len(messages) {
		m := messages[i]
		switch m.Role {
		case "user":
			result = append(result, anthropic.NewUserMessage(anthropic.NewTextBlock(m.Content)))
			i++
		case "assistant":
			result = append(result, anthropic.NewAssistantMessage(anthropic.NewTextBlock(m.Content)))
			i++
		case "tool":
			var blocks []anthropic.ContentBlockParamUnion
			for i < len(messages) && messages[i].Role == "tool" {
				toolMsg := messages[i]
				toolUseID := strings.TrimPrefix(toolMsg.ToolCallID, "call_function_")
				log.Printf("[convertMessages] Tool result: originalID=%q, normalizedID=%q, contentLen=%d", toolMsg.ToolCallID, toolUseID, len(toolMsg.Content))
				toolResultBlock := anthropic.ContentBlockParamUnion{
					OfToolResult: &anthropic.ToolResultBlockParam{
						ToolUseID: toolUseID,
						Content: []anthropic.ToolResultBlockParamContentUnion{
							{OfText: &anthropic.TextBlockParam{Text: toolMsg.Content}},
						},
					},
				}
				blocks = append(blocks, toolResultBlock)
				i++
			}
			if len(blocks) > 0 {
				result = append(result, anthropic.NewUserMessage(blocks...))
			}
		default:
			i++
		}
	}
	return result
}

func convertTools(tools []ToolDefinition) []anthropic.ToolUnionParam {
	result := make([]anthropic.ToolUnionParam, len(tools))
	for i, t := range tools {
		name := getStringFromMap(t.Function, "name")
		description := getStringFromMap(t.Function, "description")

		var properties map[string]interface{}
		if params, ok := t.Function["parameters"].(map[string]interface{}); ok {
			if props, ok := params["properties"].(map[string]interface{}); ok {
				properties = props
			}
		}

		result[i] = anthropic.ToolUnionParam{
			OfTool: &anthropic.ToolParam{
				Name:        name,
				Description: anthropic.String(description),
				InputSchema: anthropic.ToolInputSchemaParam{
					Type:       "object",
					Properties: properties,
				},
			},
		}
	}
	return result
}

func getStringFromMap(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func parseASCIIArray(input []byte) string {
	if len(input) == 0 {
		return ""
	}

	var result []byte
	for _, b := range input {
		if b >= 32 && b <= 126 {
			result = append(result, b)
		} else if b >= 48 && b <= 57 {
			result = append(result, b)
		} else if b == 44 || b == 58 {
			result = append(result, b)
		}
	}
	return string(result)
}
