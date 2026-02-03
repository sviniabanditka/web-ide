package handlers

import (
	"log"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/webide/ide/backend/internal/db"
	"github.com/webide/ide/backend/internal/terminal"
)

func TerminalWS(c *fiber.Ctx) error {
	termIDStr := c.Params("tid")
	termID, err := uuid.Parse(termIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid terminal id"})
	}

	token := c.Cookies("session_token")
	if token == "" {
		token = c.Query("token")
	}

	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "authentication required"})
	}

	var userID uuid.UUID
	var expiresAt time.Time
	err = db.GetDB().QueryRow(`
		SELECT user_id, expires_at FROM sessions WHERE token = ?`,
		token).Scan(&userID, &expiresAt)
	if err != nil || time.Now().After(expiresAt) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid session"})
	}

	return websocket.New(func(conn *websocket.Conn) {
		defer conn.Close()

		session, err := terminal.GetSession(termID)
		if err != nil {
			log.Printf("Terminal WS: session not found %s", termID)
			return
		}

		log.Printf("Terminal WS: connected, session=%s", termID)

		conn.SetPongHandler(func(appData string) error {
			conn.SetReadDeadline(time.Now().Add(60 * time.Second))
			return nil
		})

		// Start polling for new output (50ms interval)
		go func() {
			var lastLen int
			ticker := time.NewTicker(50 * time.Millisecond)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					data := session.GetBacklog()
					if len(data) > lastLen {
						newData := data[lastLen:]
						if len(newData) > 0 {
							if err := conn.WriteMessage(websocket.TextMessage, newData); err != nil {
								return
							}
						}
						lastLen = len(data)
					}
				}
			}
		}()

		go func() {
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
						return
					}
				}
			}
		}()

		for {
			msgType, data, err := conn.ReadMessage()
			if err != nil {
				break
			}

			if msgType == websocket.TextMessage {
				msg, parseErr := terminal.ParseWSMessage(data)
				if parseErr != nil {
					if err := session.Write(data); err != nil {
						log.Printf("Terminal WS: failed to write: %v", err)
						break
					}
					continue
				}

				switch msg.Type {
				case "stdin":
					if err := session.Write([]byte(msg.Data)); err != nil {
						log.Printf("Terminal WS: failed to write: %v", err)
						break
					}
				case "resize":
					if msg.Cols > 0 && msg.Rows > 0 {
						session.Resize(msg.Cols, msg.Rows)
					}
				case "ping":
					conn.WriteMessage(websocket.PongMessage, nil)
				}
			} else if msgType == websocket.BinaryMessage {
				if err := session.Write(data); err != nil {
					log.Printf("Terminal WS: failed to write: %v", err)
					break
				}
			}
		}

	}, websocket.Config{
		HandshakeTimeout: 10 * time.Second,
	})(c)
}
