package agent

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

type AgentSession struct {
	ID           uuid.UUID
	ProjectID    uuid.UUID
	UserID       uuid.UUID
	ChatID       uuid.UUID
	Messages     []ModelMessage
	Mode         AgentMode
	PendingCalls map[string]*PendingToolCall
	RunningCmds  map[string]*CommandProcess
	Config       AgentConfig
	mu           sync.RWMutex
}

type PendingToolCall struct {
	ToolCall     ToolCall
	Args         map[string]interface{}
	Approved     bool
	RejectReason string
	CreatedAt    time.Time
}

type CommandProcess struct {
	Handle    string
	Cmd       interface{}
	StartedAt time.Time
	mu        sync.Mutex
	Done      bool
	ExitCode  int
}

func NewSession(projectID, userID, chatID uuid.UUID, config AgentConfig) *AgentSession {
	if config.Limits.MaxSteps == 0 {
		config.Limits = DefaultConfig().Limits
	}
	if config.Mode == "" {
		config.Mode = ModeSafe
	}
	return &AgentSession{
		ID:           uuid.New(),
		ProjectID:    projectID,
		UserID:       userID,
		ChatID:       chatID,
		Messages:     make([]ModelMessage, 0),
		Mode:         config.Mode,
		PendingCalls: make(map[string]*PendingToolCall),
		RunningCmds:  make(map[string]*CommandProcess),
		Config:       config,
	}
}

func (s *AgentSession) AddMessage(role MessageRole, content string) {
	msg := ModelMessage{
		ID:        generateID(),
		Role:      role,
		Content:   content,
		Timestamp: now(),
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Messages = append(s.Messages, msg)
}

func (s *AgentSession) AddSystemMessage(content string) {
	s.AddMessage(RoleSystem, content)
}

func (s *AgentSession) AddUserMessage(content string) {
	s.AddMessage(RoleUser, content)
}

func (s *AgentSession) AddAssistantMessage(content string, toolCalls []ToolCall) {
	msg := NewAssistantMessage(content, toolCalls)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Messages = append(s.Messages, msg)
}

func (s *AgentSession) AddToolResult(toolCallID, name, content string) {
	msg := NewToolResultMessage(toolCallID, name, content)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Messages = append(s.Messages, msg)
}

func (s *AgentSession) GetMessages() []ModelMessage {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Messages
}

func (s *AgentSession) GetPendingToolCall(id string) (*PendingToolCall, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	call, ok := s.PendingCalls[id]
	return call, ok
}

func (s *AgentSession) RemovePendingToolCall(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.PendingCalls, id)
}

func (s *AgentSession) SetPendingToolCall(id string, call *PendingToolCall) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.PendingCalls[id] = call
}

func (s *AgentSession) AddRunningCommand(handle string, proc *CommandProcess) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.RunningCmds[handle] = proc
}

func (s *AgentSession) GetRunningCommand(handle string) (*CommandProcess, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	proc, ok := s.RunningCmds[handle]
	return proc, ok
}

func (s *AgentSession) RemoveRunningCommand(handle string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.RunningCmds, handle)
}

func (s *AgentSession) GetMode() AgentMode {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Mode
}

func (s *AgentSession) SetMode(mode AgentMode) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Mode = mode
}

func (s *AgentSession) GetConfig() AgentConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Config
}

func (s *AgentSession) Context() context.Context {
	return context.WithValue(context.Background(), "session", s)
}
