# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Frontend Development
- `cd frontend && pnpm install` - Install frontend dependencies
- `cd frontend && pnpm run dev` - Start frontend development server
- `cd frontend && pnpm run build` - Build frontend for production
- `cd frontend && pnpm run lint` - Run ESLint on frontend code

### Backend Development
- `cd backend && make deps` - Download Go dependencies (alias for `go mod tidy`)
- `cd backend && make dev` - Start backend development server
- `cd backend && go run main.go` - Alternative way to start backend development server
- `cd backend && go build -o build/xsha .` - Build backend binary

### Full Stack Development
- `make frontend-build` - Build frontend and copy to backend/static
- `make build-embedded-amd64` - Build production binary with embedded frontend
- `make deploy` - Build production deployment binary for AMD64
- `docker compose up -d` - Start full application with Docker

### Testing
- Frontend: Use `pnpm run lint` for code quality checks
- Backend: Run `go test ./...` from the backend directory
- Integration: Test via Docker compose environment

## Architecture Overview

XSha is an AI-driven project management and development platform that enables conversational task development with Claude Code integration. The system consists of:

### Backend Architecture (Go)
- **Framework**: Gin web framework with embedded static file serving
- **Database**: GORM with SQLite (default) or MySQL support
- **Authentication**: JWT-based with middleware protection
- **Task Execution**: Docker-containerized execution environment with concurrent task processing
- **File Management**: Attachment handling with configurable storage

**Key Service Layers**:
- `handlers/` - HTTP request handlers and API endpoints
- `services/` - Business logic layer with dependency injection
- `repository/` - Data access layer with interface abstractions
- `scheduler/` - Task scheduling and execution management
- `services/executor/` - Docker-based task execution with streaming logs
- `middleware/` - Authentication, logging, i18n, and rate limiting
- `database/` - Database connection and model management

### Frontend Architecture (React + TypeScript)
- **Framework**: React 19 with TypeScript and Vite
- **UI Components**: Radix UI primitives with custom styling
- **State Management**: React Context for auth and navigation
- **Styling**: Tailwind CSS with shadcn/ui component system
- **Data Management**: TanStack Table for data grids
- **Internationalization**: i18next with dynamic language switching

**Key Component Structure**:
- `pages/` - Route-level page components
- `components/ui/` - Reusable UI components (shadcn/ui)
- `components/forms/` - Form components with validation
- `components/kanban/` - Task management Kanban interface
- `components/data-table/` - Table components with sorting/filtering
- `hooks/` - Custom React hooks for state and API management
- `lib/api/` - API client layer with TypeScript types

### Key Integrations
- **Docker Integration**: Each task runs in isolated Docker containers
- **Git Integration**: Repository cloning, credential management, and branch operations
- **AI Integration**: Claude Code task execution with conversation history
- **File Attachments**: Upload and management system for task contexts
- **Real-time Logging**: WebSocket-based streaming for task execution logs

### Database Models
Core entities include User, Project, GitCredential, DevEnvironment, Task, TaskConversation, TaskConversationResult, and TaskConversationAttachment with proper foreign key relationships.

### Security Features
- JWT-based authentication with token blacklisting
- Docker container isolation for task execution
- File upload validation and storage limits
- Rate limiting middleware
- Operation logging for audit trails

## Development Notes

### Adding New Features
1. Backend: Create repository interface → implement service → add handler → register route
2. Frontend: Add API types → create hooks → build components → integrate with pages
3. Database: Add models to `database/models.go` and run migrations

### Docker Integration
- Task execution happens in isolated Docker containers
- Workspace management handles Git operations and file systems
- Log streaming provides real-time feedback during task execution

### Internationalization
Both frontend and backend support i18n:
- Frontend: Uses react-i18next with JSON locale files
- Backend: Custom i18n system with JSON locale files
- Supported languages: English (en-US) and Chinese (zh-CN)

### Environment Configuration
Key environment variables for development:
- `XSHA_PORT` - Backend server port (default: 8080)
- `XSHA_DATABASE_TYPE` - Database type (sqlite/mysql)
- `XSHA_JWT_SECRET` - JWT signing secret
- `XSHA_WORKSPACE_BASE_DIR` - Base directory for workspaces
- `XSHA_MAX_CONCURRENT_TASKS` - Maximum concurrent task execution

### File Structure Conventions
- Backend: Package-based organization with clear separation of concerns
- Frontend: Feature-based component organization with shared UI components
- Both codebases follow TypeScript/Go naming conventions and include comprehensive type definitions