package terminal

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/creack/pty"
	"github.com/google/uuid"
	"github.com/webide/ide/backend/internal/db"
)

type TerminalSession struct {
	ID        uuid.UUID
	ProjectID uuid.UUID
	Title     string
	Cwd       string
	Shell     string
	Status    string
	Pty       *os.File
	Cmd       *exec.Cmd
	Buffer    *RingBuffer
	CreatedAt time.Time
	LastSeen  time.Time
	mu        sync.Mutex
}

type TermSize struct {
	Cols int `json:"cols"`
	Rows int `json:"rows"`
}

type WSMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

const bufferSize = 64 * 1024

var sessions = make(map[uuid.UUID]*TerminalSession)
var sessionsMu sync.RWMutex

func CreateSession(projectID uuid.UUID, cwd, title, shell string) (*TerminalSession, error) {
	id := uuid.New()

	if shell == "" {
		shell = "/bin/bash"
	}

	cmd := exec.Command(shell, "-i")
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")
	if cwd != "" {
		cmd.Dir = cwd
	}

	ptyFile, err := pty.Start(cmd)
	if err != nil {
		return nil, err
	}

	session := &TerminalSession{
		ID:        id,
		ProjectID: projectID,
		Title:     title,
		Cwd:       cwd,
		Shell:     shell,
		Status:    "running",
		Pty:       ptyFile,
		Cmd:       cmd,
		Buffer:    NewRingBuffer(bufferSize),
		CreatedAt: time.Now(),
		LastSeen:  time.Now(),
	}

	sessionsMu.Lock()
	sessions[id] = session
	sessionsMu.Unlock()

	_, err = db.GetDB().Exec(`
		INSERT INTO terminal_sessions (id, project_id, title, cwd, shell, status, created_at, last_attached_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		id, projectID, title, cwd, shell, "running", time.Now(), time.Now())
	if err != nil {
		log.Printf("Failed to save terminal session: %v", err)
	}

	go session.readOutput()

	return session, nil
}

func (s *TerminalSession) readOutput() {
	buf := make([]byte, 1024)
	for {
		n, err := s.Pty.Read(buf)
		if err != nil {
			if err == io.EOF {
				s.mu.Lock()
				s.Status = "closed"
				s.mu.Unlock()
			}
			return
		}

		data := buf[:n]
		s.Buffer.Write(data)

		s.mu.Lock()
		s.LastSeen = time.Now()
		s.mu.Unlock()
	}
}

func (s *TerminalSession) Write(data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.Pty.Write(data)
	return err
}

func (s *TerminalSession) Resize(cols, rows int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return pty.Setsize(s.Pty, &pty.Winsize{
		Cols: uint16(cols),
		Rows: uint16(rows),
	})
}

func (s *TerminalSession) GetBacklog() []byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Buffer.ReadAll()
}

func (s *TerminalSession) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Status == "closed" {
		return nil
	}

	s.Status = "closed"
	s.Pty.Close()
	s.Cmd.Process.Kill()
	s.Cmd.Wait()

	delete(sessions, s.ID)

	db.GetDB().Exec("UPDATE terminal_sessions SET status = 'closed' WHERE id = ?", s.ID)

	return nil
}

func GetSession(id uuid.UUID) (*TerminalSession, error) {
	sessionsMu.RLock()
	defer sessionsMu.RUnlock()

	session, ok := sessions[id]
	if !ok {
		return nil, os.ErrNotExist
	}
	return session, nil
}

func GetProjectSessions(projectID uuid.UUID) ([]*TerminalSession, error) {
	sessionsMu.RLock()
	defer sessionsMu.RUnlock()

	var result []*TerminalSession
	for _, s := range sessions {
		if s.ProjectID == projectID {
			result = append(result, s)
		}
	}
	return result, nil
}

type RingBuffer struct {
	data   []byte
	size   int
	start  int
	length int
	mu     sync.Mutex
}

func NewRingBuffer(size int) *RingBuffer {
	return &RingBuffer{
		data: make([]byte, size),
		size: size,
	}
}

func (r *RingBuffer) Write(p []byte) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for len(p) > 0 {
		avail := r.size - r.length
		if avail == 0 {
			r.start = (r.start + 1) % r.size
			r.length--
			avail = r.size - r.length
		}

		toWrite := len(p)
		if toWrite > avail {
			toWrite = avail
		}

		end := (r.start + r.length) % r.size
		if end+toWrite <= r.size {
			copy(r.data[end:end+toWrite], p[:toWrite])
		} else {
			first := r.size - end
			copy(r.data[end:], p[:first])
			copy(r.data[:toWrite-first], p[first:])
		}

		r.length += toWrite
		p = p[toWrite:]
	}
}

func (r *RingBuffer) ReadAll() []byte {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := make([]byte, r.length)
	if r.length == 0 {
		return result
	}

	if r.start+r.length <= r.size {
		copy(result, r.data[r.start:r.start+r.length])
	} else {
		first := r.size - r.start
		copy(result[:first], r.data[r.start:])
		copy(result[first:], r.data[:r.length-first])
	}

	return result
}

func (r *RingBuffer) Read(p []byte) (n int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.length == 0 {
		return 0, io.EOF
	}

	toRead := len(p)
	if toRead > r.length {
		toRead = r.length
	}

	if r.start+toRead <= r.size {
		copy(p, r.data[r.start:r.start+toRead])
	} else {
		first := r.size - r.start
		copy(p[:first], r.data[r.start:])
		copy(p[first:], r.data[:toRead-first])
	}

	r.start = (r.start + toRead) % r.size
	r.length -= toRead

	return toRead, nil
}

func CleanupOldSessions(maxAge time.Duration) {
	sessionsMu.Lock()
	defer sessionsMu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	for _, s := range sessions {
		if s.LastSeen.Before(cutoff) {
			s.Close()
		}
	}
}

func CollectGarbage() {
	for {
		time.Sleep(5 * time.Minute)
		CleanupOldSessions(24 * time.Hour)
	}
}

type TerminalMessage struct {
	Type string `json:"type"`
	Data string `json:"data,omitempty"`
	Cols int    `json:"cols,omitempty"`
	Rows int    `json:"rows,omitempty"`
}

func ParseWSMessage(data []byte) (*TerminalMessage, error) {
	var msg TerminalMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

func FormatWSMessage(msg *TerminalMessage) ([]byte, error) {
	return json.Marshal(msg)
}

func NewStdoutMessage(data []byte) *TerminalMessage {
	return &TerminalMessage{
		Type: "stdout",
		Data: string(data),
	}
}

func NewResizeMessage(cols, rows int) *TerminalMessage {
	return &TerminalMessage{
		Type: "resize",
		Cols: cols,
		Rows: rows,
	}
}

func NewPingMessage() *TerminalMessage {
	return &TerminalMessage{
		Type: "ping",
	}
}

func NewPongMessage() *TerminalMessage {
	return &TerminalMessage{
		Type: "pong",
	}
}

type TerminalOutput struct {
	TerminalID string `json:"terminal_id"`
	Data       string `json:"data"`
}

func FormatTerminalOutput(id uuid.UUID, data []byte) ([]byte, error) {
	output := TerminalOutput{
		TerminalID: id.String(),
		Data:       string(data),
	}
	return json.Marshal(map[string]interface{}{
		"type":    "terminal_output",
		"payload": output,
	})
}

var bufPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}
