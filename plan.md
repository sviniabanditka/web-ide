Based on my exploration, I now have a clear understanding of the codebase. Let me create a comprehensive implementation plan for the tools layer.

## Implementation Plan: Tools Layer for WebIDE

### Executive Summary

The plan adds a tools layer over the existing WebSocket chat by:
1. Creating an `agent` package with `AgentOrchestrator` that wraps the existing chat streaming
2. Extending the MiniMax provider to support function calling
3. Implementing a secure `tool` registry with path validation
4. Adding new WS event types without breaking existing `assistant.delta` flow
5. Updating Vue frontend to display tool events with approval UI

---

## Phase 1: Backend Core Infrastructure

### 1.1 Create Agent Package Structure

**New Directory**: `backend/internal/ai/agent/`

```
agent/
├── agent.go              # AgentSession, AgentOrchestrator main logic
├── messages.go           # ModelMessage types with tool roles
├── config.go             # Session config, limits, modes
├── loop.go               # Main agent iteration loop
├── policy.go             # Tool approval policies
└── events.go             # WS event types for tools
```

**Key Types to Add**:

```go
// agent/agent.go
type AgentSession struct {
    ID            uuid.UUID
    ProjectID     uuid.UUID
    UserID        uuid.UUID
    Messages      []ModelMessage
    Mode          AgentMode  // safe, write, exec
    PendingCalls  map[string]*PendingToolCall
    RunningCmds   map[string]*CommandProcess
    Config        AgentConfig
}

type AgentMode string

const (
    ModeSafe  AgentMode = "safe"   // Read-only unless explicit approval
    ModeWrite AgentMode = "write"  // Propose patches, wait for approval
    ModeExec  AgentMode = "exec"   // Allow commands with confirmation
)

// agent/messages.go
type ModelMessage struct {
    Role    string          // "system", "user", "assistant", "tool"
    Content string
    ToolCalls []ToolCall    // For assistant messages
    ToolCallID string       // For tool responses
}

// agent/events.go
type ToolEvent struct {
    SessionID string    `json:"sessionId"`
    ProjectID string    `json:"projectId"`
    Type      string    `json:"type"` // "tool.call", "tool.approval_required", "tool.result", "tool.error"
    TS        time.Time `json:"ts"`
    ID        string    `json:"id,omitempty"`
    Payload   any       `json:"payload"`
}
```

### 1.2 Extend MiniMax Provider for Function Calling

**File to Modify**: `backend/internal/ai/provider/minimax.go`

**Required Changes**:

```go
// Add to existing MinimaxProvider struct
type ToolDefinition struct {
    Type     string                 `json:"type"`
    Function map[string]interface{} `json:"function"`
}

type ToolCall struct {
    ID       string                 `json:"id"`
    Type     string                 `json:"type"`
    Function struct {
        Name      string                 `json:"name"`
        Arguments string                 `json:"arguments"`
    } `json:"function"`
}

type StreamChunk struct {
    Content    string      `json:"content,omitempty"`
    Done       bool        `json:"done"`
    ToolCalls  []ToolCall  `json:"tool_calls,omitempty"`
    ToolCallID string      `json:"tool_call_id,omitempty"` // For tool results
}

// New method: StreamWithTools
func (p *MinimaxProvider) StreamWithTools(
    ctx context.Context, 
    messages []provider.Message, 
    tools []ToolDefinition,
    toolChoice string, // "auto" or specific
) (<-chan StreamChunk, error) {
    // Build request body with tools
    // Implement tool_choice:auto
    // Parse streaming response for both text chunks and tool_calls
    // Return unified channel with both types
}
```

**MiniMax API Compatibility**:
- Check if MiniMax API supports OpenAI-style function calling
- If not, adapt to their specific tool call format
- Handle streaming of partial tool calls (may need buffering)

### 1.3 Add Tool Registry Package

**New Directory**: `backend/internal/ai/tools/`

```
tools/
├── registry.go        # ToolRegistry, tool definitions
├── executor.go        # Tool execution logic
├── schema.go          # JSONSchema validation
├── errors.go          # Standardized error formats
└── builtin/
    ├── list_dir.go
    ├── read_file.go
    ├── search.go
    ├── apply_patch.go
    └── run_command.go
```

