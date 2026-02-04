package provider

import (
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

type minimaxRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

type minimaxResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
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
			Role    string `json:"role"`
			Content string `json:"content"`
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
		url = "https://api.minimax.io/anthropic/v1/messages"
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

func (m *MiniMax) Name() string {
	return "minimax"
}
