package ai

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
	"github.com/webide/ide/backend/internal/ai/provider"
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

func RegisterChatWSRoutes(app *fiber.App) {
	app.Use("/ws/ai/chats/:chatId", func(c *fiber.Ctx) error {
		if !websocket.IsWebSocketUpgrade(c) {
			return fiber.ErrUpgradeRequired
		}
		return nil
	})

	app.Get("/ws/ai/chats/:chatId", websocket.New(func(c *websocket.Conn) {
		chatIDStr := c.Params("chatId")
		chatID, err := uuid.Parse(chatIDStr)
		if err != nil {
			log.Printf("Invalid chat ID: %v", err)
			c.Close()
			return
		}

		ctx, cancel := context.WithCancel(context.Background())
		client := &ChatWSClient{
			chatID: chatID,
			userID: uuid.Nil,
			conn:   c,
			send:   make(chan []byte, 256),
			ctx:    ctx,
			cancel: cancel,
		}

		ChatHub.register <- client

		go client.writePump()

		client.readPump(ctx)
	}))
}

func (c *ChatWSClient) readPump(ctx context.Context) {
	defer func() {
		c.cancel()
		c.conn.Close()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				return
			}

			var msg ChatWSMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				continue
			}

			switch msg.Type {
			case "send_message":
				c.handleSendMessage(msg.Payload)
			case "stop":
				c.cancel()
			}
		}
	}
}

func (c *ChatWSClient) handleSendMessage(payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}

	var sendPayload SendMessagePayload
	if err := json.Unmarshal(data, &sendPayload); err != nil {
		return
	}

	if sendPayload.Content == "" {
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

	if err := db.Insert(ctx, "chat_messages", userMsg); err != nil {
		log.Printf("Failed to save user message: %v", err)
		return
	}

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

	if err := db.Insert(ctx, "chat_messages", aiMsg); err != nil {
		log.Printf("Failed to create AI message: %v", err)
		return
	}

	messages, err := c.getChatMessages()
	if err != nil {
		log.Printf("Failed to get chat messages: %v", err)
		return
	}

	cfg := provider.Config{
		APIKey: "",
		Model:  "abab6.5s-chat",
	}

	p := provider.NewMiniMax()
	chunks, err := p.Stream(ctx, messages, cfg)
	if err != nil {
		log.Printf("Failed to start streaming: %v", err)
		return
	}

	var contentBuilder strings.Builder
	for chunk := range chunks {
		if chunk.Done {
			break
		}
		contentBuilder.WriteString(chunk.Content)

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
