package provider

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
		url = "https://api.minimax.chat/v1/text/chatcompletion_v2"
	}

	reqBody := minimaxRequest{
		Model:       cfg.Model,
		Messages:    messages,
		MaxTokens:   cfg.MaxTokens,
		Temperature: cfg.Temperature,
		Stream:      true,
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

	ch := make(chan Chunk, 100)

	go func() {
		defer close(ch)
		defer resp.Body.Close()

		reader := bufio.NewReader(resp.Body)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				lineBytes, err := reader.ReadBytes('\n')
				if err != nil {
					if err != io.EOF {
						ch <- Chunk{Done: true}
					}
					return
				}

				lineBytes = bytes.TrimSpace(lineBytes)
				if len(lineBytes) == 0 {
					continue
				}

				if !bytes.HasPrefix(lineBytes, []byte("data:")) {
					continue
				}

				data := bytes.TrimSpace(lineBytes[5:])
				if string(data) == "[DONE]" {
					ch <- Chunk{Done: true}
					return
				}

				var chunk streamChunk
				if err := json.Unmarshal(data, &chunk); err != nil {
					continue
				}

				if len(chunk.Choices) > 0 {
					content := chunk.Choices[0].Delta.Content
					if content != "" {
						ch <- Chunk{Content: content, Done: false}
					}
					if chunk.Choices[0].FinishReason != "" {
						ch <- Chunk{Done: true}
						return
					}
				}
			}
		}
	}()

	return ch, nil
}

func (m *MiniMax) Name() string {
	return "minimax"
}
