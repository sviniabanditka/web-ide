package agent

import (
	"strings"
)

type PolicyDecision string

const (
	DecisionAllow   PolicyDecision = "allow"
	DecisionConfirm PolicyDecision = "confirm"
	DecisionDeny    PolicyDecision = "deny"
)

type ToolPolicy struct {
	Name      string
	ToolName  string
	Condition func(session *AgentSession, args map[string]interface{}) PolicyDecision
}

type PolicyEngine struct {
	policies []ToolPolicy
}

func NewPolicyEngine() *PolicyEngine {
	return &PolicyEngine{
		policies: []ToolPolicy{
			{
				Name:     "read_file_default",
				ToolName: "read_file",
				Condition: func(s *AgentSession, args map[string]interface{}) PolicyDecision {
					return DecisionAllow
				},
			},
			{
				Name:     "list_dir_default",
				ToolName: "list_dir",
				Condition: func(s *AgentSession, args map[string]interface{}) PolicyDecision {
					return DecisionAllow
				},
			},
			{
				Name:     "search_in_files_default",
				ToolName: "search_in_files",
				Condition: func(s *AgentSession, args map[string]interface{}) PolicyDecision {
					return DecisionAllow
				},
			},
			{
				Name:     "apply_patch_default",
				ToolName: "apply_patch",
				Condition: func(s *AgentSession, args map[string]interface{}) PolicyDecision {
					return DecisionConfirm
				},
			},
			{
				Name:     "run_command_default",
				ToolName: "run_command",
				Condition: func(s *AgentSession, args map[string]interface{}) PolicyDecision {
					cmd, ok := args["cmd"].(string)
					if !ok {
						return DecisionDeny
					}
					if isDangerousCommand(cmd) {
						return DecisionConfirm
					}
					return DecisionConfirm
				},
			},
		},
	}
}

func isDangerousCommand(cmd string) bool {
	lower := strings.ToLower(cmd)
	dangerousPatterns := []string{
		"rm -rf",
		"rm /",
		"mkfs",
		"dd if=",
		":(){:|:&};:",
		"chmod 777",
		"chown",
		"curl ",
		"wget ",
		"> /dev/",
	}
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}
	return false
}

func (e *PolicyEngine) Decide(toolName string, session *AgentSession, args map[string]interface{}) PolicyDecision {
	for _, p := range e.policies {
		if p.ToolName == toolName {
			return p.Condition(session, args)
		}
	}
	return DecisionConfirm
}

func (e *PolicyEngine) AddPolicy(policy ToolPolicy) {
	e.policies = append(e.policies, policy)
}

func GenerateToolSummary(toolName string, args map[string]interface{}) string {
	switch toolName {
	case "read_file":
		path, _ := args["path"].(string)
		return "Read file: " + path
	case "list_dir":
		path, _ := args["path"].(string)
		if path == "" || path == "." {
			return "List directory contents"
		}
		return "List directory: " + path
	case "search_in_files":
		query, _ := args["query"].(string)
		return "Search for: " + query
	case "apply_patch":
		return "Apply code changes"
	case "run_command":
		cmd, _ := args["cmd"].(string)
		return "Run command: " + truncateString(cmd, 50)
	}
	return "Tool: " + toolName
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
