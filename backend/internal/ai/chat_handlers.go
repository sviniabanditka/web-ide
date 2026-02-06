package ai

import (
	"database/sql"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/webide/ide/backend/internal/db"
	"github.com/webide/ide/backend/internal/models"
)

func RegisterChatRoutes(router fiber.Router) {
	log.Println("RegisterChatRoutes: starting...")

	chats := router.Group("/projects/:id/ai/chats")
	log.Println("RegisterChatRoutes: created group /projects/:id/ai/chats")

	chats.Get("", HandleListChats)
	log.Println("RegisterChatRoutes: registered GET HandleListChats")

	chats.Post("", HandleCreateChat)
	log.Println("RegisterChatRoutes: registered POST HandleCreateChat")

	chat := chats.Group("/:chatId")
	chat.Get("", HandleGetChat)
	chat.Put("/title", HandleUpdateChatTitle)
	chat.Post("/generate-title", HandleGenerateTitle)
	chat.Delete("", HandleDeleteChat)

	chatMessages := chat.Group("/messages")
	chatMessages.Get("", HandleListChatMessages)
	chatMessages.Post("", HandleCreateChatMessage)

	chatChangesets := chat.Group("/changesets")
	chatChangesets.Get("", HandleListChatChangeSets)

	log.Println("RegisterChatRoutes: all routes registered")
}

func HandleListChats(c *fiber.Ctx) error {
	ctx := c.Context()
	projectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project_id"})
	}

	rows, err := db.Query(ctx, "SELECT id, project_id, title, status, created_at, updated_at FROM chats WHERE project_id = $1 ORDER BY updated_at DESC", projectID.String())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to query chats"})
	}
	defer rows.Close()

	var chats []models.Chat
	for rows.Next() {
		var chat models.Chat
		err := rows.Scan(&chat.ID, &chat.ProjectID, &chat.Title, &chat.Status, &chat.CreatedAt, &chat.UpdatedAt)
		if err != nil {
			continue
		}
		chats = append(chats, chat)
	}

	return c.JSON(chats)
}

func HandleCreateChat(c *fiber.Ctx) error {
	ctx := c.Context()
	projectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid project_id"})
	}

	var req struct {
		Title string `json:"title"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if req.Title == "" {
		req.Title = "New Chat"
	}

	chat := &models.Chat{
		ID:        uuid.New(),
		ProjectID: projectID,
		Title:     req.Title,
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := db.Insert(ctx, "chats", chat); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create chat"})
	}

	return c.JSON(chat)
}

func HandleGetChat(c *fiber.Ctx) error {
	ctx := c.Context()
	chatID, err := uuid.Parse(c.Params("chatId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid chat_id"})
	}

	var chat models.Chat
	if err := db.Get(ctx, &chat, "SELECT id, project_id, title, status, created_at, updated_at FROM chats WHERE id = $1", chatID.String()); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "chat not found"})
	}

	return c.JSON(chat)
}

func HandleUpdateChatTitle(c *fiber.Ctx) error {
	ctx := c.Context()
	chatID, err := uuid.Parse(c.Params("chatId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid chat_id"})
	}

	var req struct {
		Title string `json:"title"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if req.Title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "title is required"})
	}

	_, err = db.Exec(ctx, "UPDATE chats SET title = ?, updated_at = ? WHERE id = ?", req.Title, time.Now(), chatID.String())
	if err != nil {
		log.Printf("[HandleUpdateChatTitle] Failed to update: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update title"})
	}

	return c.JSON(fiber.Map{"success": true})
}

