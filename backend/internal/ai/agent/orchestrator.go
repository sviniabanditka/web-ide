package agent

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/webide/ide/backend/internal/ai/provider"
	"github.com/webide/ide/backend/internal/ai/tools"
)

type WebSocketSender func(event WSEvent) error

type AgentOrchestrator struct {
	toolRegistry *tools.ToolRegistry
	anthropic    *provider.Anthropic
	policy       *PolicyEngine
	mu           sync.RWMutex
	sessions     map[uuid.UUID]*AgentSession
}

func NewOrchestrator(registry *tools.ToolRegistry, anthropic *provider.Anthropic) *AgentOrchestrator {
	return &AgentOrchestrator{
		toolRegistry: registry,
		anthropic:    anthropic,
		policy:       NewPolicyEngine(),
		sessions:     make(map[uuid.UUID]*AgentSession),
	}
}

func (o *AgentOrchestrator) Run(ctx context.Context, session *AgentSession, userContent string, send WebSocketSender) error {
	session.AddUserMessage(userContent)

	if len(session.Messages) == 0 {
		systemPrompt := session.Config.SystemPrompt
		if systemPrompt == "" {
			systemPrompt = DefaultSystemPrompt
		}
		session.AddSystemMessage(systemPrompt)
	}

	o.mu.Lock()
	o.sessions[session.ID] = session
	o.mu.Unlock()

	defer func() {
		o.mu.Lock()
		delete(o.sessions, session.ID)
		o.mu.Unlock()
	}()

	step := 0
	maxSteps := session.Config.Limits.MaxSteps
	if maxSteps == 0 {
		maxSteps = DefaultConfig().Limits.MaxSteps
	}

	for step < maxSteps {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		toolDefs := o.toolRegistry.ListForModel()
		messages := o.sessionToProviderMessages(session)

		toolChoice := "auto"

		log.Printf("[Agent] Step %d/%d", step, maxSteps)
		log.Printf("[Agent] Messages count: %d", len(messages))
		log.Printf("[Agent] Tools available: %d", len(toolDefs))
		for i, td := range toolDefs {
			log.Printf("[Agent] Tool %d: %s", i, td.Function["name"])
		}

		providerTools := convertToProviderTools(toolDefs)
		stream, err := o.anthropic.StreamWithTools(ctx, messages, provider.Config{
			APIKey:      "",
			Model:       "minimax",
			MaxTokens:   4096,
			Temperature: 0.7,
		}, providerTools, toolChoice)

		if err != nil {
			log.Printf("[Agent] Stream error: %v", err)
			send(WSEvent{
				Type:      EventAgentError,
				SessionID: session.ID.String(),
				ProjectID: session.ProjectID.String(),
				Payload: AgentErrorPayload{
					Code:    "STREAM_ERROR",
					Message: err.Error(),
				},
			})
			return err
		}

		assistantText := ""
		var toolCalls []provider.ToolCall

		for chunk := range stream {
			if chunk.Content != "" {
				assistantText += chunk.Content
				send(WSEvent{
					Type:      "assistant.delta",
					SessionID: session.ID.String(),
					ProjectID: session.ProjectID.String(),
					Payload: map[string]interface{}{
						"content": chunk.Content,
						"done":    false,
					},
				})
			}

			for _, tc := range chunk.ToolCalls {
				toolCalls = append(toolCalls, tc)
			}

			if chunk.Done {
				break
			}
		}

		if len(toolCalls) == 0 && assistantText != "" {
			session.AddAssistantMessage(assistantText, nil)
			send(WSEvent{
				Type:      "assistant.final",
				SessionID: session.ID.String(),
				ProjectID: session.ProjectID.String(),
				Payload: map[string]interface{}{
					"content": assistantText,
				},
			})
			return nil
		}

		if len(toolCalls) > 0 {
			var agentToolCalls []ToolCall
			for _, tc := range toolCalls {
				agentToolCalls = append(agentToolCalls, ToolCall{
					ID:   tc.ID,
					Type: tc.Type,
					Function: ToolCallFunction{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				})
			}
			session.AddAssistantMessage(assistantText, agentToolCalls)

			for _, tc := range toolCalls {
				_, ok := o.toolRegistry.Get(tc.Function.Name)
				if !ok {
					send(WSEvent{
						Type:      EventToolError,
						SessionID: session.ID.String(),
						ProjectID: session.ProjectID.String(),
						ID:        tc.ID,
						Payload: map[string]interface{}{
							"error": map[string]interface{}{
								"code":    "UNKNOWN_TOOL",
								"message": "Unknown tool: " + tc.Function.Name,
							},
						},
					})

					result := tools.ToolResult{
						OK: false,
						Error: &tools.ToolError{
							Code:    tools.ErrCodeNotFound,
							Message: "Tool not found: " + tc.Function.Name,
						},
					}
					session.AddToolResult(tc.ID, tc.Function.Name, formatToolResult(result))

					continue
				}

				var args map[string]interface{}
				json.Unmarshal([]byte(tc.Function.Arguments), &args)

				decision := o.policy.Decide(tc.Function.Name, session, args)

				summary := GenerateToolSummary(tc.Function.Name, args)

				switch decision {
				case DecisionAllow:
					result := o.executeTool(session, tc.Function.Name, args)
					send(WSEvent{
						Type:      EventToolResult,
						SessionID: session.ID.String(),
						ProjectID: session.ProjectID.String(),
						ID:        tc.ID,
						Payload: map[string]interface{}{
							"id":       tc.ID,
							"name":     tc.Function.Name,
							"ok":       result.OK,
							"result":   result.Data,
							"error":    result.Error,
							"duration": result.Meta.DurationMs,
						},
					})
					session.AddToolResult(tc.ID, tc.Function.Name, formatToolResult(result))

				case DecisionConfirm:
					send(WSEvent{
						Type:      EventToolApprovalRequired,
						SessionID: session.ID.String(),
						ProjectID: session.ProjectID.String(),
						ID:        tc.ID,
						Payload: map[string]interface{}{
							"id":        tc.ID,
							"name":      tc.Function.Name,
							"arguments": args,
							"summary":   summary,
							"policy":    "confirm",
						},
					})

					session.SetPendingToolCall(tc.ID, &PendingToolCall{
						ToolCall: ToolCall{
							ID:   tc.ID,
							Type: tc.Type,
							Function: ToolCallFunction{
								Name:      tc.Function.Name,
								Arguments: tc.Function.Arguments,
							},
						},
						Args:      args,
						CreatedAt: time.Now(),
					})

					return nil

				case DecisionDeny:
					result := tools.ToolResult{
						OK: false,
						Error: &tools.ToolError{
							Code:    tools.ErrCodePermission,
							Message: "Tool blocked by policy",
						},
					}
					send(WSEvent{
						Type:      EventToolResult,
						SessionID: session.ID.String(),
						ProjectID: session.ProjectID.String(),
						ID:        tc.ID,
						Payload: map[string]interface{}{
							"id":    tc.ID,
							"name":  tc.Function.Name,
							"ok":    false,
							"error": result.Error,
						},
					})
					session.AddToolResult(tc.ID, tc.Function.Name, formatToolResult(result))
				}
			}
		}

		step++
	}

	send(WSEvent{
		Type:      EventAgentDone,
		SessionID: session.ID.String(),
		ProjectID: session.ProjectID.String(),
		Payload: AgentDonePayload{
			Steps:    step,
			FinalMsg: "Agent stopped: maximum steps reached",
		},
	})

	return nil
}

func (o *AgentOrchestrator) HandleApproval(ctx context.Context, sessionID uuid.UUID, toolCallID string, approved bool, reason string, send WebSocketSender) error {
	o.mu.RLock()
	session, ok := o.sessions[sessionID]
	o.mu.RUnlock()

	if !ok {
		return nil
	}

	pending, found := session.GetPendingToolCall(toolCallID)
	if !found {
		return nil
	}

	var result tools.ToolResult
	if approved {
		result = o.executeTool(session, pending.ToolCall.Function.Name, pending.Args)
	} else {
		result = tools.ToolResult{
			OK: false,
			Error: &tools.ToolError{
				Code:    tools.ErrCodeUserRejected,
				Message: "User rejected: " + reason,
			},
		}
	}

	session.RemovePendingToolCall(toolCallID)

	send(WSEvent{
		Type:      EventToolResult,
		SessionID: session.ID.String(),
		ProjectID: session.ProjectID.String(),
		ID:        toolCallID,
		Payload: map[string]interface{}{
			"id":     toolCallID,
			"name":   pending.ToolCall.Function.Name,
			"ok":     result.OK,
			"result": result.Data,
			"error":  result.Error,
		},
	})

	session.AddToolResult(toolCallID, pending.ToolCall.Function.Name, formatToolResult(result))

	go func() {
		runCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		o.Run(runCtx, session, "", send)
	}()

	return nil
}

func (o *AgentOrchestrator) executeTool(session *AgentSession, toolName string, args map[string]interface{}) tools.ToolResult {
	start := time.Now()

	tool, ok := o.toolRegistry.Get(toolName)
	if !ok {
		return tools.ToolResult{
			OK: false,
			Error: &tools.ToolError{
				Code:    tools.ErrCodeNotFound,
				Message: "Tool not found: " + toolName,
			},
		}
	}

	tc := tools.ToolContext{
		SessionID:   session.ID,
		ProjectID:   session.ProjectID,
		UserID:      session.UserID,
		ProjectRoot: session.Config.ProjectRoot,
		Mode:        string(session.Mode),
		Limits: tools.ToolLimits{
			MaxFileBytes:     session.Config.Limits.MaxFileBytes,
			MaxOutputBytes:   session.Config.Limits.MaxOutputBytes,
			MaxSearchResults: session.Config.Limits.MaxSearchResults,
			MaxPatchFiles:    session.Config.Limits.MaxPatchFiles,
			MaxToolTime:      session.Config.Limits.MaxToolTimeMs,
		},
	}

	result, err := tool.Execute(context.Background(), args, tc)

	if err != nil {
		return tools.ToolResult{
			OK: false,
			Error: &tools.ToolError{
				Code:    tools.ErrCodeExecution,
				Message: err.Error(),
			},
		}
	}

	if result.Meta == nil {
		result.Meta = &tools.ResultMeta{}
	}
	result.Meta.DurationMs = time.Since(start).Milliseconds()

	return result
}

func (o *AgentOrchestrator) sessionToProviderMessages(session *AgentSession) []provider.Message {
	var msgs []provider.Message
	for _, msg := range session.GetMessages() {
		m := provider.Message{
			Role:       string(msg.Role),
			Content:    msg.Content,
			ToolCallID: msg.ToolCallID,
		}
		if len(msg.ToolCalls) > 0 {
			var tcJSON []map[string]interface{}
			for _, tc := range msg.ToolCalls {
				tcJSON = append(tcJSON, map[string]interface{}{
					"id":   tc.ID,
					"type": tc.Type,
					"function": map[string]interface{}{
						"name":      tc.Function.Name,
						"arguments": tc.Function.Arguments,
					},
				})
			}
			b, _ := json.Marshal(tcJSON)
			m.Content = string(b)
		}
		msgs = append(msgs, m)
	}
	return msgs
}

func convertToProviderTools(toolDefs []tools.ToolDefinition) []provider.ToolDefinition {
	result := make([]provider.ToolDefinition, len(toolDefs))
	for i, td := range toolDefs {
		result[i] = provider.ToolDefinition{
			Type:     td.Type,
			Function: td.Function,
		}
	}
	return result
}

func formatToolResult(result tools.ToolResult) string {
	data, _ := json.Marshal(result)
	return string(data)
}

const DefaultSystemPrompt = `You are an AI assistant inside a WebIDE.

### IMPORTANT: You MUST always specify arguments for tools!
If you call a tool without arguments like {"name": "list_dir"} or {"name": "apply_patch" with empty arguments {}, THE TOOL WILL FAIL!

Examples of CORRECT tool calls:
- {"name": "list_dir", "arguments": {"path": ".", "depth": 1}}
- {"name": "read_file", "arguments": {"path": "main.go", "start_line": 1, "end_line": 50}}
- {"name": "search_in_files", "arguments": {"query": "func main", "max_results": 20}}
- {"name": "apply_patch", "arguments": {"patch": "--- /dev/null\n+++ hello.go\n@@ -0,0 +1,5 @@\n+package main\n+", "dry_run": true}}
- {"name": "run_command", "arguments": {"cmd": "ls -la", "timeout_ms": 60000}}

NEVER call a tool without all required arguments!

### Core rules
1. DO use tools to interact with files and run commands.
2. When asked to create a file, use apply_patch with a unified diff OR run_command with 'cat > file'.
3. When asked to modify a file, use apply_patch with a unified diff.
4. Once a file is created, STOP - do not call list_dir again or try to create the same file!
5. After successful tool execution, check the RESULT. If the file was created successfully, respond to the user with confirmation.
6. Never tell users you "cannot" do something - use the tools available to you.

### Tool descriptions
- list_dir: List directory contents. Required args: path, depth
- read_file: Read file contents. Required args: path, optional: start_line, end_line
- search_in_files: Search for text patterns. Required args: query, optional: max_results
- apply_patch: Create or modify files using unified diffs. Required args: patch, optional: dry_run
- run_command: Execute shell commands. Required args: cmd, optional: timeout_ms

### How to create a NEW file
Choose ONE method:
- Use apply_patch with {"patch": "--- /dev/null\n+++ filename.go\n@@ -0,0 +1,3 @@\n+package main\n+"} and {"dry_run": false}
- OR use run_command: {"cmd": "cat > filename.go << 'EOF'\npackage main\nEOF"}

### How to modify an EXISTING file
Use apply_patch with {"patch": "diff content"} showing the changes.

### Workflow for creating files
1. Call list_dir ONCE to check current structure.
2. Create the file with ONE tool call (apply_patch OR run_command).
3. Check the tool result - if successful, the file exists!
4. STOP - do NOT call list_dir again or try to create the file again!
5. Respond to the user with confirmation that the file was created.

### CRITICAL: Avoid repetition!
- Call list_dir only ONCE at the beginning
- Do NOT verify the same file multiple times
- If tool result shows "ok": true, the file WAS created - trust it!
- If tool result shows "exit_code": 0, the command SUCCEEDED
`
