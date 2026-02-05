package provider

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type MiniMax struct {
	client *http.Client
}

func NewMiniMax() *MiniMax {
	return &MiniMax{
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
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

type minimaxRequest struct {
	Model       string           `json:"model"`
	Messages    []Message        `json:"messages"`
	MaxTokens   int              `json:"max_tokens,omitempty"`
	Temperature float64          `json:"temperature,omitempty"`
	Stream      bool             `json:"stream,omitempty"`
	Tools       []ToolDefinition `json:"tools,omitempty"`
	ToolChoice  string           `json:"tool_choice,omitempty"`
}

type minimaxResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role      string     `json:"role"`
			Content   string     `json:"content"`
			ToolCalls []ToolCall `json:"tool_calls,omitempty"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type streamChunk struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Role      string     `json:"role"`
			Content   string     `json:"content"`
			ToolCalls []ToolCall `json:"tool_calls,omitempty"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

func (m *MiniMax) Complete(ctx context.Context, messages []Message, cfg Config) (*Response, error) {
	url := cfg.URL
	if url == "" {
		url = "https://api.minimax.chat/v1/text/chatcompletion_v2"
	}

	reqBody := minimaxRequest{
		Model:       cfg.Model,
		Messages:    messages,
		MaxTokens:   cfg.MaxTokens,
		Temperature: cfg.Temperature,
		Stream:      false,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("api error: %s", string(respBody))
	}

	var result minimaxResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	choice := result.Choices[0]
	return &Response{
		Content: choice.Message.Content,
		Usage: TokenUsage{
			PromptTokens:     result.Usage.PromptTokens,
			CompletionTokens: result.Usage.CompletionTokens,
			TotalTokens:      result.Usage.TotalTokens,
		},
		StopReason: choice.FinishReason,
	}, nil
}

func (m *MiniMax) Stream(ctx context.Context, messages []Message, cfg Config) (<-chan Chunk, error) {
	url := cfg.URL
	if url == "" {
		url = "https://api.minimax.chat/v1/text/chatcompletion_v2"
	}

	type anthropicRequest struct {
		Model       string    `json:"model"`
		Messages    []Message `json:"messages"`
		MaxTokens   int       `json:"max_tokens_to_sample"`
		Temperature float64   `json:"temperature,omitempty"`
		Stream      bool      `json:"stream"`
	}

	reqBody := anthropicRequest{
		Model:       cfg.Model,
		Messages:    messages,
		MaxTokens:   cfg.MaxTokens,
		Temperature: cfg.Temperature,
		Stream:      false,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	log.Printf("[MiniMax] URL: %s", url)
	log.Printf("[MiniMax] Request: %s", string(body))

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("[MiniMax] Status: %d", resp.StatusCode)

	if resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("api error: %s", string(respBody))
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	log.Printf("[MiniMax] Response body: %s", string(respBody))

	type minimaxContentBlock struct {
		Type     string `json:"type"`
		Text     string `json:"text,omitempty"`
		Thinking string `json:"thinking,omitempty"`
	}

	type minimaxResponse struct {
		Content []minimaxContentBlock `json:"content"`
	}

	var respStruct struct {
		Content []minimaxContentBlock `json:"content"`
	}

	if err := json.Unmarshal(respBody, &respStruct); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(respStruct.Content) == 0 {
		return nil, fmt.Errorf("no content in response")
	}

	var content strings.Builder
	for _, block := range respStruct.Content {
		if block.Type == "text" && block.Text != "" {
			content.WriteString(block.Text)
		}
	}

	fullContent := content.String()
	log.Printf("[MiniMax] Full content: %s", fullContent)

	ch := make(chan Chunk, 100)

	go func() {
		defer close(ch)

		for _, r := range fullContent {
			ch <- Chunk{Content: string(r), Done: false}
			time.Sleep(10 * time.Millisecond)
		}

		ch <- Chunk{Done: true}
	}()

	return ch, nil
}

func (m *MiniMax) StreamWithTools(ctx context.Context, messages []Message, cfg Config, tools []ToolDefinition, toolChoice string) (<-chan StreamChunk, error) {
	url := cfg.URL
	if url == "" {
		url = "https://api.minimax.chat/v1/text/chatcompletion_v2"
	}

	reqBody := minimaxRequest{
		Model:       cfg.Model,
		Messages:    messages,
		MaxTokens:   cfg.MaxTokens,
		Temperature: cfg.Temperature,
		Stream:      true,
		Tools:       tools,
		ToolChoice:  toolChoice,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	log.Printf("[MiniMax] StreamWithTools URL: %s", url)
	log.Printf("[MiniMax] StreamWithTools Request: %s", string(body))
	log.Printf("[MiniMax] ToolChoice: %s", toolChoice)
	log.Printf("[MiniMax] Tools count: %d", len(tools))

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("api error: %s", string(respBody))
	}

	respBody, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	log.Printf("[MiniMax] Response: %s", string(respBody))

	ch := make(chan StreamChunk, 100)

	go func() {
		defer close(ch)

		var respStruct struct {
			Choices []struct {
				Message struct {
					Content   string     `json:"content"`
					ToolCalls []ToolCall `json:"tool_calls"`
				} `json:"message"`
				FinishReason string `json:"finish_reason"`
			} `json:"choices"`
		}

		if err := json.Unmarshal(respBody, &respStruct); err != nil {
			log.Printf("[MiniMax] Failed to parse response: %v", err)
			return
		}

		if len(respStruct.Choices) == 0 {
			log.Printf("[MiniMax] No choices in response")
			ch <- StreamChunk{Done: true}
			return
		}

		choice := respStruct.Choices[0]

		for _, tc := range choice.Message.ToolCalls {
			log.Printf("[MiniMax] Sending tool call: %s", tc.Function.Name)
			ch <- StreamChunk{
				ToolCalls: []ToolCall{tc},
			}
		}

		if choice.Message.Content != "" {
			for _, r := range choice.Message.Content {
				ch <- StreamChunk{Content: string(r)}
				time.Sleep(10 * time.Millisecond)
			}
		}

		log.Printf("[MiniMax] Stream done, finish_reason: %s", choice.FinishReason)
		ch <- StreamChunk{Done: true}
	}()

	return ch, nil
}

func (m *MiniMax) parseStreamChunk(line string) *StreamChunk {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}

	if strings.HasPrefix(line, "data: ") {
		line = strings.TrimPrefix(line, "data: ")
	}

	var sc streamChunk
	if err := json.Unmarshal([]byte(line), &sc); err != nil {
		log.Printf("[MiniMax] Failed to parse chunk: %v", err)
		return nil
	}

	result := &StreamChunk{}

	if len(sc.Choices) > 0 {
		choice := sc.Choices[0]
		result.Content = choice.Delta.Content
		result.ToolCalls = choice.Delta.ToolCalls

		if choice.FinishReason != "" && choice.FinishReason != "null" {
			result.Done = true
		}
	}

	return result
}

func (m *MiniMax) Name() string {
	return "minimax"
}

type StreamReader struct {
	scanner *bufio.Scanner
}

func NewStreamReader(r io.Reader) *StreamReader {
	return &StreamReader{
		scanner: bufio.NewScanner(r),
	}
}

func (sr *StreamReader) ReadLine() (string, error) {
	if !sr.scanner.Scan() {
		if err := sr.scanner.Err(); err != nil {
			return "", err
		}
		return "", io.EOF
	}
	return sr.scanner.Text(), nil
}
