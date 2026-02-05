package agent

import "time"

type AgentMode string

const (
	ModeSafe  AgentMode = "safe"
	ModeWrite AgentMode = "write"
	ModeExec  AgentMode = "exec"
)

type Limits struct {
	MaxSteps         int
	MaxToolTimeMs    time.Duration
	MaxOutputBytes   int64
	MaxFileBytes     int64
	MaxSearchResults int
	MaxPatchFiles    int
}

type AgentConfig struct {
	Mode         AgentMode
	Limits       Limits
	SystemPrompt string
	ProjectRoot  string
	ChatID       string
}

func DefaultConfig() AgentConfig {
	return AgentConfig{
		Mode: ModeSafe,
		Limits: Limits{
			MaxSteps:         12,
			MaxToolTimeMs:    5 * time.Minute,
			MaxOutputBytes:   1024 * 1024, // 1MB
			MaxFileBytes:     512 * 1024,  // 512KB
			MaxSearchResults: 200,
			MaxPatchFiles:    10,
		},
	}
}