**Core Implementation**:

```go
// tools/registry.go
type Tool struct {
    Name        string
    Description string
    Parameters  map[string]interface{}  // JSONSchema
    Policy      ToolPolicy
    Execute     func(ctx context.Context, args map[string]interface{}, ctx2 ToolContext) (any, error)
}

type ToolPolicy string

const (
    PolicyAllow    ToolPolicy = "allow"
    PolicyConfirm  ToolPolicy = "confirm"
    PolicyDeny     ToolPolicy = "deny"
)

type ToolContext struct {
    SessionID  uuid.UUID
    ProjectID  uuid.UUID
    UserID     uuid.UUID
    Mode       agent.AgentMode
    Limits     agent.Limits
}

type ToolRegistry struct {
    mu    sync.RWMutex
    tools map[string]Tool
}

func (r *ToolRegistry) Register(t Tool) error
func (r *ToolRegistry) Get(name string) (Tool, bool)
func (r *ToolRegistry) List() []Tool
func (r *ToolRegistry) ListForModel() []map[string]interface{}  // OpenAI-style
func (r *ToolRegistry) ValidateArgs(name string, args map[string]interface{}) error
```

**Standardized Result Format**:

```go
// tools/errors.go
type ToolResult struct {
    OK    bool        `json:"ok"`
    Data  any         `json:"data,omitempty"`
    Meta  ResultMeta  `json:"meta,omitempty"`
    Error *ToolError  `json:"error,omitempty"`
}

type ResultMeta struct {
    DurationMs int64  `json:"duration_ms"`
    Truncated  bool   `json:"truncated,omitempty"`
}

type ToolError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details any    `json:"details,omitempty"`
}

// Error codes
const (
    ErrCodeValidation    = "VALIDATION_ERROR"
    ErrCodeNotFound      = "FILE_NOT_FOUND"
    ErrCodePermission    = "PERMISSION_DENIED"
    ErrCodeTimeout      = "TOOL_TIMEOUT"
    ErrCodeSizeLimit    = "SIZE_LIMIT_EXCEEDED"
    ErrCodeUserRejected = "USER_REJECTED"
)
```

### 1.4 Project Guard (Path Security)

**New File**: `backend/internal/ai/tools/path_guard.go`

```go
type PathGuard struct {
    ProjectRoot string
    MaxFileBytes int64
    Limits       Limits
}

func (g *PathGuard) ResolveProjectPath(userPath string) (string, error) {
    // Reject paths with ..
    // Resolve symlinks
    // Verify result is within ProjectRoot
    // Reject /proc, /sys, /dev access
    // Check MaxFileBytes
    // Normalize slashes
}

func (g *PathGuard) ValidateFileAccess(absPath string) error
func (g *PathGuard) ValidateDirAccess(absPath string) error
```

### 1.5 Tool Implementations (MVP)

#### 1.5.1 list_dir

```go
// tools/builtin/list_dir.go
type ListDirArgs struct {
    Path          string `json:"path"`
    Depth         int    `json:"depth"`
    IncludeHidden bool   `json:"include_hidden"`
}

type ListDirResult struct {
    Path    string      `json:"path"`
    Entries []DirEntry  `json:"entries"`
}

type DirEntry struct {
    Name    string `json:"name"`
    Type    string `json:"type"` // "file", "dir", "symlink"
    Size    int64  `json:"size"`
    MTime   int64  `json:"mtime"`
}
```

#### 1.5.2 read_file

```go
// tools/builtin/read_file.go
type ReadFileArgs struct {
    Path       string `json:"path"`
    MaxBytes   int    `json:"max_bytes"`
    StartLine  *int   `json:"start_line,omitempty"`
    EndLine    *int   `json:"end_line,omitempty"`
}

type ReadFileResult struct {
    Path      string `json:"path"`
    SHA       string `json:"sha"`
    Content   string `json:"content"`
    Truncated bool   `json:"truncated"`
    LineStart int    `json:"line_start,omitempty"`
    LineEnd   int    `json:"line_end,omitempty"`
}
```

#### 1.5.3 search_in_files

