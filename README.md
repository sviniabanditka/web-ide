# WebIDE

A browser-based IDE with AI assistance, built with Go and Vue.js.

## Features

- **File Manager** - Browse, create, edit files and directories
- **Code Editor** - Monaco Editor with syntax highlighting
- **Terminal** - Multiple terminal sessions with xterm.js
- **Git Integration** - Basic git operations (status, diff, commit, branches)
- **AI Chat** - AI-powered coding assistant with streaming responses
- **Real-time Collaboration** - WebSocket-based terminal sessions

## Tech Stack

### Backend
- **Go** with [Fiber](https://gofiber.io/) web framework
- **SQLite** database
- **WebSocket** for real-time communication
- **MiniMax API** for AI streaming responses

### Frontend
- **Vue 3** with Composition API
- **Pinia** for state management
- **TypeScript**
- **Monaco Editor** - VS Code's editor
- **xterm.js** - Terminal emulator

## Project Structure

```
webide/
├── backend/                  # Go backend
│   ├── cmd/ide-server/      # Application entry point
│   └── internal/
│       ├── ai/              # AI module (chat, streaming, providers)
│       ├── api/handlers/   # REST API handlers
│       ├── auth/           # Authentication
│       ├── config/          # Configuration
│       ├── db/              # Database (SQLite)
│       ├── files/           # File operations
│       ├── git/             # Git operations
│       ├── middleware/      # Auth middleware
│       ├── models/          # Data models
│       ├── projects/        # Project management
│       └── terminal/         # Terminal sessions
│
└── frontend/               # Vue 3 frontend
    └── src/
        ├── api.ts           # Axios configuration
        ├── stores/          # Pinia stores
        ├── components/      # Vue components
        └── pages/           # Page components
```

## Installation

### Prerequisites

- Go 1.21+
- Node.js 18+
- npm or yarn

### Backend Setup

```bash
cd backend

# Build the server
go build -o bin/ide-server ./cmd/ide-server/

# Or run directly
go run ./cmd/ide-server/
```

### Frontend Setup

```bash
cd frontend

# Install dependencies
npm install

# Development server
npm run dev

# Production build
npm run build
```

## Configuration

### Environment Variables

Create a `.env` file in the project root (see `.env.example`):

| Variable | Description | Default |
|----------|-------------|---------|
| `IDE_DATA_DIR` | Database directory | `/data` |
| `IDE_PROJECTS_DIR` | Projects directory | `/projects` |
| `IDE_HTTP_ADDR` | HTTP server address | `:8080` |
| `IDE_SESSION_TTL_HOURS` | Session lifetime (hours) | 168 (7 days) |
| `IDE_ALLOW_PROJECTS_SCAN` | Allow scanning for projects | `true` |
| `IDE_MINIMAX_API_KEY` | MiniMax API key for AI | - |
| `IDE_MINIMAX_MODEL` | AI model name | `abab6.5s-chat` |
| `IDE_MINIMAX_URL` | MiniMax API URL (optional) | `https://api.minimax.chat/v1/text/chatcompletion_v2` |
| `IDE_USER_BOOTSTRAP_EMAIL` | Default user email | - |
| `IDE_USER_BOOTSTRAP_PASSWORD` | Default user password | - |

### .env File

```bash
# Copy example to .env
cp .env.example .env

# Edit .env with your settings
nano .env
```

### MiniMax URL Configuration

If you're using MiniMax Coding Plan or a custom endpoint:

```bash
IDE_MINIMAX_URL=https://your-custom-endpoint.com/v1/text/chatcompletion_v2
```

### Default Credentials

When no users exist, a default user is created:
- **Email**: `test@example.com`
- **Password**: `test123`

## API Reference

### Authentication

```
POST /api/v1/auth/login
POST /api/v1/auth/logout
GET  /api/v1/auth/me
```

### Projects

```
GET  /api/v1/projects              # List projects
POST /api/v1/projects             # Create project
GET  /api/v1/projects/:id         # Get project
POST /api/v1/projects/scan        # Scan for projects
```

### File System

```
GET  /api/v1/projects/:id/fs/tree         # File tree
GET  /api/v1/projects/:id/fs/file         # Read file
PUT  /api/v1/projects/:id/fs/file         # Write file
POST /api/v1/projects/:id/fs/mkdir        # Create directory
POST /api/v1/projects/:id/fs/rename      # Rename
DELETE /api/v1/projects/:id/fs/remove     # Delete
```

### Terminals

```
GET  /api/v1/projects/:id/terminals       # List terminals
POST /api/v1/projects/:id/terminals        # Create terminal
GET  /api/v1/terminals/:id                # Get terminal
POST /api/v1/terminals/:id/resize         # Resize
POST /api/v1/terminals/:id/close          # Close
WS   /api/v1/terminals/:id/ws             # WebSocket
```

### Git

```
GET  /api/v1/projects/:id/git/status      # Status
GET  /api/v1/projects/:id/git/diff         # Diff
POST /api/v1/projects/:id/git/stage        # Stage
POST /api/v1/projects/:id/git/unstage      # Unstage
POST /api/v1/projects/:id/git/commit       # Commit
POST /api/v1/projects/:id/git/push          # Push
GET  /api/v1/projects/:id/git/branches     # Branches
GET  /api/v1/projects/:id/git/log          # Log
```

### AI Chat

```
GET  /api/v1/projects/:id/ai/chats              # List chats
POST /api/v1/projects/:id/ai/chats               # Create chat
GET  /api/v1/projects/:id/ai/chats/:chatId      # Get chat
DELETE /api/v1/projects/:id/ai/chats/:chatId    # Delete chat
GET  /api/v1/projects/:id/ai/chats/:chatId/messages      # Messages
POST /api/v1/projects/:id/ai/chats/:chatId/messages       # Send message
GET  /api/v1/projects/:id/ai/chats/:chatId/changesets     # Changesets
WS   /api/v1/ai/chats/:chatId                   # Chat WebSocket
```

## WebSocket Protocol

### Terminal WebSocket

Connect to: `ws://host/api/v1/terminals/:id/ws?token=:session_token`

Send:
```json
{"type": "stdin", "data": "ls -la\n"}
{"type": "resize", "cols": 80, "rows": 24}
{"type": "ping"}
```

Receive: Raw terminal output as text

### AI Chat WebSocket

Connect to: `ws://host/api/v1/ai/chats/:chatId?token=:session_token`

Send:
```json
{"type": "send_message", "payload": {"content": "Hello AI!"}}
{"type": "stop"}
```

Receive streaming chunks:
```json
{"type": "chunk", "payload": {"message_id": "...", "content": "H", "done": false}}
{"type": "chunk", "payload": {"message_id": "...", "content": "i", "done": false}}
{"type": "chunk", "payload": {"message_id": "...", "content": "", "done": true}}
```

Complete message:
```json
{"type": "message_created", "payload": {"id": "...", "role": "assistant", "content": "Hi!"}}
```

## Database Schema

### Main Tables

- `users` - User accounts
- `sessions` - Authentication sessions
- `projects` - Projects
- `terminal_sessions` - Terminal sessions
- `jobs` - AI tasks/jobs
- `changesets` - Code changes
- `chats` - AI chat sessions
- `chat_messages` - Chat messages
- `chat_changesets` - Changes from chat
- `review_threads` - Code review threads
- `review_comments` - Review comments
- `workspace_state` - UI state per user/project

## Development

### Adding a New API Endpoint

1. Create handler in `internal/api/handlers/`
2. Register route in `cmd/ide-server/main.go`
3. Add API function in `frontend/src/api.ts`
4. Add store action in `frontend/src/stores/`

### Adding a New AI Provider

1. Implement `provider.Provider` interface in `internal/ai/provider/`
2. Register in `internal/ai/provider/factory.go`

## License

MIT
