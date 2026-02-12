package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/google/uuid"
	"github.com/webide/ide/backend/internal/ai"
	"github.com/webide/ide/backend/internal/api/handlers"
	"github.com/webide/ide/backend/internal/auth"
	"github.com/webide/ide/backend/internal/config"
	"github.com/webide/ide/backend/internal/db"
	"github.com/webide/ide/backend/internal/middleware"
	"github.com/webide/ide/backend/internal/terminal"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	config.PrintEnvVars()

	log.Printf("[MAIN] Using DataDir: %s", cfg.DataDir)

	if err := db.Init(cfg.DataDir); err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}
	defer db.Close()

	if err := bootstrapUser(cfg); err != nil {
		log.Printf("Warning: bootstrap user failed: %v", err)
	}

	app := fiber.New(fiber.Config{
		AppName:      "WebIDE",
		ErrorHandler: customErrorHandler,
	})

	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${latency} ${ip} ${method} ${path}\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	setupRoutes(app, cfg)

	go terminal.CollectGarbage()

	log.Printf("Starting IDE server on %s", cfg.HTTPAddr)

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down server...")
		app.Shutdown()
	}()

	if err := app.Listen(cfg.HTTPAddr); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func setupRoutes(app *fiber.App, cfg *config.Config) {
	api := app.Group("/api/v1")

	public := api.Group("")
	public.Post("/auth/login", func(c *fiber.Ctx) error {
		return handlers.Login(c, cfg)
	})
	public.Post("/auth/logout", func(c *fiber.Ctx) error {
		return handlers.Logout(c)
	})

	protected := api.Group("", middleware.AuthRequired())
	protected.Get("/auth/me", func(c *fiber.Ctx) error {
		return handlers.GetMe(c)
	})
	protected.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	settingsGroup := protected.Group("/settings")
	settingsGroup.Get("", func(c *fiber.Ctx) error {
		return handlers.GetSettings(c)
	})
	settingsGroup.Put("", func(c *fiber.Ctx) error {
		return handlers.SaveSettings(c)
	})
	settingsGroup.Get("/themes", func(c *fiber.Ctx) error {
		return handlers.GetThemes(c)
	})
	settingsGroup.Get("/themes/custom", func(c *fiber.Ctx) error {
		return handlers.GetCustomThemes(c)
	})
	settingsGroup.Post("/themes/custom", func(c *fiber.Ctx) error {
		return handlers.CreateCustomTheme(c)
	})
	settingsGroup.Put("/themes/custom/:id", func(c *fiber.Ctx) error {
		return handlers.UpdateCustomTheme(c)
	})
	settingsGroup.Delete("/themes/custom/:id", func(c *fiber.Ctx) error {
		return handlers.DeleteCustomTheme(c)
	})

	projectsGroup := protected.Group("/projects")
	projectsGroup.Get("", func(c *fiber.Ctx) error {
		return handlers.ListProjects(c, cfg.ProjectsDir)
	})
	projectsGroup.Post("", func(c *fiber.Ctx) error {
		return handlers.CreateProject(c)
	})
	projectsGroup.Post("/scan", func(c *fiber.Ctx) error {
		return handlers.ScanProjects(c, cfg.ProjectsDir)
	})
	projectsGroup.Get("/:id", func(c *fiber.Ctx) error {
		return handlers.GetProject(c)
	})

	fsGroup := protected.Group("/projects/:id/fs")
	fsGroup.Get("/tree", func(c *fiber.Ctx) error {
		return handlers.GetFileTree(c)
	})
	fsGroup.Get("/file", func(c *fiber.Ctx) error {
		return handlers.GetFile(c)
	})
	fsGroup.Put("/file", func(c *fiber.Ctx) error {
		return handlers.PutFile(c)
	})
	fsGroup.Post("/mkdir", func(c *fiber.Ctx) error {
		return handlers.Mkdir(c)
	})
	fsGroup.Post("/rename", func(c *fiber.Ctx) error {
		return handlers.RenameFile(c)
	})
	fsGroup.Delete("/remove", func(c *fiber.Ctx) error {
		return handlers.RemoveFile(c)
	})

	terminalsGroup := protected.Group("/projects/:id/terminals")
	terminalsGroup.Get("", func(c *fiber.Ctx) error {
		return handlers.ListTerminals(c)
	})
	terminalsGroup.Post("", func(c *fiber.Ctx) error {
		return handlers.CreateTerminal(c)
	})

	terminalGroup := protected.Group("/terminals/:tid")
	terminalGroup.Get("", func(c *fiber.Ctx) error {
		return handlers.GetTerminal(c)
	})
	terminalGroup.Post("/resize", func(c *fiber.Ctx) error {
		return handlers.ResizeTerminal(c)
	})
	terminalGroup.Post("/close", func(c *fiber.Ctx) error {
		return handlers.CloseTerminal(c)
	})
	terminalGroup.Get("/ws", func(c *fiber.Ctx) error {
		return handlers.TerminalWS(c)
	})

	gitGroup := protected.Group("/projects/:id/git")
	gitGroup.Get("/status", func(c *fiber.Ctx) error {
		return handlers.GetGitStatus(c)
	})
	gitGroup.Get("/diff", func(c *fiber.Ctx) error {
		return handlers.GetGitDiff(c)
	})
	gitGroup.Post("/stage", func(c *fiber.Ctx) error {
		return handlers.StageFiles(c)
	})
	gitGroup.Post("/unstage", func(c *fiber.Ctx) error {
		return handlers.UnstageFiles(c)
	})
	gitGroup.Post("/commit", func(c *fiber.Ctx) error {
		return handlers.GitCommit(c)
	})
	gitGroup.Post("/push", func(c *fiber.Ctx) error {
		return handlers.GitPush(c)
	})
	gitGroup.Get("/branches", func(c *fiber.Ctx) error {
		return handlers.GetGitBranches(c)
	})
	gitGroup.Get("/log", func(c *fiber.Ctx) error {
		return handlers.GetGitLog(c)
	})

	workspaceGroup := protected.Group("/projects/:id/workspace")
	workspaceGroup.Get("", func(c *fiber.Ctx) error {
		return handlers.GetWorkspace(c)
	})
	workspaceGroup.Put("", func(c *fiber.Ctx) error {
		return handlers.SaveWorkspace(c)
	})

	ai.RegisterRoutes(protected)
	ai.RegisterChatRoutes(protected)
	ai.RegisterUsageRoutes(protected, cfg)
	ai.RegisterWebSocketRoutes(app)
	ai.RegisterChatWSRoutes(protected)
}

func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}
	return c.Status(code).JSON(fiber.Map{
		"error": err.Error(),
	})
}

func bootstrapUser(cfg *config.Config) error {
	var count int
	err := db.GetDB().QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		log.Printf("Users already exist, skipping bootstrap")
		return nil
	}

	email := cfg.BootstrapEmail
	password := cfg.BootstrapPassword

	if email == "" {
		email = "test@example.com"
	}
	if password == "" {
		password = "test123"
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		return err
	}

	userID := uuid.New()
	_, err = db.GetDB().Exec(`
		INSERT INTO users (id, email, password_hash, created_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)`,
		userID, email, hash)
	if err != nil {
		return err
	}

	log.Printf("Bootstrap user %s created", email)
	return nil
}

func newUUID() uuid.UUID {
	return uuid.New()
}