```go
// tools/builtin/search.go
type SearchArgs struct {
    Query      string   `json:"query"`
    Globs      []string `json:"globs,omitempty"`
    MaxResults int      `json:"max_results"`
}

type SearchResult struct {
    Query     string    `json:"query"`
    Matches   []Match   `json:"matches"`
    Truncated bool      `json:"truncated"`
}

type Match struct {
    Path    string `json:"path"`
    Line    int    `json:"line"`
    Col     int    `json:"col"`
    Preview string `json:"preview"`
}
```

#### 1.5.4 apply_patch

```go
// tools/builtin/apply_patch.go
type ApplyPatchArgs struct {
    Patch   string `json:"patch"`
    DryRun  bool   `json:"dry_run"`
}

type ApplyPatchResult struct {
    Applied     bool            `json:"applied"`
    Files       []FileChange    `json:"files,omitempty"`
    Rejects     []Reject        `json:"rejects,omitempty"`
    PreviewSummary string       `json:"preview_summary,omitempty"`
}

type FileChange struct {
    Path       string `json:"path"`
    SHABefore  string `json:"sha_before"`
    SHAAfter   string `json:"sha_after"`
}

type Reject struct {
    Path   string `json:"path"`
    Reason string `json:"reason"`
    Hunk   string `json:"hunk"`
}
```

#### 1.5.5 run_command + streaming

```go
// tools/builtin/run_command.go
type RunCommandArgs struct {
    Cmd      string            `json:"cmd"`
    CWD      string            `json:"cwd"`
    Timeout  int               `json:"timeout_ms"`
    Env      map[string]string `json:"env,omitempty"`
    Stream   bool              `json:"stream"`
}

type CommandProcess struct {
    Handle     string
    Cmd        *exec.Cmd
    StartedAt  time.Time
    OutputBuf  *ring.Buffer
    Done       bool
    ExitCode   int
    mu         sync.Mutex
}

func (p *CommandProcess) AppendOutput(stream string, data []byte)
func (p *CommandProcess) GetOutput(from, limit int) ([]OutputLine, bool, int)
func (p *CommandProcess) Cancel() error

// Additional tools
func GetCommandOutput(ctx context.Context, handle string, from, limit int) (*CommandOutputResult, error)
func CancelCommand(ctx context.Context, handle string) (*CancelResult, error)
```

---

## Phase 2: WebSocket Protocol Extensions

### 2.1 New Event Types

**File to Modify**: `backend/internal/ai/chat_ws.go`

**Add to ChatWSMessage struct**:

```go
// Extended server→client events
type ServerEvent struct {
    Type      string          `json:"type"`
    SessionID string          `json:"sessionId"`
    ProjectID string          `json:"projectId"`
    TS        time.Time       `json:"ts"`
    ID        string          `json:"id,omitempty"`
    Payload   json.RawMessage  `json:"payload"`
}

// New event types
const (
    EventToolCall             = "tool.call"
    EventToolApprovalRequired = "tool.approval_required"
    EventToolResult           = "tool.result"
    EventToolError            = "tool.error"
    EventCommandOutput        = "command.output"
    EventCommandDone          = "command.done"
)

// Payloads
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
    Policy     string                 `json:"policy"` // "confirm" or "deny"
}

type ToolResultPayload struct {
    ToolCallID string      `json:"id"`
    Name       string      `json:"name"`
    OK         bool        `json:"ok"`
    Result     any         `json:"result,omitempty"`
    Error      *ToolError  `json:"error,omitempty"`
}

type CommandOutputPayload struct {
    Handle  string `json:"handle"`
    Stream  string `json:"stream"` // "stdout" | "stderr"
    Text    string `json:"text"`
    TS      int64  `json:"ts"`
}

type CommandDonePayload struct {
    Handle   string `json:"handle"`
    ExitCode int    `json:"exit_code"`
    Duration int64  `json:"duration_ms"`
}
```

**Client→server events**:

```go
// New client events
const (
    EventToolApprove = "tool.approve"
    EventToolReject  = "tool.reject"
)

type ToolApprovePayload struct {
    ToolCallID string `json:"tool_call_id"`
}

type ToolRejectPayload struct {
    ToolCallID string `json:"tool_call_id"`
    Reason     string `json:"reason,omitempty"`
}
```

### 2.2 Modify Message Handling

**In `handleSendMessage()` function**:

