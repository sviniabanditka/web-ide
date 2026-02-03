# AI Assistant Guidelines

This document provides guidelines for AI assistants working with this codebase.

## Project Overview

WebIDE is a browser-based IDE with AI assistance, built with:
- **Backend**: Go + Fiber + SQLite
- **Frontend**: Vue 3 + TypeScript + Pinia + Monaco Editor

## Key Concepts

### Authentication

- Session-based authentication using cookies
- Token stored in `session_token` cookie
- Protected routes require `AuthRequired` middleware
- Bearer token also supported in Authorization header

### State Management

- **Pinia stores** in `frontend/src/stores/`:
  - `auth.ts` - Authentication state
  - `projects.ts` - Projects list
  - `editor.ts` - Open files, file tree, workspace state
  - `terminals.ts` - Terminal sessions, WebSocket connections
  - `ai.ts` - AI chats, messages, streaming
  - `git.ts` - Git status, branches, commits

### WebSocket Connections

- **Terminal**: One connection per terminal session
- **AI Chat**: One connection per chat session
- Store manages WebSocket lifecycle to prevent duplicates

### Database

- SQLite with auto-migration on startup
- Models in `backend/internal/models/models.go`
- Helpers in `backend/internal/db/helpers.go`

## Common Patterns

### Adding a New API Endpoint

1. Create handler in `backend/internal/api/handlers/`
2. Register route in `backend/cmd/ide-server/main.go`
3. Add API wrapper in `frontend/src/api.ts`
4. Add store action in `frontend/src/stores/`

### Adding a New AI Provider

1. Implement `provider.Provider` interface in `backend/internal/ai/provider/`
2. Add provider factory in `backend/internal/ai/provider/factory.go`
3. Streaming is required for AI responses

### Modifying Database Schema

1. Add model to `backend/internal/models/models.go`
2. Add migration in `backend/internal/db/db.go`
3. Update `Insert`/`Update` helpers if needed

## Running Commands

```bash
# Backend
cd backend && go build -o bin/ide-server ./cmd/ide-server/
./bin/ide-server

# Frontend
cd frontend && npm install
npm run dev      # Development
npm run build    # Production build

# Login (default)
Email: test@example.com
Password: test123
```

## Testing Tips

- Check browser console for WebSocket logs (`[WS]`, `[TERM]` tags)
- Backend logs show request/response details
- Use curl to test API endpoints directly:
  ```bash
  curl -H "Authorization: Bearer <token>" http://localhost:8080/api/v1/...
  ```
