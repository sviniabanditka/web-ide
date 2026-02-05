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

	ch := make(chan StreamChunk, 100)

	go func() {
		defer close(ch)

		msg, err := a.client.Messages.New(ctx, anthropic.MessageNewParams{
			Model:       anthropic.Model(cfg.Model),
			Messages:    apiMessages,
			MaxTokens:   int64(cfg.MaxTokens),
			Tools:       apiTools,
			Temperature: anthropic.Float(cfg.Temperature),
		})

		if err != nil {
			log.Printf("[Anthropic] Non-streaming error: %v", err)
			return
		}

		var content strings.Builder
		var toolCalls []ToolCall

		for _, block := range msg.Content {
			switch b := block.AsAny().(type) {
			case anthropic.TextBlock:
				content.WriteString(b.Text)

			case anthropic.ToolUseBlock:
				inputStr := string(b.Input)
				if len(b.Input) > 0 && b.Input[0] == '[' {
					inputStr = parseASCIIArray(b.Input)
				}

				tc := ToolCall{
					ID:   b.ID,
					Type: "function",
					Function: struct {
						Name      string `json:"name"`
						Arguments string `json:"arguments"`
					}{
						Name:      b.Name,
						Arguments: inputStr,
					},
				}
				toolCalls = append(toolCalls, tc)
			}
		}

		if content.Len() > 0 {
			for _, r := range content.String() {
				ch <- StreamChunk{Content: string(r), Done: false}
			}
		}

		for _, tc := range toolCalls {
			ch <- StreamChunk{
				ToolCalls: []ToolCall{tc},
				Done:      false,
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
	for _, m := range messages {
		result = append(result, anthropic.NewUserMessage(anthropic.NewTextBlock(m.Content)))
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
