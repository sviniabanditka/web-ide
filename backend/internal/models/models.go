package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type Session struct {
	ID         uuid.UUID `json:"id" db:"id"`
	UserID     uuid.UUID `json:"user_id" db:"user_id"`
	Token      string    `json:"-" db:"token"`
	ExpiresAt  time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	LastSeenAt time.Time `json:"last_seen_at" db:"last_seen_at"`
}

type Project struct {
	ID           uuid.UUID `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	RootPath     string    `json:"root_path" db:"root_path"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	LastOpenedAt time.Time `json:"last_opened_at" db:"last_opened_at"`
}

type TerminalSession struct {
	ID             uuid.UUID `json:"id" db:"id"`
	ProjectID      uuid.UUID `json:"project_id" db:"project_id"`
	Title          string    `json:"title" db:"title"`
	Cwd            string    `json:"cwd" db:"cwd"`
	Shell          string    `json:"shell" db:"shell"`
	Status         string    `json:"status" db:"status"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	LastAttachedAt time.Time `json:"last_attached_at" db:"last_attached_at"`
}

type Job struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	ProjectID   uuid.UUID  `json:"project_id" db:"project_id"`
	Type        string     `json:"type" db:"type"`
	Status      string     `json:"status" db:"status"`
	PayloadJSON string     `json:"-" db:"payload_json"`
	ResultJSON  string     `json:"-" db:"result_json"`
	ErrorText   string     `json:"-" db:"error_text"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	StartedAt   *time.Time `json:"started_at" db:"started_at"`
	FinishedAt  *time.Time `json:"finished_at" db:"finished_at"`
}

type ChangeSet struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	ProjectID   uuid.UUID  `json:"project_id" db:"project_id"`
	JobID       *uuid.UUID `json:"job_id" db:"job_id"`
	Title       string     `json:"title" db:"title"`
	BaseRef     string     `json:"base_ref" db:"base_ref"`
	TargetRef   *string    `json:"target_ref" db:"target_ref"`
	ApplyMode   string     `json:"apply_mode" db:"apply_mode"`
	Status      string     `json:"status" db:"status"`
	SummaryText string     `json:"summary_text" db:"summary_text"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

type ReviewThread struct {
	ID          uuid.UUID `json:"id" db:"id"`
	ChangeSetID uuid.UUID `json:"changeset_id" db:"changeset_id"`
	FilePath    string    `json:"file_path" db:"file_path"`
	AnchorJSON  string    `json:"anchor_json" db:"anchor_json"`
	Status      string    `json:"status" db:"status"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type ReviewComment struct {
	ID           uuid.UUID `json:"id" db:"id"`
	ThreadID     uuid.UUID `json:"thread_id" db:"thread_id"`
	AuthorUserID uuid.UUID `json:"author_user_id" db:"author_user_id"`
	Body         string    `json:"body" db:"body"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type Chat struct {
	ID        uuid.UUID `json:"id" db:"id"`
	ProjectID uuid.UUID `json:"project_id" db:"project_id"`
	Title     string    `json:"title" db:"title"`
	Status    string    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type ChatMessage struct {
	ID              uuid.UUID `json:"id" db:"id"`
	ChatID          uuid.UUID `json:"chat_id" db:"chat_id"`
	Role            string    `json:"role" db:"role"`
	Content         string    `json:"content" db:"content"`
	ToolCallsJSON   string    `json:"tool_calls_json,omitempty" db:"tool_calls_json"`
	ToolResultsJSON string    `json:"tool_results_json,omitempty" db:"tool_results_json"`
	Thinking        string    `json:"thinking,omitempty" db:"thinking"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

type ChatChangeSet struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	ChatID      uuid.UUID  `json:"chat_id" db:"chat_id"`
	JobID       *uuid.UUID `json:"job_id" db:"job_id"`
	Title       string     `json:"title" db:"title"`
	Diff        string     `json:"diff" db:"diff"`
	Status      string     `json:"status" db:"status"`
	SummaryText string     `json:"summary_text" db:"summary_text"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}