```go
func (hub *ChatWSHub) handleSendMessage(client *ChatWSClient, payload []byte) error {
    var msg struct {
        Content   string                 `json:"content"`
        Mode      string                 `json:"mode,omitempty"` // safe, write, exec
        StreamTool bool                  `json:"stream_tool,omitempty"`
    }
    json.Unmarshal(payload, &msg)
    
    // Instead of direct streaming, create agent session
    session := agent.NewAgentSession(...)
    
    // Run agent orchestrator
    return hub.runAgentLoop(client, session, msg.Content)
}
```

**New function: `runAgentLoop`**:

```go
func (hub *ChatWSHub) runAgentLoop(client *ChatWSClient, session *agent.AgentSession, userMessage string) error {
    // 1. Add user message to history
    session.AddMessage("user", userMessage)
    
    // 2. Get tools from registry
    tools := toolRegistry.ListForModel()
    
    // 3. Stream loop
    for step := 0; step < session.Config.MaxSteps; step++ {
        // Call MiniMax with tools
        stream, err := minimaxProvider.StreamWithTools(
            ctx, 
            session.ToProviderMessages(), 
            tools,
            "auto",
        )
        
        for chunk := range stream {
            // Handle text chunks (existing behavior)
            if chunk.Content != "" {
                hub.sendToClient(client, makeChunkEvent(chunk.Content, false))
                session.AppendAssistantContent(chunk.Content)
            }
            
            // Handle tool calls (new)
            for _, tc := range chunk.ToolCalls {
                toolCallID := generateToolCallID()
                
                // Get tool from registry
                tool, ok := toolRegistry.Get(tc.Function.Name)
                if !ok {
                    sendToolError(client, toolCallID, "UNKNOWN_TOOL", "Unknown tool: "+tc.Function.Name)
                    continue
                }
                
                // Parse arguments
                var args map[string]interface{}
                json.Unmarshal([]byte(tc.Function.Arguments), &args)
                
                // Apply policy
                decision := applyToolPolicy(tool, session.Mode, args)
                
                switch decision {
                case PolicyAllow:
                    // Execute immediately
                    result := executeToolSync(tool, args, session)
                    sendToolResult(client, toolCallID, tool.Name, result)
                    session.AddToolResult(toolCallID, tool.Name, result)
                    
                case PolicyConfirm:
                    // Send approval required, pause loop
                    sendApprovalRequired(client, toolCallID, tool.Name, args, tool.Policy)
                    session.PendingCalls[toolCallID] = &PendingToolCall{
                        Tool:      tool,
                        Args:      args,
                        CreatedAt: time.Now(),
                    }
                    return nil // Pause, wait for client response
                    
                case PolicyDeny:
                    result := makeErrorResult(ErrCodePermission, "Tool blocked by policy")
                    sendToolResult(client, toolCallID, tool.Name, result)
                    session.AddToolResult(toolCallID, tool.Name, result)
                }
            }
        }
        
        // Check for final message without tool calls
        if !stream.HasMore() {
            break
        }
    }
    
    // Send final message
    finalContent := session.GetFinalAssistantMessage()
    hub.sendToClient(client, makeFinalEvent(finalContent))
    
    // Save to database
    saveChatMessage(...)
    
    return nil
}
```

---

## Phase 3: Agent Orchestrator Implementation

### 3.1 Core Agent Loop

**File**: `backend/internal/ai/agent/loop.go`

