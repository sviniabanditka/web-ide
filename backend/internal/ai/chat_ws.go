package ai

import (
	"context"
	"encoding/json"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/webide/ide/backend/internal/ai/provider"
	"github.com/webide/ide/backend/internal/ai/tools"
	_ "github.com/webide/ide/backend/internal/ai/tools/builtin"
	"github.com/webide/ide/backend/internal/config"
	"github.com/webide/ide/backend/internal/db"
	"github.com/webide/ide/backend/internal/models"
	"github.com/webide/ide/backend/internal/projects"
)

func init() {
	log.Printf("[Tools] GlobalRegistry tools count: %d", len(tools.GlobalRegistry.List()))
}

type FrontendToolCall struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
	Status    string                 `json:"status"`
}

func transformToolCallsToFrontend(tc []provider.ToolCall) []FrontendToolCall {
	result := make([]FrontendToolCall, len(tc))
	for i, t := range tc {
		var args map[string]interface{}
		json.Unmarshal([]byte(t.Function.Arguments), &args)
		result[i] = FrontendToolCall{
			ID:        t.ID,
			Name:      t.Function.Name,
			Arguments: args,
			Status:    "pending",
		}
	}
	return result
}

type ChatWSClient struct {
	chatID    uuid.UUID
	userID    uuid.UUID
	projectID uuid.UUID
	conn      *websocket.Conn
	send      chan []byte
	mu        sync.Mutex
	ctx       context.Context
	cancel    context.CancelFunc
}

type ChatWSHub struct {
	clients  map[uuid.UUID]*ChatWSClient
	mu       sync.RWMutex
	register chan *ChatWSClient
}

var ChatHub = &ChatWSHub{
	clients:  make(map[uuid.UUID]*ChatWSClient),
	register: make(chan *ChatWSClient, 10),
}

func (h *ChatWSHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.chatID] = client
			h.mu.Unlock()
		}
	}
}

func init() {
	go ChatHub.Run()
}

type ChatWSMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type SendMessagePayload struct {
	Content string `json:"content"`
}

type MessageChunkPayload struct {
	MessageID string `json:"message_id"`
	Content   string `json:"content"`
	Done      bool   `json:"done"`
}

