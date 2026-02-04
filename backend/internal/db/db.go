package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

const dbFilename = "ide.db"

func Init(dataDir string) error {
	dbPath := filepath.Join(dataDir, dbFilename)

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Printf("Warning: failed to create data dir %s: %v, using temp dir", dataDir, err)
		dataDir = filepath.Join(os.TempDir(), "ide-data")
		dbPath = filepath.Join(dataDir, dbFilename)
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			return fmt.Errorf("failed to create temp data dir: %w", err)
		}
	}

	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	if err := runMigrations(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Printf("Database initialized at %s", dbPath)
	return nil
}

func Close() {
	if db != nil {
		db.Close()
	}
}

func GetDB() *sql.DB {
	return db
}

func runMigrations() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			token TEXT UNIQUE NOT NULL,
			expires_at DATETIME NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_seen_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS projects (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			root_path TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_opened_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS terminal_sessions (
			id TEXT PRIMARY KEY,
			project_id TEXT NOT NULL,
			title TEXT,
			cwd TEXT NOT NULL,
			shell TEXT NOT NULL,
			status TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_attached_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS jobs (
			id TEXT PRIMARY KEY,
			project_id TEXT NOT NULL,
			type TEXT NOT NULL,
			status TEXT NOT NULL,
			payload_json TEXT,
			result_json TEXT,
			error_text TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			started_at DATETIME,
			finished_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS changesets (
			id TEXT PRIMARY KEY,
			project_id TEXT NOT NULL,
			job_id TEXT,
			title TEXT NOT NULL,
			base_ref TEXT NOT NULL,
			target_ref TEXT,
			apply_mode TEXT NOT NULL,
			status TEXT NOT NULL,
			summary_text TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS review_threads (
			id TEXT PRIMARY KEY,
			changeset_id TEXT NOT NULL,
			file_path TEXT NOT NULL,
			anchor_json TEXT NOT NULL,
			status TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS review_comments (
			id TEXT PRIMARY KEY,
			thread_id TEXT NOT NULL,
			author_user_id TEXT NOT NULL,
			body TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(token)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_jobs_project ON jobs(project_id)`,
		`CREATE INDEX IF NOT EXISTS idx_changesets_project ON changesets(project_id)`,
		`CREATE INDEX IF NOT EXISTS idx_review_threads_changeset ON review_threads(changeset_id)`,

		`CREATE TABLE IF NOT EXISTS workspace_state (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			project_id TEXT NOT NULL,
			open_files_json TEXT NOT NULL DEFAULT '[]',
			expanded_dirs_json TEXT NOT NULL DEFAULT '[]',
			active_file TEXT,
			active_tab TEXT DEFAULT 'terminal',
			open_terminals_json TEXT NOT NULL DEFAULT '[]',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, project_id),
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (project_id) REFERENCES projects(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_workspace_state_user ON workspace_state(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_workspace_state_project ON workspace_state(project_id)`,

		`CREATE TABLE IF NOT EXISTS chats (
			id TEXT PRIMARY KEY,
			project_id TEXT NOT NULL,
			title TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'active',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_chats_project ON chats(project_id)`,

		`CREATE TABLE IF NOT EXISTS chat_messages (
			id TEXT PRIMARY KEY,
			chat_id TEXT NOT NULL,
			role TEXT NOT NULL,
			content TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_chat_messages_chat ON chat_messages(chat_id)`,
		`ALTER TABLE chat_messages ADD COLUMN updated_at DATETIME`,

		`CREATE TABLE IF NOT EXISTS chat_changesets (
			id TEXT PRIMARY KEY,
			chat_id TEXT NOT NULL,
			job_id TEXT,
			title TEXT NOT NULL,
			diff TEXT,
			status TEXT NOT NULL DEFAULT 'pending',
			summary_text TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_chat_changesets_chat ON chat_changesets(chat_id)`,
	}

	for _, m := range migrations {
		if _, err := db.Exec(m); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	return nil
}