```go
func (s *AgentSession) Run(ctx context.Context, client *ChatWSClient) error {
    for step := 0; step < s.Config.MaxSteps; step++ {
        // Get tools for current mode
        tools := s.getToolsForMode()
        
        // Stream to MiniMax
        stream, err := minimaxProvider.StreamWithTools(ctx, s.Messages, tools, "auto")
        if err != nil {
            return err
        }
        
        // Process stream
        toolCalls, assistantText := s.processStream(stream)
        
        // If no tool calls, we're done
        if len(toolCalls) == 0 {
            // Send final text to client
            sendAssistantFinal(client, assistantText)
            return nil
        }
        
        // Process each tool call
        for _, tc := range toolCalls {
            decision := s.policy.Decide(tc.Tool, tc.Args)
            
            switch decision {
            case DecisionAllow:
                result := s.executeTool(tc.Tool, tc.Args)
                s.addToolMessage(tc.ID, tc.Tool, result)
                sendToolResult(client, tc.ID, tc.Tool, result)
                
            case DecisionConfirm:
                sendApprovalRequired(client, tc.ID, tc.Tool, tc.Args, tc.Tool.Policy)
                s.PendingCalls[tc.ID] = &PendingToolCall{
                    ToolCall: tc,
                    CreatedAt: time.Now(),
                }
                return nil // Wait for user response
                
            case DecisionDeny:
                result := makeErrorResult(ErrCodePermission, "Blocked by policy")
                s.addToolMessage(tc.ID, tc.Tool, result)
                sendToolResult(client, tc.ID, tc.Tool, result)
            }
        }
    }
    
    // Max steps reached
    sendAssistantFinal(client, "Agent stopped: maximum steps reached")
    return nil
}

func (s *AgentSession) HandleApproval(client *ChatWSClient, toolCallID string, approved bool, reason string) error {
    pending, ok := s.PendingCalls[toolCallID]
    if !ok {
        return errors.New("unknown tool call")
    }
    
    var result tools.ToolResult
    if approved {
        result = s.executeTool(pending.ToolCall.Tool, pending.ToolCall.Args)
    } else {
        result = tools.ToolResult{
            OK: false,
            Error: &tools.ToolError{
                Code:    ErrCodeUserRejected,
                Message: "User rejected: " + reason,
            },
        }
    }
    
    s.addToolMessage(toolCallID, pending.ToolCall.Tool, result)
    sendToolResult(client, toolCallID, pending.ToolCall.Tool, result)
    delete(s.PendingCalls, toolCallID)
    
    // Continue loop
    go s.Run(context.Background(), client)
    return nil
}
```

### 3.2 Tool Execution

```go
func (s *AgentSession) executeTool(name string, args map[string]interface{}) tools.ToolResult {
    start := time.Now()
    
    tool, ok := toolRegistry.Get(name)
    if !ok {
        return tools.ToolResult{
            OK: false,
            Error: &tools.ToolError{
                Code: ErrCodeNotFound,
                Message: "Tool not found: " + name,
            },
        }
    }
    
    // Validate args against schema
    if err := toolRegistry.ValidateArgs(name, args); err != nil {
        return tools.ToolResult{
            OK: false,
            Error: &tools.ToolError{
                Code: ErrCodeValidation,
                Message: err.Error(),
            },
        }
    }
    
    // Create tool context
    ctx := tools.ToolContext{
        SessionID:  s.ID,
        ProjectID:  s.ProjectID,
        UserID:     s.UserID,
        Mode:       s.Mode,
        Limits:     s.Config.Limits,
    }
    
    // Execute
    result, err := tool.Execute(context.Background(), args, ctx)
    
    if err != nil {
        return tools.ToolResult{
            OK: false,
            Error: &tools.ToolError{
                Code: ErrCodeExecution,
                Message: err.Error(),
            },
        }
    }
    
    return tools.ToolResult{
        OK:   true,
        Data: result,
        Meta: tools.ResultMeta{
            DurationMs: time.Since(start).Milliseconds(),
        },
    }
}
```

---

## Phase 4: Frontend Changes

### 4.1 Extend AI Store

**File**: `frontend/src/stores/ai.ts`

**Add Types**:

```typescript
export interface ToolCall {
    id: string
    name: string
    arguments: Record<string, unknown>
    status: 'pending' | 'approved' | 'rejected' | 'executing' | 'completed' | 'error'
    result?: ToolResult
    policy?: string
    summary?: string
}

export interface ToolResult {
    ok: boolean
    data?: unknown
    error?: {
        code: string
        message: string
    }
    meta?: {
        duration_ms: number
        truncated?: boolean
    }
}

export interface PendingApproval {
    toolCall: ToolCall
    timestamp: number
}

export interface CommandOutput {
    handle: string
    stream: 'stdout' | 'stderr'
    text: string
    ts: number
}
```

**Extend State**:

```typescript
interface AIState {
    // ... existing fields
    toolCalls: Map<string, ToolCall>
    pendingApprovals: PendingApproval[]
    commandOutputs: Map<string, CommandOutput[]>
}
```

