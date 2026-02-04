package ai

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/webide/ide/backend/internal/ai/provider"
	"github.com/webide/ide/backend/internal/config"
	"github.com/webide/ide/backend/internal/db"
	"github.com/webide/ide/backend/internal/models"
)

type ChatWSClient struct {
	chatID uuid.UUID
	userID uuid.UUID
	conn   *websocket.Conn
	send   chan []byte
	mu     sync.Mutex
	ctx    context.Context
	cancel context.CancelFunc
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
	ID        string    `json:"id"`
	ChatID    string    `json:"chat_id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
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

	return websocket.New(func(conn *websocket.Conn) {
		defer conn.Close()

		ctx, cancel := context.WithCancel(context.Background())
		client := &ChatWSClient{
			chatID: chatID,
			userID: userID,
			conn:   conn,
			send:   make(chan []byte, 256),
			ctx:    ctx,
			cancel: cancel,
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

	p := provider.NewMiniMax()
	chunks, err := p.Stream(ctx, messages, providerCfg)
	if err != nil {
		log.Printf("Failed to start streaming: %v", err)
		return
	}

	var contentBuilder strings.Builder
	log.Printf("[WS-CHAT] Starting to read chunks...")
	for chunk := range chunks {
		log.Printf("[WS-CHAT] Received chunk: content='%s', done=%v", chunk.Content, chunk.Done)
		if chunk.Done {
			log.Printf("[WS-CHAT] Chunk marked as done, breaking loop")
			break
		}
		contentBuilder.WriteString(chunk.Content)
		log.Printf("[WS-CHAT] Current built content: '%s' (len=%d)", contentBuilder.String(), contentBuilder.Len())

		chunkJSON, _ := json.Marshal(ChatWSMessage{
			Type: "chunk",
			Payload: MessageChunkPayload{
				MessageID: aiMsgID.String(),
				Content:   chunk.Content,
				Done:      false,
			},
		})
		c.send <- chunkJSON
	}

	aiMsg.Content = contentBuilder.String()
	log.Printf("[WS-CHAT] Final AI message content: '%s' (len=%d)", aiMsg.Content, len(aiMsg.Content))
	db.Update(ctx, "chat_messages", aiMsg)

	doneJSON, _ := json.Marshal(ChatWSMessage{
		Type: "chunk",
		Payload: MessageChunkPayload{
			MessageID: aiMsgID.String(),
			Content:   "",
			Done:      true,
		},
	})
	c.send <- doneJSON

	aiMsgJSON, _ := json.Marshal(MessageCreatedPayload{
		ID:        aiMsg.ID.String(),
		ChatID:    c.chatID.String(),
		Role:      "assistant",
		Content:   aiMsg.Content,
		CreatedAt: aiMsg.CreatedAt,
	})
	c.send <- aiMsgJSON
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
