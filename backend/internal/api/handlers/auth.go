package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/webide/ide/backend/internal/auth"
	"github.com/webide/ide/backend/internal/config"
	"github.com/webide/ide/backend/internal/db"
	"github.com/webide/ide/backend/internal/models"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	User  *UserResponse `json:"user"`
	Token string        `json:"token"`
}

type UserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func Login(c *fiber.Ctx, cfg *config.Config) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	if req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "email and password required"})
	}

	var user models.User
	err := db.GetDB().QueryRow(`
		SELECT id, email, password_hash, created_at FROM users WHERE email = ?`,
		req.Email).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt)

	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid credentials"})
	}

	if !auth.VerifyPassword(req.Password, user.PasswordHash) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid credentials"})
	}

	token, err := auth.GenerateSessionToken()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create session"})
	}

	expiresAt := auth.SessionExpiry(cfg.SessionTTLHours)

	session := models.Session{
		ID:         uuid.New(),
		UserID:     user.ID,
		Token:      token,
		ExpiresAt:  expiresAt,
		CreatedAt:  time.Now(),
		LastSeenAt: time.Now(),
	}

	_, err = db.GetDB().Exec(`
		INSERT INTO sessions (id, user_id, token, expires_at, created_at, last_seen_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		session.ID, session.UserID, session.Token, session.ExpiresAt, session.CreatedAt, session.LastSeenAt)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to save session"})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "session_token",
		Value:    token,
		Expires:  expiresAt,
		HTTPOnly: true,
		SameSite: "lax",
		Path:     "/",
	})

	return c.JSON(LoginResponse{
		User: &UserResponse{
			ID:        user.ID.String(),
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
		Token: token,
	})
}

func GetMe(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "not authenticated"})
	}

	var user models.User
	err := db.GetDB().QueryRow(`
		SELECT id, email, created_at FROM users WHERE id = ?`,
		userID).Scan(&user.ID, &user.Email, &user.CreatedAt)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}

	return c.JSON(UserResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	})
}

func Logout(c *fiber.Ctx) error {
	token := c.Cookies("session_token")
	if token == "" {
		return c.JSON(fiber.Map{"message": "already logged out"})
	}

	_, err := db.GetDB().Exec("DELETE FROM sessions WHERE token = ?", token)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to logout"})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
	})

	return c.JSON(fiber.Map{"message": "logged out"})
}