**Extend Message Handler**:

```typescript
function handleChatWSMessage(data: any) {
    switch (data.type) {
        case 'chunk':
            // Existing streaming logic
            break
            
        case 'message_created':
            // Existing final message logic
            break
            
        case 'tool.call':
            handleToolCall(data.payload)
            break
            
        case 'tool.approval_required':
            handleApprovalRequired(data.payload)
            break
            
        case 'tool.result':
            handleToolResult(data.payload)
            break
            
        case 'tool.error':
            handleToolError(data.payload)
            break
            
        case 'command.output':
            handleCommandOutput(data.payload)
            break
            
        case 'command.done':
            handleCommandDone(data.payload)
            break
    }
}
```

**Add Methods**:

```typescript
function handleToolCall(payload: ToolCallPayload) {
    const toolCall: ToolCall = {
        id: payload.id,
        name: payload.name,
        arguments: payload.arguments,
        status: 'pending'
    }
    toolCalls.value.set(payload.id, toolCall)
}

function handleApprovalRequired(payload: ToolApprovalPayload) {
    const toolCall = toolCalls.value.get(payload.id)
    if (toolCall) {
        toolCall.status = 'pending'
        toolCall.policy = payload.policy
        toolCall.summary = payload.summary
        
        pendingApprovals.value.push({
            toolCall,
            timestamp: Date.now()
        })
    }
}

async function approveTool(toolCallId: string) {
    const response = await fetch(`/api/v1/ws/ai/chats/${chatId}`, {
        method: 'WS_SEND', // or whatever the WS send pattern is
        body: JSON.stringify({
            type: 'tool.approve',
            tool_call_id: toolCallId
        })
    })
}

async function rejectTool(toolCallId: string, reason?: string) {
    // Similar to approveTool
}
```

### 4.2 Add UI Components

**New Components**:

```
frontend/src/components/
├── ToolCallCard.vue       # Display tool call with arguments
├── ToolApprovalCard.vue   # Approve/Reject buttons
├── ToolResultCard.vue     # Display tool result
└── CommandOutput.vue      # Stream terminal output
```

**ToolCallCard.vue**:

```vue
<template>
  <div class="tool-call">
    <div class="tool-header">
      <span class="tool-icon">🔧</span>
      <span class="tool-name">{{ tool.name }}</span>
    </div>
    <pre class="tool-args">{{ formattedArgs }}</pre>
  </div>
</template>
```

**ToolApprovalCard.vue**:

```vue
<template>
  <div class="tool-approval" :class="tool.policy">
    <div class="approval-header">
      <span class="warning-icon">⚠️</span>
      <span>Tool requires approval</span>
    </div>
    <div class="tool-info">
      <strong>{{ tool.name }}</strong>
      <pre>{{ tool.arguments }}</pre>
      <p v-if="tool.summary" class="summary">{{ tool.summary }}</p>
    </div>
    <div class="actions">
      <button @click="$emit('approve')" class="approve">Approve</button>
      <button @click="$emit('reject')" class="reject">Reject</button>
    </div>
  </div>
</template>
```

### 4.3 Update AIPane.vue

```vue
<template>
  <div class="ai-pane">
    <!-- Message list -->
    <div v-for="msg in messages" :key="msg.id" class="message">
      <!-- Existing message rendering -->
      
      <!-- Tool calls -->
      <ToolCallCard 
        v-for="tc in msg.tool_calls" 
        :key="tc.id" 
        :tool="tc"
      />
      
      <!-- Tool results -->
      <ToolResultCard 
        v-for="tr in msg.tool_results" 
        :key="tr.id" 
        :result="tr"
      />
    </div>
    
    <!-- Pending approvals panel -->
    <div v-if="pendingApprovals.length" class="pending-approvals">
      <ToolApprovalCard 
        v-for="approval in pendingApprovals"
        :key="approval.toolCall.id"
        :tool="approval.toolCall"
        @approve="approveTool(approval.toolCall.id)"
        @reject="rejectTool(approval.toolCall.id)"
      />
    </div>
  </div>
</template>
```

---

## Phase 5: System Prompt

**File**: `backend/internal/ai/agent/system_prompt.go`