type MessageCreatedPayload struct {
	ID              string    `json:"id"`
	ChatID          string    `json:"chat_id"`
	Role            string    `json:"role"`
	Content         string    `json:"content"`
	ToolCallsJSON   string    `json:"tool_calls_json,omitempty"`
	ToolResultsJSON string    `json:"tool_results_json,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

func RegisterChatWSRoutes(router fiber.Router) {
	log.Println("[WS-CHAT] Registering chat WebSocket routes...")
	router.Get("/ws/ai/chats/:chatId", ChatWebSocketHandler)
}

func ChatWebSocketHandler(c *fiber.Ctx) error {
	log.Printf("[WS-CHAT] Handler called for chat: %s", c.Params("chatId"))
	log.Printf("[WS-CHAT] Is WebSocket upgrade: %v", websocket.IsWebSocketUpgrade(c))

	chatIDStr := c.Params("chatId")
	log.Printf("[WS-CHAT] New connection attempt for chat: %s", chatIDStr)

	chatID, err := uuid.Parse(chatIDStr)
	if err != nil {
		log.Printf("[WS-CHAT] Invalid chat ID: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid chat_id"})
	}

	userIDVal := c.Locals("user_id")
	if userIDVal == nil {
		log.Printf("[WS-CHAT] User ID not found in context")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "authentication required"})
	}

	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		log.Printf("[WS-CHAT] Invalid user ID type")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "authentication required"})
	}

	log.Printf("[WS-CHAT] Authenticated user: %s", userID)

	ctx := c.Context()
	var chat models.Chat
	if err := db.Get(ctx, &chat, "SELECT id, project_id, title, status, created_at, updated_at FROM chats WHERE id = $1", chatID.String()); err != nil {
		log.Printf("[WS-CHAT] Chat not found: %v", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "chat not found"})
	}

	log.Printf("[WS-CHAT] Chat %s belongs to project: %s", chatID, chat.ProjectID)

	project, err := projects.GetProject(chat.ProjectID)
	if err != nil {
		log.Printf("[WS-CHAT] Failed to get project %s: %v", chat.ProjectID, err)
	} else {
		log.Printf("[WS-CHAT] Project: %s, RootPath: %s", project.Name, project.RootPath)
	}

	return websocket.New(func(conn *websocket.Conn) {
		defer conn.Close()

		wsCtx, cancel := context.WithCancel(context.Background())
		client := &ChatWSClient{
			chatID:    chatID,
			userID:    userID,
			projectID: chat.ProjectID,
			conn:      conn,
			send:      make(chan []byte, 256),
			ctx:       wsCtx,
			cancel:    cancel,
		}

		ChatHub.register <- client

		go client.writePump()

		client.readPump(ctx)
	}, websocket.Config{
		HandshakeTimeout: 10 * time.Second,
	})(c)
}

func (c *ChatWSClient) readPump(ctx context.Context) {
	defer func() {
		c.cancel()
		c.conn.Close()
	}()

	log.Printf("[WS-CHAT] readPump started for chat: %s", c.chatID)

	for {
		select {
		case <-ctx.Done():
			log.Printf("[WS-CHAT] Context cancelled for chat: %s", c.chatID)
			return
		default:
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				log.Printf("[WS-CHAT] Read error for chat %s: %v", c.chatID, err)
				return
			}

			log.Printf("[WS-CHAT] Received message for chat %s: %s", c.chatID, string(message))

			var msg ChatWSMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				log.Printf("[WS-CHAT] Failed to parse message: %v", err)
				continue
			}

			switch msg.Type {
			case "send_message":
				log.Printf("[WS-CHAT] Processing send_message for chat: %s", c.chatID)
				c.handleSendMessage(msg.Payload)
			case "stop":
				log.Printf("[WS-CHAT] Stop requested for chat: %s", c.chatID)
				c.cancel()
			}
		}
	}
}

func (c *ChatWSClient) handleSendMessage(payload interface{}) {
	log.Printf("[WS-CHAT] handleSendMessage called with payload: %v", payload)

	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[WS-CHAT] Failed to marshal payload: %v", err)
		return
	}

	var sendPayload SendMessagePayload
	if err := json.Unmarshal(data, &sendPayload); err != nil {
		log.Printf("[WS-CHAT] Failed to unmarshal payload: %v", err)
		return
	}

	log.Printf("[WS-CHAT] Content: '%s'", sendPayload.Content)

	if sendPayload.Content == "" {
		log.Printf("[WS-CHAT] Empty content, returning")
		return
	}

	ctx := c.ctx
	now := time.Now()

	userMsg := &models.ChatMessage{
		ID:        uuid.New(),
		ChatID:    c.chatID,
		Role:      "user",
		Content:   sendPayload.Content,
		CreatedAt: now,
	}

	log.Printf("[WS-CHAT] Saving user message to DB...")
	if err := db.Insert(ctx, "chat_messages", userMsg); err != nil {
		log.Printf("[WS-CHAT] Failed to save user message: %v", err)
		return
	}
	log.Printf("[WS-CHAT] User message saved: %s", userMsg.ID)

	userMsgJSON, _ := json.Marshal(MessageCreatedPayload{
		ID:        userMsg.ID.String(),
		ChatID:    c.chatID.String(),
		Role:      "user",
		Content:   userMsg.Content,
		CreatedAt: userMsg.CreatedAt,
	})
	c.send <- userMsgJSON

	aiMsgID := uuid.New()
	aiMsg := &models.ChatMessage{
		ID:        aiMsgID,
		ChatID:    c.chatID,
		Role:      "assistant",
		Content:   "",
		CreatedAt: time.Now(),
	}

	log.Printf("[WS-CHAT] Creating AI message...")
	if err := db.Insert(ctx, "chat_messages", aiMsg); err != nil {
		log.Printf("[WS-CHAT] Failed to create AI message: %v", err)
		return
	}
	log.Printf("[WS-CHAT] AI message created: %s", aiMsgID)

	messages, err := c.getChatMessages()
	if err != nil {
		log.Printf("[WS-CHAT] Failed to get chat messages: %v", err)
		return
	}
	log.Printf("[WS-CHAT] Got %d messages for context", len(messages))

	projectRoot := ""
	project, err := projects.GetProject(c.projectID)
	if err != nil {
		log.Printf("[WS-CHAT] Failed to get project %s: %v, using empty root", c.projectID, err)
	} else {
		projectRoot = project.RootPath
		log.Printf("[WS-CHAT] Project root: %s", projectRoot)
	}

	cfg, err := config.Load()
	if err != nil || cfg == nil {
		log.Printf("Failed to load config: %v", err)
		return
	}

	providerCfg := provider.Config{
		URL:    cfg.MiniMaxURL,
		APIKey: cfg.MiniMaxAPIKey,
		Model:  cfg.MiniMaxModel,
	}

	p := provider.NewAnthropic(cfg.MiniMaxAPIKey, cfg.MiniMaxURL)

	toolsList := tools.GlobalRegistry.ListForModel()
	log.Printf("[WS-CHAT] Sending %d tools to model", len(toolsList))
	for i, t := range toolsList {
		log.Printf("[WS-CHAT] Tool %d: %s", i, t.Function["name"])
	}

	providerTools := make([]provider.ToolDefinition, len(toolsList))
	for i, t := range toolsList {
		providerTools[i] = provider.ToolDefinition{
			Type:     t.Type,
			Function: t.Function,
		}
	}

	log.Printf("[WS-CHAT] Starting AI response processing...")

	var allToolCalls []provider.ToolCall
	var toolResults []map[string]interface{}

	aiResponseIndex := 0
	maxIterations := 5

	for aiResponseIndex < maxIterations {
		aiResponseIndex++

		newToolCalls := []provider.ToolCall{}

		thinkingTime := time.Now().Add(-time.Millisecond)
		currentThinkingMsgID := uuid.New()
		currentThinkingMsg := &models.ChatMessage{
			ID:        currentThinkingMsgID,
			ChatID:    c.chatID,
			Role:      "thinking",
			Content:   "",
			CreatedAt: thinkingTime,
		}
		if err := db.Insert(ctx, "chat_messages", currentThinkingMsg); err != nil {
			log.Printf("[WS-CHAT] Failed to create thinking message: %v", err)
		}
		thinkingMsgJSON, _ := json.Marshal(ChatWSMessage{
			Type: "message_created",
			Payload: MessageCreatedPayload{
				ID:        currentThinkingMsgID.String(),
				ChatID:    c.chatID.String(),
				Role:      "thinking",
				Content:   "",
				CreatedAt: thinkingTime,
			},
		})
		c.send <- thinkingMsgJSON

		statusJSON, _ := json.Marshal(ChatWSMessage{
			Type: "status",
			Payload: map[string]interface{}{
				"status": "thinking",
			},
		})
		c.send <- statusJSON

		chunks, err := p.StreamWithTools(ctx, messages, providerCfg, providerTools, "auto")
		if err != nil {
			log.Printf("[WS-CHAT] Streaming failed: %v", err)
			break
		}

		hasNewContent := false
		hasNewThinking := false

		currentAIMsgID := uuid.New()
		currentAIMsg := &models.ChatMessage{
			ID:        currentAIMsgID,
			ChatID:    c.chatID,
			Role:      "assistant",
			Content:   "",
			CreatedAt: time.Now(),
		}

		if err := db.Insert(ctx, "chat_messages", currentAIMsg); err != nil {
			log.Printf("[WS-CHAT] Failed to create AI message: %v", err)
			return
		}

		aiMsgJSON, _ := json.Marshal(ChatWSMessage{
			Type: "message_created",
			Payload: MessageCreatedPayload{
				ID:        currentAIMsg.ID.String(),
				ChatID:    c.chatID.String(),
				Role:      "assistant",
				Content:   "",
				CreatedAt: currentAIMsg.CreatedAt,
			},
		})
		c.send <- aiMsgJSON

		var contentBuilder strings.Builder
		var thinkingBuilder strings.Builder

		for chunk := range chunks {
			log.Printf("[WS-CHAT] Chunk: content_len=%d, thinking_len=%d, done=%v, tool_calls=%d",
				len(chunk.Content), len(chunk.Thinking), chunk.Done, len(chunk.ToolCalls))

			if len(chunk.ToolCalls) > 0 {
				log.Printf("[WS-CHAT] Tool calls detected: %d", len(chunk.ToolCalls))
				for _, tc := range chunk.ToolCalls {
					log.Printf("[WS-CHAT] Tool call: %s(%s)", tc.Function.Name, tc.Function.Arguments)
				}
				newToolCalls = append(newToolCalls, chunk.ToolCalls...)
				allToolCalls = append(allToolCalls, chunk.ToolCalls...)
			}

			if chunk.Thinking != "" {
				hasNewThinking = true
				thinkingBuilder.WriteString(chunk.Thinking)
				currentThinkingMsg.Content = thinkingBuilder.String()
				db.Update(ctx, "chat_messages", currentThinkingMsg)
				chunkThinkingJSON, _ := json.Marshal(ChatWSMessage{
					Type: "chunk",
					Payload: MessageChunkPayload{
						MessageID: currentThinkingMsgID.String(),
						Content:   chunk.Thinking,
						Done:      false,
					},
				})
				c.send <- chunkThinkingJSON
			}

			if chunk.Content != "" {
				hasNewContent = true
				contentBuilder.WriteString(chunk.Content)
				chunkJSON, _ := json.Marshal(ChatWSMessage{
					Type: "chunk",
					Payload: MessageChunkPayload{
						MessageID: currentAIMsgID.String(),
						Content:   chunk.Content,
						Done:      false,
					},
				})
				c.send <- chunkJSON
			}

			if chunk.Done {
				break
			}
		}

		currentAIMsg.Content = contentBuilder.String()
		currentAIMsg.Thinking = thinkingBuilder.String()
		toolCallsJSON, _ := json.Marshal(allToolCalls)
		currentAIMsg.ToolCallsJSON = string(toolCallsJSON)
		db.Update(ctx, "chat_messages", currentAIMsg)

		doneThinkingJSON, _ := json.Marshal(ChatWSMessage{
			Type: "chunk",
			Payload: MessageChunkPayload{
				MessageID: currentThinkingMsgID.String(),
				Content:   "",
				Done:      true,
			},
		})
		c.send <- doneThinkingJSON

		doneJSON, _ := json.Marshal(ChatWSMessage{
			Type: "chunk",
			Payload: MessageChunkPayload{
				MessageID: currentAIMsgID.String(),
				Content:   "",
				Done:      true,
			},
		})
		c.send <- doneJSON

		frontendToolCalls := transformToolCallsToFrontend(newToolCalls)
		frontendToolCallsJSON, _ := json.Marshal(frontendToolCalls)
		toolResultsJSON, _ := json.Marshal(toolResults)
		aiMsgJSON, _ = json.Marshal(ChatWSMessage{
			Type: "message_created",
			Payload: MessageCreatedPayload{
				ID:              currentAIMsg.ID.String(),
				ChatID:          c.chatID.String(),
				Role:            "assistant",
				Content:         currentAIMsg.Content,
				ToolCallsJSON:   string(frontendToolCallsJSON),
				ToolResultsJSON: string(toolResultsJSON),
				CreatedAt:       currentAIMsg.CreatedAt,
			},
		})
		c.send <- aiMsgJSON

		log.Printf("[WS-CHAT] AI response %d done: content='%s', thinking='%s', new_tool_calls=%d, total=%d",
			aiResponseIndex, currentAIMsg.Content, currentAIMsg.Thinking, len(newToolCalls), len(allToolCalls))

		// If no new tool calls were made, we're done
		if len(newToolCalls) == 0 {
			log.Printf("[WS-CHAT] No new tool calls, finishing")
			break
		}

		// Execute tool calls
		for _, tc := range newToolCalls {
			log.Printf("[WS-CHAT] Executing tool: %s", tc.Function.Name)

			var args map[string]interface{}
			if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
				log.Printf("[WS-CHAT] Failed to parse tool args: %v", err)
				continue
			}

			toolCallJSON, _ := json.Marshal(ChatWSMessage{
				Type: "tool_call",
				Payload: map[string]interface{}{
					"id":               tc.ID,
					"name":             tc.Function.Name,
					"arguments":        args,
					"assistant_msg_id": currentAIMsgID.String(),
				},
			})
			c.send <- toolCallJSON

			tcToolContext := tools.ToolContext{
				SessionID:   c.userID,
				ProjectID:   c.chatID,
				UserID:      c.userID,
				ProjectRoot: projectRoot,
				Mode:        "write",
				Limits: tools.ToolLimits{
					MaxFileBytes:     512 * 1024,
					MaxOutputBytes:   1024 * 1024,
					MaxSearchResults: 200,
					MaxPatchFiles:    10,
					MaxToolTime:      5 * time.Minute,
				},
			}

			log.Printf("[WS-CHAT] Executing tool with projectRoot=%s", projectRoot)
			result := tools.GlobalExecute(ctx, tc.Function.Name, args, tcToolContext)

			log.Printf("[WS-CHAT] Tool result: ok=%v", result.OK)
			if !result.OK && result.Error != nil {
				log.Printf("[WS-CHAT] Tool error: %s - %s", result.Error.Code, result.Error.Message)
			}

			resultJSON, _ := json.Marshal(ChatWSMessage{
				Type: "tool.result",
				Payload: map[string]interface{}{
					"id":               tc.ID,
					"name":             tc.Function.Name,
					"ok":               result.OK,
					"result":           result.Data,
					"error":            result.Error,
					"assistant_msg_id": currentAIMsgID.String(),
				},
			})
			c.send <- resultJSON

			toolResults = append(toolResults, map[string]interface{}{
				"id":     tc.ID,
				"name":   tc.Function.Name,
				"ok":     result.OK,
				"result": result.Data,
				"error":  result.Error,
			})

			resultContent, _ := json.Marshal(result)
			messages = append(messages, provider.Message{
				Role:    "tool",
				Content: string(resultContent),
			})
		}

		// If no new tool calls were added during this iteration, we're done
		if len(newToolCalls) == 0 && !hasNewContent && !hasNewThinking {
			log.Printf("[WS-CHAT] No new tool calls, content, or thinking, stopping iteration")
			break
		}

		log.Printf("[WS-CHAT] Making follow-up AI call %d", aiResponseIndex+1)
	}

	// Send final idle status
	statusIdleJSON, _ := json.Marshal(ChatWSMessage{
		Type: "status",
		Payload: map[string]interface{}{
			"status": "idle",
		},
	})
	c.send <- statusIdleJSON
}

func (c *ChatWSClient) getChatMessages() ([]provider.Message, error) {
	ctx := c.ctx
	rows, err := db.Query(ctx, "SELECT role, content FROM chat_messages WHERE chat_id = $1 ORDER BY created_at ASC", c.chatID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []provider.Message
	for rows.Next() {
		var role, content string
		rows.Scan(&role, &content)
		messages = append(messages, provider.Message{
			Role:    role,
			Content: content,
		})
	}

	return messages, nil
}

func resolveToolPaths(args map[string]interface{}, projectRoot string) map[string]interface{} {
	if projectRoot == "" {
		return args
	}

	pathFields := []string{"path", "file_path", "dir", "directory", "target", "source"}
	modified := false

	for _, field := range pathFields {
		if val, ok := args[field]; ok {
			if strVal, isStr := val.(string); isStr {
				resolved := resolvePath(strVal, projectRoot)
				if resolved != strVal {
					args[field] = resolved
					modified = true
				}
			}
		}
	}

	if modified {
		log.Printf("[WS-CHAT] Resolved paths relative to: %s", projectRoot)
	}

	return args
}

func resolvePath(path, projectRoot string) string {
	if path == "" {
		return path
	}

	if filepath.IsAbs(path) {
		return path
	}

	if path == "." || path == "/" || path == "\\" {
		return projectRoot
	}

	cleanPath := filepath.Clean(path)
	if strings.HasPrefix(cleanPath, projectRoot) || strings.HasPrefix(filepath.Join(projectRoot, cleanPath), projectRoot) {
		return cleanPath
	}

	if !strings.HasPrefix(path, "/") && !strings.HasPrefix(path, "\\") {
		if !strings.HasPrefix(path, "./") && !strings.HasPrefix(path, ".\\") {
			return filepath.Join(projectRoot, path)
		}
	}

	cleaned := filepath.Clean(filepath.Join(projectRoot, path))
	return cleaned
}

func (c *ChatWSClient) writePump() {
	defer c.conn.Close()
	for {
		message, ok := <-c.send
		if !ok {
			c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}
		c.conn.WriteMessage(websocket.TextMessage, message)
	}
}
