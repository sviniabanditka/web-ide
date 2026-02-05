package config

import (
	"log"
	"os"
	"path/filepath"
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
	MiniMaxURL        string
}

func init() {
	loadEnvFile()
}

func loadEnvFile() {
	wd, err := os.Getwd()
	if err != nil {
		return
	}

	envPath := filepath.Join(wd, ".env")
	data, err := os.ReadFile(envPath)
	if err != nil {
		return
	}

	lines := string(data)
	for _, line := range splitLines(lines) {
		line = trimSpace(line)
		if len(line) == 0 || startsWith(line, "#") {
			continue
		}

		parts := split2(line, "=")
		if len(parts) == 2 {
			key := trimSpace(parts[0])
			value := trimSpace(parts[1])

			if _, exists := os.LookupEnv(key); !exists {
				os.Setenv(key, value)
			}
		}
	}
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' || s[i] == '\r' {
			lines = append(lines, s[start:i])
			start = i + 1
			if i+1 < len(s) && s[i] == '\r' && s[i+1] == '\n' {
				i++
			}
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func split2(s string, sep string) []string {
	idx := indexOf(s, sep)
	if idx == -1 {
		return []string{s}
	}
	return []string{s[:idx], s[idx+len(sep):]}
}

func indexOf(s string, sep string) int {
	for i := 0; i <= len(s)-len(sep); i++ {
		if s[i:i+len(sep)] == sep {
			return i
		}
	}
	return -1
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}

func startsWith(s string, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
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
	miniMaxURL := os.Getenv("IDE_MINIMAX_URL")

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
		MiniMaxURL:        miniMaxURL,
	}, nil
}

func PrintEnvVars() {
	envVars := []string{
		"IDE_DATA_DIR",
		"IDE_PROJECTS_DIR",
		"IDE_HTTP_ADDR",
		"IDE_SESSION_TTL_HOURS",
		"IDE_USER_BOOTSTRAP_EMAIL",
		"IDE_MINIMAX_API_KEY",
		"IDE_MINIMAX_MODEL",
		"IDE_MINIMAX_URL",
	}

	log.Println("=== Loaded Environment Variables ===")
	for _, key := range envVars {
		value := os.Getenv(key)
		if value != "" {
			log.Printf("  %s=%s", key, value)
		} else {
			log.Printf("  %s=(not set)", key)
		}
	}
	log.Println("====================================")
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
