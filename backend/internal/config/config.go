package config

import (
	"os"
	"strconv"
)

type Config struct {
	DataDir           string
	ProjectsDir       string
	HTTPAddr          string
	SessionTTLHours   int
	AllowProjectsScan bool
	BootstrapEmail    string
	BootstrapPassword string
	MiniMaxAPIKey     string
	MiniMaxModel      string
}

func Load() (*Config, error) {
	dataDir := getEnv("IDE_DATA_DIR", "/data")
	projectsDir := getEnv("IDE_PROJECTS_DIR", "/projects")
	httpAddr := getEnv("IDE_HTTP_ADDR", ":8080")
	sessionTTL := getEnvInt("IDE_SESSION_TTL_HOURS", 168)
	allowScan := getEnvBool("IDE_ALLOW_PROJECTS_SCAN", true)
	bootstrapEmail := os.Getenv("IDE_USER_BOOTSTRAP_EMAIL")
	bootstrapPassword := os.Getenv("IDE_USER_BOOTSTRAP_PASSWORD")
	miniMaxAPIKey := os.Getenv("IDE_MINIMAX_API_KEY")
	miniMaxModel := getEnv("IDE_MINIMAX_MODEL", "abab6.5s-chat")

	return &Config{
		DataDir:           dataDir,
		ProjectsDir:       projectsDir,
		HTTPAddr:          httpAddr,
		SessionTTLHours:   sessionTTL,
		AllowProjectsScan: allowScan,
		BootstrapEmail:    bootstrapEmail,
		BootstrapPassword: bootstrapPassword,
		MiniMaxAPIKey:     miniMaxAPIKey,
		MiniMaxModel:      miniMaxModel,
	}, nil
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if v := os.Getenv(key); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if v := os.Getenv(key); v != "" {
		return v == "true" || v == "1"
	}
	return defaultValue
}