```go
const SystemPrompt = `You are an AI assistant inside a WebIDE. You can interact with the project ONLY via the provided tools.
Your job is to help the user by inspecting files, running commands, and proposing safe code changes.

### Core rules
1. **Do not fabricate** file contents, directory listings, command output, build/test results, or git state.
   If you need information, **use tools**.
2. The workspace is the only place you can access. **Never assume** anything about files that you have not read.
3. Prefer minimal, reversible changes. Keep edits small and focused.

### How to use tools
* If you need to see what exists: call `list_dir`.
* If you need to inspect code: call `read_file` (use line ranges when possible).
* If you need to find references: call `search_in_files`.
* **Never write full files from scratch.**
  To modify code, always use `apply_patch` with a **unified diff**.
* Default behavior for edits:
  1. Generate an `apply_patch` call with `dry_run=true` first (preview).
  2. Wait for user approval.
  3. After approval, call `apply_patch` with `dry_run=false`.

### Commands and output
* To run commands, use `run_command`.
* If command output is long, use `get_command_output` to fetch only the relevant part.
* If a command is hanging, use `cancel_command`.

### Safety and confirmation
* Assume that **apply_patch** and **run_command** may require user confirmation.
  If confirmation is required, stop and wait.
* Never run destructive commands unless explicitly requested and confirmed.

### Workflow
When solving tasks, follow this loop:
1. Observe (read files / search / inspect output via tools).
2. Plan minimal fix.
3. Propose patch (dry run).
4. Validate by running appropriate commands.
5. Summarize final result.
`

func (s *AgentSession) BuildSystemPrompt() string {
    prompt := SystemPrompt
    switch s.Mode {
    case ModeSafe:
        prompt += "\nCurrent mode: SAFE (read-only unless user explicitly approves)."
    case ModeWrite:
        prompt += "\nCurrent mode: WRITE (you may propose patches, but wait for approval to apply)."
    case ModeExec:
        prompt += "\nCurrent mode: EXEC (commands may still require confirmation; prefer safe commands)."
    }
    return prompt
}
```

---

## Phase 6: Testing Plan

### 6.1 Unit Tests (Go)

| Test | File | Description |
|------|------|-------------|
| PathGuard.ResolveProjectPath | `tools/path_guard_test.go` | Test `..` escape, symlink escape, forbidden paths |
| JSONSchema validation | `tools/schema_test.go` | Valid/invalid argument validation |
| apply_patch dry_run | `builtin/apply_patch_test.go` | Patch parsing, preview generation |
| apply_patch apply | `builtin/apply_patch_test.go` | File modification, SHA tracking |
| CommandProcess lifecycle | `builtin/run_command_test.go` | Start, output streaming, cancel, cleanup |

### 6.2 Integration Scenarios

**Scenario 1: "Find build error"**

```
User: "Run npm test and fix any failures"

1. Agent calls run_command {cmd: "npm test", stream: true}
2. UI shows command output streaming
3. Agent sees test failure in output
4. Agent calls read_file {path: "src/failing_test.go"}
5. Agent calls apply_patch {patch: "...", dry_run: true}
6. UI shows preview, user approves
7. Agent calls apply_patch {patch: "...", dry_run: false}
8. Agent calls run_command {cmd: "npm test"} again
9. Success shown to user
```

**Scenario 2: "Refactor file"**

```
User: "Extract the calculateTotal function to a separate file"

1. Agent calls read_file {path: "src/utils.go"}
2. Agent calls apply_patch {patch: "...", dry_run: true} (for extraction)
3. User reviews preview, approves
4. Agent applies patch
5. Agent creates new file with extracted function
```

**Scenario 3: "Search and edit"**

```
User: "Find all uses of deprecatedFunction and replace with newFunction"

1. Agent calls search_in_files {query: "deprecatedFunction", max_results: 50}
2. Agent reviews matches
3. Agent reads relevant files
4. Agent applies patches to each file
```

---

## Phase 7: Implementation Order

### Step 1: Foundation (Days 1-2)
- [ ] Create `backend/internal/ai/agent/` package structure
- [ ] Implement `AgentSession` and `AgentOrchestrator` types
- [ ] Add basic loop without tools (pass-through to existing stream)

