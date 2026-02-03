package middleware

import (
	"database/sql"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/webide/ide/backend/internal/db"
)

func AuthRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Cookies("session_token")
		authHeader := c.Get("Authorization")

		log.Printf("AuthRequired: cookie token=%s, authHeader=%s", token, authHeader)

		if authHeader != "" {
			if strings.HasPrefix(authHeader, "Bearer ") {
				token = strings.TrimPrefix(authHeader, "Bearer ")
			} else {
				token = authHeader
			}
		}

		log.Printf("AuthRequired: using token=%s (len=%d)", token, len(token))

		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "authentication required"})
		}

		var userID uuid.UUID
		var expiresAt time.Time
		log.Printf("AuthRequired: querying DB...")

		row := db.GetDB().QueryRow(`SELECT user_id, expires_at FROM sessions WHERE token = ?`, token)
		log.Printf("AuthRequired: row=%v", row)

		err := row.Scan(&userID, &expiresAt)
		log.Printf("AuthRequired: scan result err=%v", err)

		if err == sql.ErrNoRows {
			log.Printf("AuthRequired: no rows found")
			// Debug: check what tokens exist
			var count int
			db.GetDB().QueryRow("SELECT COUNT(*) FROM sessions").Scan(&count)
			log.Printf("AuthRequired: total sessions in DB: %d", count)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid session"})
		}
		if err != nil {
			log.Printf("AuthRequired: DB error=%v", err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid session"})
		}

		log.Printf("AuthRequired: found userID=%s, expiresAt=%v", userID, expiresAt)

		if time.Now().After(expiresAt) {
			db.GetDB().Exec("DELETE FROM sessions WHERE token = ?", token)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "session expired"})
		}

		db.GetDB().Exec("UPDATE sessions SET last_seen_at = ? WHERE token = ?", time.Now(), token)

		c.Locals("user_id", userID)
		return c.Next()
	}
}

func OptionalAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Cookies("session_token")
		authHeader := c.Get("Authorization")

		if authHeader != "" {
			if strings.HasPrefix(authHeader, "Bearer ") {
				token = strings.TrimPrefix(authHeader, "Bearer ")
			} else {
				token = authHeader
			}
		}

		if token == "" {
			return c.Next()
		}

		var userID uuid.UUID
		var expiresAt time.Time
		err := db.GetDB().QueryRow(`
			SELECT user_id, expires_at FROM sessions WHERE token = ?`,
			token).Scan(&userID, &expiresAt)
		if err == nil && time.Now().Before(expiresAt) {
			db.GetDB().Exec("UPDATE sessions SET last_seen_at = ? WHERE token = ?", time.Now(), token)
			c.Locals("user_id", userID)
		}

		return c.Next()
	}
}
