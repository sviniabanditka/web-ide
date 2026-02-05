package builtin

import (
	"bufio"
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/webide/ide/backend/internal/ai/tools"
)

type CommandManager struct {
	mu    sync.RWMutex
	procs map[string]*TrackedProcess
}

type TrackedProcess struct {
	Cmd       *exec.Cmd
	Handle    string
	StartedAt time.Time
	Output    *OutputBuffer
	Done      bool
	ExitCode  int
	mu        sync.Mutex
}

type OutputBuffer struct {
	mu      sync.Mutex
	entries []OutputEntry
	maxSize int
}

type OutputEntry struct {
	Stream string `json:"stream"`
	Text   string `json:"text"`
	TS     int64  `json:"ts"`
}

var CmdManager *CommandManager

func init() {
	CmdManager = &CommandManager{
		procs: make(map[string]*TrackedProcess),
	}
}

func RunCommand() tools.Tool {
	return tools.Tool{
		Name:        "run_command",
		Description: "Execute a shell command with optional streaming output",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"cmd": map[string]interface{}{
					"type": "string",
				},
				"cwd": map[string]interface{}{
					"type":    "string",
					"default": ".",
				},
				"timeout_ms": map[string]interface{}{
					"type":    "integer",
					"default": 600000,
					"minimum": 1000,
					"maximum": 1800000,
				},
				"env": map[string]interface{}{
					"type": "object",
				},
				"stream": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
			},
			"required": []string{"cmd"},
		},
		Policy: tools.PolicyConfirm,
		Execute: func(ctx context.Context, args map[string]interface{}, tc tools.ToolContext) (tools.ToolResult, error) {
			startTime := time.Now()

			cmdStr, _ := args["cmd"].(string)
			if cmdStr == "" {
				return tools.NewErrorResult(tools.ErrCodeValidation, "cmd is required", nil), nil
			}

			cwd := "."
			if c, ok := args["cwd"].(string); ok && c != "" {
				cwd = c
			}

			absCwd, err := filepath.Abs(filepath.Join(tc.ProjectRoot, cwd))
			if err != nil {
				return tools.NewErrorResult(tools.ErrCodeInvalidPath, "invalid cwd", nil), nil
			}

			if !strings.HasPrefix(absCwd, tc.ProjectRoot) {
				return tools.NewErrorResult(tools.ErrCodePermission, "cwd outside project", nil), nil
			}

			timeout := 600000
			if t, ok := args["timeout_ms"].(float64); ok {
				timeout = int(t)
			}
			if timeout > 1800000 {
				timeout = 1800000
			}

			env := os.Environ()
			if e, ok := args["env"].(map[string]interface{}); ok {
				for k, v := range e {
					if s, ok := v.(string); ok {
						env = append(env, k+"="+s)
					}
				}
			}

			stream := true
			if s, ok := args["stream"].(bool); ok {
				stream = s
			}

			handle := uuid.New().String()

			cmd := exec.CommandContext(ctx, "sh", "-c", cmdStr)
			cmd.Dir = absCwd
			cmd.Env = env

			outputBuf := &OutputBuffer{
				maxSize: int(tc.Limits.MaxOutputBytes),
			}

			tracked := &TrackedProcess{
				Cmd:       cmd,
				Handle:    handle,
				StartedAt: time.Now(),
				Output:    outputBuf,
				Done:      false,
			}

			CmdManager.mu.Lock()
			CmdManager.procs[handle] = tracked
			CmdManager.mu.Unlock()

			stdout, err := cmd.StdoutPipe()
			if err != nil {
				return tools.NewErrorResult(tools.ErrCodeExecution, "stdout pipe error", nil), nil
			}

			stderr, err := cmd.StderrPipe()
			if err != nil {
				return tools.NewErrorResult(tools.ErrCodeExecution, "stderr pipe error", nil), nil
			}

			err = cmd.Start()
			if err != nil {
				return tools.NewErrorResult(tools.ErrCodeExecution, "command start error", err.Error()), nil
			}

			if stream {
				go streamOutput(stdout, outputBuf, "stdout", int(tc.Limits.MaxOutputBytes))
				go streamOutput(stderr, outputBuf, "stderr", int(tc.Limits.MaxOutputBytes))
			}

			done := make(chan error, 1)
			go func() {
				done <- cmd.Wait()
			}()

			select {
			case <-ctx.Done():
				cmd.Cancel()
				tracked.mu.Lock()
				tracked.Done = true
				tracked.ExitCode = -1
				tracked.mu.Unlock()
				return tools.NewErrorResult(tools.ErrCodeTimeout, "command cancelled", nil), nil
			case err = <-done:
				tracked.mu.Lock()
				tracked.Done = true
				if err != nil {
					if exitErr, ok := err.(*exec.ExitError); ok {
						tracked.ExitCode = exitErr.ExitCode()
					} else {
						tracked.ExitCode = -1
					}
				} else {
					tracked.ExitCode = 0
				}
				tracked.mu.Unlock()
			}

			CmdManager.mu.Lock()
			delete(CmdManager.procs, handle)
			CmdManager.mu.Unlock()

			return tools.ToolResult{
				OK: true,
				Data: map[string]interface{}{
					"handle":    handle,
					"started":   tracked.StartedAt.Unix(),
					"cwd":       absCwd,
					"exit_code": tracked.ExitCode,
				},
				Meta: &tools.ResultMeta{
					DurationMs: time.Since(startTime).Milliseconds(),
				},
			}, nil
		},
	}
}