func HandleGenerateTitle(c *fiber.Ctx) error {
	ctx := c.Context()
	chatID, err := uuid.Parse(c.Params("chatId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid chat_id"})
	}

	var req struct {
		Message string `json:"message"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if req.Message == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "message is required"})
	}

	title := req.Message
	if len(title) > 50 {
		title = title[:47] + "..."
	}

	_, err = db.Exec(ctx, "UPDATE chats SET title = ?, updated_at = ? WHERE id = ?", title, time.Now(), chatID.String())
	if err != nil {
		log.Printf("[HandleGenerateTitle] Failed to update: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update title"})
	}

	return c.JSON(fiber.Map{"title": title})
}

func HandleDeleteChat(c *fiber.Ctx) error {
	ctx := c.Context()
	chatID, err := uuid.Parse(c.Params("chatId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid chat_id"})
	}

	_, err = db.Exec(ctx, "DELETE FROM chats WHERE id = $1", chatID.String())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to delete chat"})
	}

	return c.SendStatus(fiber.StatusOK)
}

func HandleListChatMessages(c *fiber.Ctx) error {
	ctx := c.Context()
	chatID, err := uuid.Parse(c.Params("chatId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid chat_id"})
	}

	log.Printf("[HandleListChatMessages] Loading messages for chat: %s", chatID.String())

	rows, err := db.Query(ctx, "SELECT id, chat_id, role, COALESCE(content, ''), COALESCE(tool_calls_json, ''), COALESCE(tool_results_json, ''), COALESCE(thinking, ''), created_at FROM chat_messages WHERE chat_id = ? ORDER BY created_at ASC", chatID.String())
	if err != nil {
		log.Printf("[HandleListChatMessages] Query error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to query messages"})
	}
	defer rows.Close()

	var messages []models.ChatMessage
	for rows.Next() {
		var msg models.ChatMessage
		var content, toolCallsJSON, toolResultsJSON, thinking string
		err := rows.Scan(&msg.ID, &msg.ChatID, &msg.Role, &content, &toolCallsJSON, &toolResultsJSON, &thinking, &msg.CreatedAt)
		if err != nil {
			log.Printf("[HandleListChatMessages] Scan error: %v", err)
			continue
		}
		msg.Content = content
		msg.ToolCallsJSON = toolCallsJSON
		msg.ToolResultsJSON = toolResultsJSON
		msg.Thinking = thinking
		log.Printf("[HandleListChatMessages] Loaded message: id=%s, role=%s, content_len=%d", msg.ID.String(), msg.Role, len(msg.Content))
		messages = append(messages, msg)
	}

	log.Printf("[HandleListChatMessages] Total messages loaded: %d", len(messages))
	return c.JSON(messages)
}

func HandleCreateChatMessage(c *fiber.Ctx) error {
	ctx := c.Context()
	chatID, err := uuid.Parse(c.Params("chatId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid chat_id"})
	}

	var req struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if req.Role == "" {
		req.Role = "user"
	}
	if req.Content == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "content is required"})
	}

	msg := &models.ChatMessage{
		ID:        uuid.New(),
		ChatID:    chatID,
		Role:      req.Role,
		Content:   req.Content,
		CreatedAt: time.Now(),
	}

	if err := db.Insert(ctx, "chat_messages", msg); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create message"})
	}

	return c.JSON(msg)
}

func HandleListChatChangeSets(c *fiber.Ctx) error {
	ctx := c.Context()
	chatID, err := uuid.Parse(c.Params("chatId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid chat_id"})
	}

	rows, err := db.Query(ctx, "SELECT id, chat_id, COALESCE(job_id, ''), title, COALESCE(diff, ''), status, COALESCE(summary_text, ''), created_at FROM chat_changesets WHERE chat_id = $1 ORDER BY created_at DESC", chatID.String())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to query changesets"})
	}
	defer rows.Close()

	var changesets []models.ChatChangeSet
	for rows.Next() {
		var cs models.ChatChangeSet
		var jobID, diff, summary sql.NullString
		err := rows.Scan(&cs.ID, &cs.ChatID, &jobID, &cs.Title, &diff, &cs.Status, &summary, &cs.CreatedAt)
		if err != nil {
			continue
		}
		if jobID.Valid && jobID.String != "" {
			id := uuid.MustParse(jobID.String)
			cs.JobID = &id
		}
		if diff.Valid {
			cs.Diff = diff.String
		}
		if summary.Valid {
			cs.SummaryText = summary.String
		}
		changesets = append(changesets, cs)
	}

	return c.JSON(changesets)
}