### Step 2: Tool Registry (Days 2-3)
- [ ] Create `backend/internal/ai/tools/` package
- [ ] Implement `ToolRegistry` with JSONSchema validation
- [ ] Add path guard with security checks
- [ ] Implement `list_dir`, `read_file`, `search_in_files`

### Step 3: MiniMax Integration (Days 3-4)
- [ ] Extend MiniMax provider for function calling
- [ ] Parse streaming tool calls from MiniMax response
- [ ] Test tool call detection and buffering

### Step 4: Tool Execution (Days 4-5)
- [ ] Implement `apply_patch` tool with dry_run
- [ ] Implement `run_command` with streaming output
- [ ] Add `get_command_output`, `cancel_command`
- [ ] Wire tools into agent loop

### Step 5: WebSocket Events (Days 5-6)
- [ ] Add new event types to chat_ws.go
- [ ] Implement tool call/result event handlers
- [ ] Add approve/reject client handlers
- [ ] Ensure backward compatibility with existing `assistant.delta`

### Step 6: Frontend (Days 6-7)
- [ ] Extend AI store with tool types
- [ ] Add message handlers for new event types
- [ ] Create `ToolCallCard`, `ToolApprovalCard`, `ToolResultCard`
- [ ] Update `AIPane.vue` to display tool events

### Step 7: Testing (Days 7-8)
- [ ] Write unit tests for path guard
- [ ] Write unit tests for schema validation
- [ ] Write integration tests for all tools
- [ ] Test full agent loop scenarios

### Step 8: Polish (Day 8-9)
- [ ] Add system prompt with mode-specific additions
- [ ] Test all three modes (safe, write, exec)
- [ ] Performance testing for streaming tools
- [ ] Documentation and code review

---

## Files to Create/Modify Summary

### New Files (Backend)
- `backend/internal/ai/agent/agent.go`
- `backend/internal/ai/agent/messages.go`
- `backend/internal/ai/agent/config.go`
- `backend/internal/ai/agent/loop.go`
- `backend/internal/ai/agent/policy.go`
- `backend/internal/ai/agent/events.go`
- `backend/internal/ai/tools/registry.go`
- `backend/internal/ai/tools/executor.go`
- `backend/internal/ai/tools/schema.go`
- `backend/internal/ai/tools/errors.go`
- `backend/internal/ai/tools/path_guard.go`
- `backend/internal/ai/tools/builtin/list_dir.go`
- `backend/internal/ai/tools/builtin/read_file.go`
- `backend/internal/ai/tools/builtin/search.go`
- `backend/internal/ai/tools/builtin/apply_patch.go`
- `backend/internal/ai/tools/builtin/run_command.go`

### Modified Files (Backend)
- `backend/internal/ai/chat_ws.go` - Add event handlers, modify handleSendMessage
- `backend/internal/ai/provider/minimax.go` - Add StreamWithTools, tool parsing
- `backend/cmd/ide-server/main.go` - Register new routes if needed

### New Files (Frontend)
- `frontend/src/components/ToolCallCard.vue`
- `frontend/src/components/ToolApprovalCard.vue`
- `frontend/src/components/ToolResultCard.vue`
- `frontend/src/components/CommandOutput.vue`

### Modified Files (Frontend)
- `frontend/src/stores/ai.ts` - Add tool types, handlers
- `frontend/src/pages/AIPane.vue` - Display tool components

---

## Key Considerations

1. **Backward Compatibility**: The existing `assistant.delta` and `message_created` events must continue working exactly as before for non-tool conversations.

2. **MiniMax API Compatibility**: Verify the exact format MiniMax expects for function calling. May need to adapt to their specific API if they don't follow OpenAI's format exactly.

3. **Streaming Tool Output**: The `run_command` tool needs to stream output via WebSocket. Consider rate limiting to prevent overwhelming the connection.

4. **Security**: Path guard is critical. Test symlink escape, `..` traversal, and access to `/proc`, `/sys`, `/dev`.

5. **Tool Timeouts**: Implement proper timeouts for all tools. `run_command` especially needs a timeout to prevent hanging.

6. **Resource Limits**: Implement `MaxOutputBytes`, `MaxFileBytes`, and `MaxSteps` to prevent runaway agent behavior.

---

Do you want me to clarify any part of this plan, or shall I begin implementation?