func streamOutput(rd io.ReadCloser, buf *OutputBuffer, stream string, maxBytes int) {
	scanner := bufio.NewScanner(rd)
	for scanner.Scan() {
		text := scanner.Text()
		buf.mu.Lock()
		if buf.entries == nil {
			buf.entries = make([]OutputEntry, 0)
		}
		totalSize := 0
		for _, e := range buf.entries {
			totalSize += len(e.Text)
		}
		if totalSize+len(text) > buf.maxSize {
			buf.entries = buf.entries[1:]
		}
		buf.entries = append(buf.entries, OutputEntry{
			Stream: stream,
			Text:   text,
			TS:     time.Now().UnixMilli(),
		})
		buf.mu.Unlock()
	}
	rd.Close()
}

func GetCommandOutput() tools.Tool {
	return tools.Tool{
		Name:        "get_command_output",
		Description: "Get buffered output from a running or completed command",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"handle": map[string]interface{}{
					"type": "string",
				},
				"from": map[string]interface{}{
					"type":    "integer",
					"default": 0,
				},
				"limit": map[string]interface{}{
					"type":    "integer",
					"default": 200,
				},
			},
			"required": []string{"handle"},
		},
		Policy: tools.PolicyAllow,
		Execute: func(ctx context.Context, args map[string]interface{}, tc tools.ToolContext) (tools.ToolResult, error) {
			handle, _ := args["handle"].(string)
			if handle == "" {
				return tools.NewErrorResult(tools.ErrCodeValidation, "handle is required", nil), nil
			}

			CmdManager.mu.RLock()
			tracked, ok := CmdManager.procs[handle]
			CmdManager.mu.RUnlock()

			if !ok {
				return tools.NewErrorResult(tools.ErrCodeNotFound, "command not found", nil), nil
			}

			from := 0
			if f, ok := args["from"].(float64); ok {
				from = int(f)
			}

			limit := 200
			if l, ok := args["limit"].(float64); ok {
				limit = int(l)
			}

			tracked.mu.Lock()
			entries := tracked.Output.entries
			done := tracked.Done
			exitCode := tracked.ExitCode
			tracked.mu.Unlock()

			end := from + limit
			if end > len(entries) {
				end = len(entries)
			}

			var resultEntries []OutputEntry
			if from < len(entries) {
				if end > len(entries) {
					end = len(entries)
				}
				resultEntries = entries[from:end]
			}

			return tools.ToolResult{
				OK: true,
				Data: map[string]interface{}{
					"lines":     resultEntries,
					"next":      end,
					"done":      done,
					"exit_code": exitCode,
				},
				Meta: &tools.ResultMeta{},
			}, nil
		},
	}
}

func CancelCommand() tools.Tool {
	return tools.Tool{
		Name:        "cancel_command",
		Description: "Cancel a running command",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"handle": map[string]interface{}{
					"type": "string",
				},
			},
			"required": []string{"handle"},
		},
		Policy: tools.PolicyAllow,
		Execute: func(ctx context.Context, args map[string]interface{}, tc tools.ToolContext) (tools.ToolResult, error) {
			handle, _ := args["handle"].(string)
			if handle == "" {
				return tools.NewErrorResult(tools.ErrCodeValidation, "handle is required", nil), nil
			}

			CmdManager.mu.RLock()
			tracked, ok := CmdManager.procs[handle]
			CmdManager.mu.RUnlock()

			if !ok {
				return tools.NewErrorResult(tools.ErrCodeNotFound, "command not found", nil), nil
			}

			if tracked.Cmd.Process != nil {
				tracked.Cmd.Process.Kill()
			}

			tracked.mu.Lock()
			tracked.Done = true
			tracked.ExitCode = -1
			tracked.mu.Unlock()

			return tools.ToolResult{
				OK:   true,
				Data: map[string]interface{}{"cancelled": true},
			}, nil
		},
	}
}
