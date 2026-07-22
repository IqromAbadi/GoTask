# GoTask — REST API Task Management

A production-ready task management REST API built with Go, featuring workflow-based task tracking similar to mini-Jira.

## Features

- 🔐 **JWT Authentication** — Register, login, refresh token rotation, logout
- 📋 **Task Lists** — Create and manage multiple task boards
- ✅ **Task Management** — Full CRUD with filtering, pagination, search
- 🔄 **Workflow Engine** — Backlog → Todo → In Progress → Review → Done
- 📊 **Progress Tracking** — Track task completion percentage with notes
- 🔍 **Review System** — Submit for review, approve, or request changes
- 💬 **Comments** — Add and manage comments on tasks
- 📜 **Activity Log** — Automatic audit trail for all task changes
- 📈 **Dashboard** — Summary, progress analytics, deadlines, priority distribution

## Technology Stack

| Component        | Technology              |
| ---------------- | ----------------------- |
| Language         | Go 1.24                 |
| HTTP Router      | go-chi/chi v5           |
| Database         | PostgreSQL 16           |
| Migrations       | golang-migrate          |
| Query Generator  | sqlc                    |
| Validation       | go-playground/validator |
| Authentication   | JWT (golang-jwt)        |
| Password Hashing | bcrypt (x/crypto)       |
| Logging          | log/slog (structured)   |
| UUID             | google/uuid             |
| Containerization | Docker + Docker Compose |

## Project Structure

```
gotask/
├── cmd/api/main.go              # Application entry point
├── internal/
│   ├── activity/                # Activity log module
│   │   ├── handler.go, service.go, repository.go, repository_impl.go, model.go
│   ├── auth/                    # Authentication module
│   │   ├── handler.go, service.go, repository.go, repository_impl.go
│   │   ├── model.go, dto.go, token.go, password.go
│   ├── comment/                 # Comment module
│   │   ├── handler.go, repository_impl.go (includes Service), model.go
│   ├── dashboard/               # Dashboard analytics module
│   │   ├── handler.go, service.go
│   ├── middleware/               # HTTP middleware
│   │   ├── auth.go, logging.go, recovery.go, request_id.go
│   │   ├── cors.go, content_type.go, body_limit.go, rate_limiter.go
│   ├── platform/                # Platform utilities
│   │   ├── config/config.go
│   │   ├── database/postgres.go
│   │   ├── logger/logger.go
│   │   ├── response/response.go
│   │   └── validator/validator.go
│   ├── progress/                # Progress tracking module
│   │   ├── handler.go, service.go, repository.go, repository_impl.go
│   │   ├── model.go, dto.go
│   ├── review/                  # Review module
│   │   ├── handler.go, service.go, repository.go, repository_impl.go, model.go
│   ├── task/                    # Task module
│   │   ├── handler.go, service.go, repository.go, repository_impl.go
│   │   ├── model.go, dto.go
│   └── tasklist/                # Task List module
│       ├── handler.go, service.go, repository.go, repository_impl.go
│       ├── model.go, dto.go
├── db/
│   ├── migrations/              # PostgreSQL migrations
│   │   └── 000001_init_schema.{up,down}.sql
│   ├── queries/                 # SQL queries for sqlc
│   │   ├── activity.sql, auth.sql, comment.sql
│   │   ├── dashboard.sql, health.sql, progress.sql
│   │   ├── review.sql, task.sql, task_list.sql
│   └── sqlc.yaml                # sqlc configuration
├── Dockerfile                   # Multi-stage Docker build
├── docker-compose.yml           # Docker Compose setup
├── Makefile                     # Development commands
├── .env.example                 # Environment variables template
└── README.md
```

## Setup

### Prerequisites

- Go 1.24+
- PostgreSQL 16+
- Docker & Docker Compose (optional)
- [golang-migrate](https://github.com/golang-migrate/migrate) CLI
- [sqlc](https://sqlc.dev/) CLI

### Environment Variables

Copy `.env.example` to `.env`:

```bash
cp .env.example .env
```

Configure:

| Variable               | Description                          | Default                                                          |
| ---------------------- | ------------------------------------ | ---------------------------------------------------------------- |
| `APP_ENV`              | Environment (development/production) | `development`                                                    |
| `APP_PORT`             | HTTP port                            | `8080`                                                           |
| `DATABASE_URL`         | PostgreSQL connection URL            | `postgres://gotask:gotask@localhost:5432/gotask?sslmode=disable` |
| `JWT_ACCESS_SECRET`    | JWT access token secret              | `change-me`                                                      |
| `JWT_REFRESH_SECRET`   | JWT refresh token secret             | `change-me`                                                      |
| `JWT_ACCESS_TTL`       | Access token TTL                     | `15m`                                                            |
| `JWT_REFRESH_TTL`      | Refresh token TTL                    | `720h`                                                           |
| `CORS_ALLOWED_ORIGINS` | Allowed CORS origins                 | `http://localhost:3000`                                          |

### Running with Docker Compose (Recommended)

```bash
# Start PostgreSQL and API
make docker-up

# Check health
curl http://localhost:8080/health

# Stop
make docker-down
```

### Running Locally

```bash
# Start PostgreSQL (via Docker or locally)
docker compose up -d postgres

# Run migrations
make migrate-up

# Run the application
make run
```

### Generate sqlc

```bash
make sqlc
```

## API Endpoints

### Health

| Method | Path      | Auth   | Description  |
| ------ | --------- | ------ | ------------ |
| GET    | `/health` | Public | Health check |

### Authentication

| Method | Path                    | Auth   | Description          |
| ------ | ----------------------- | ------ | -------------------- |
| POST   | `/api/v1/auth/register` | Public | Register new user    |
| POST   | `/api/v1/auth/login`    | Public | Login                |
| POST   | `/api/v1/auth/refresh`  | Public | Refresh access token |
| POST   | `/api/v1/auth/logout`   | Public | Logout               |

### User

| Method | Path                        | Auth     | Description     |
| ------ | --------------------------- | -------- | --------------- |
| GET    | `/api/v1/users/me`          | Required | Get profile     |
| PATCH  | `/api/v1/users/me`          | Required | Update profile  |
| PATCH  | `/api/v1/users/me/password` | Required | Change password |

### Task Lists

| Method | Path                             | Auth     | Description       |
| ------ | -------------------------------- | -------- | ----------------- |
| POST   | `/api/v1/lists`                  | Required | Create task list  |
| GET    | `/api/v1/lists`                  | Required | List task lists   |
| GET    | `/api/v1/lists/{listId}`         | Required | Get task list     |
| PATCH  | `/api/v1/lists/{listId}`         | Required | Update task list  |
| DELETE | `/api/v1/lists/{listId}`         | Required | Delete task list  |
| PATCH  | `/api/v1/lists/{listId}/archive` | Required | Archive task list |
| PATCH  | `/api/v1/lists/{listId}/restore` | Required | Restore task list |
| GET    | `/api/v1/lists/{listId}/board`   | Required | Get board view    |

### Tasks

| Method | Path                              | Auth     | Description               |
| ------ | --------------------------------- | -------- | ------------------------- |
| POST   | `/api/v1/lists/{listId}/tasks`    | Required | Create task               |
| GET    | `/api/v1/lists/{listId}/tasks`    | Required | List tasks (with filters) |
| GET    | `/api/v1/tasks/{taskId}`          | Required | Get task                  |
| PATCH  | `/api/v1/tasks/{taskId}`          | Required | Update task               |
| DELETE | `/api/v1/tasks/{taskId}`          | Required | Soft delete task          |
| PATCH  | `/api/v1/tasks/{taskId}/status`   | Required | Change task status        |
| PATCH  | `/api/v1/tasks/{taskId}/priority` | Required | Change task priority      |
| POST   | `/api/v1/tasks/{taskId}/reopen`   | Required | Reopen completed task     |

**Task Filters**: `?status=in_progress&priority=high&search=login&due_date_from=2026-07-01&due_date_to=2026-07-31&is_overdue=true&sort_by=due_date&sort_order=asc&page=1&limit=20`

### Progress

| Method | Path                                           | Auth     | Description            |
| ------ | ---------------------------------------------- | -------- | ---------------------- |
| POST   | `/api/v1/tasks/{taskId}/progress`              | Required | Add progress update    |
| GET    | `/api/v1/tasks/{taskId}/progress`              | Required | List progress updates  |
| GET    | `/api/v1/tasks/{taskId}/progress/{progressId}` | Required | Get progress update    |
| PATCH  | `/api/v1/tasks/{taskId}/progress/{progressId}` | Required | Update progress note   |
| DELETE | `/api/v1/tasks/{taskId}/progress/{progressId}` | Required | Delete progress update |

### Review

| Method | Path                                                        | Auth     | Description       |
| ------ | ----------------------------------------------------------- | -------- | ----------------- |
| POST   | `/api/v1/tasks/{taskId}/submit-review`                      | Required | Submit for review |
| GET    | `/api/v1/tasks/{taskId}/reviews`                            | Required | List reviews      |
| GET    | `/api/v1/tasks/{taskId}/reviews/{reviewId}`                 | Required | Get review        |
| POST   | `/api/v1/tasks/{taskId}/reviews/{reviewId}/approve`         | Required | Approve review    |
| POST   | `/api/v1/tasks/{taskId}/reviews/{reviewId}/request-changes` | Required | Request changes   |

### Comments

| Method | Path                                          | Auth     | Description    |
| ------ | --------------------------------------------- | -------- | -------------- |
| POST   | `/api/v1/tasks/{taskId}/comments`             | Required | Add comment    |
| GET    | `/api/v1/tasks/{taskId}/comments`             | Required | List comments  |
| PATCH  | `/api/v1/tasks/{taskId}/comments/{commentId}` | Required | Update comment |
| DELETE | `/api/v1/tasks/{taskId}/comments/{commentId}` | Required | Delete comment |

### Activity

| Method | Path                                | Auth     | Description     |
| ------ | ----------------------------------- | -------- | --------------- |
| GET    | `/api/v1/tasks/{taskId}/activities` | Required | Task activities |
| GET    | `/api/v1/activities`                | Required | User activities |

### Dashboard

| Method | Path                                      | Auth     | Description           |
| ------ | ----------------------------------------- | -------- | --------------------- |
| GET    | `/api/v1/dashboard/summary`               | Required | Task summary          |
| GET    | `/api/v1/dashboard/progress`              | Required | Progress analytics    |
| GET    | `/api/v1/dashboard/upcoming-deadlines`    | Required | Upcoming deadlines    |
| GET    | `/api/v1/dashboard/overdue-tasks`         | Required | Overdue tasks         |
| GET    | `/api/v1/dashboard/priority-distribution` | Required | Priority distribution |

## API Examples

### Register

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Agung","email":"agung@example.com","password":"Password123!"}'
```

### Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"agung@example.com","password":"Password123!"}'
```

Response:

```json
{
  "success": true,
  "message": "Login berhasil",
  "data": {
    "access_token": "eyJhbGci...",
    "refresh_token": "dGhpcyBp...",
    "expires_in": 900,
    "token_type": "Bearer"
  }
}
```

### Create Task List

```bash
curl -X POST http://localhost:8080/api/v1/lists \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <access_token>" \
  -d '{"name":"Sprint 1","description":"First sprint tasks"}'
```

### Create Task

```bash
curl -X POST http://localhost:8080/api/v1/lists/{listId}/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <access_token>" \
  -d '{"title":"Implement login","priority":"high","status":"todo","due_date":"2026-07-30","estimated_minutes":240}'
```

### Add Progress

```bash
curl -X POST http://localhost:8080/api/v1/tasks/{taskId}/progress \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <access_token>" \
  -d '{"progress":60,"note":"Login and register endpoints done"}'
```

### Submit for Review

```bash
curl -X POST http://localhost:8080/api/v1/tasks/{taskId}/submit-review \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <access_token>" \
  -d '{"submission_note":"All auth endpoints completed and tested"}'
```

## Task Workflow

```mermaid
stateDiagram-v2
    [*] --> backlog
    backlog --> todo
    todo --> backlog
    todo --> in_progress
    in_progress --> todo
    in_progress --> review: progress = 100%
    review --> in_progress
    review --> done: approved
    done --> in_progress: reopen
```

### Valid Status Transitions

| From        | To                   |
| ----------- | -------------------- |
| backlog     | todo                 |
| todo        | backlog, in_progress |
| in_progress | todo, review         |
| review      | in_progress, done    |
| done        | in_progress          |

## Business Rules

1. Task can only enter `review` when progress is 100%
2. Task can only reach `done` through review approval
3. Status `done` cannot be set directly via status endpoint
4. Progress can only be added when task is `in_progress`
5. New progress must be ≥ previous progress (unless `allow_rollback: true` with a note)
6. Rollback requires a note explaining the reason
7. All status changes are logged to activity
8. Deleted tasks and comments use soft delete
9. Comments can only be edited/deleted by their owner
10. Each user can only access their own data

## ERD (Entity Relationship Diagram)

```mermaid
erDiagram
    users ||--o{ refresh_tokens : has
    users ||--o{ task_lists : owns
    task_lists ||--o{ tasks : contains
    users ||--o{ tasks : creates
    tasks ||--o{ task_progress_updates : has
    tasks ||--o{ task_reviews : has
    tasks ||--o{ task_comments : has
    tasks ||--o{ task_activities : logs
    users ||--o{ task_progress_updates : records
    users ||--o{ task_reviews : reviews
    users ||--o{ task_comments : writes
    users ||--o{ task_activities : performs

    users {
        uuid id PK
        string name
        string email UK
        string password_hash
        string avatar_url
        timestamp created_at
        timestamp updated_at
    }

    refresh_tokens {
        uuid id PK
        uuid user_id FK
        string token_hash
        timestamp expires_at
        timestamp revoked_at
        timestamp created_at
    }

    task_lists {
        uuid id PK
        uuid user_id FK
        string name
        text description
        boolean is_archived
        timestamp created_at
        timestamp updated_at
    }

    tasks {
        uuid id PK
        uuid list_id FK
        uuid created_by FK
        string title
        text description
        string status
        string priority
        int progress
        date due_date
        int estimated_minutes
        timestamp started_at
        timestamp completed_at
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
    }

    task_progress_updates {
        uuid id PK
        uuid task_id FK
        uuid user_id FK
        int progress
        text note
        timestamp created_at
        timestamp updated_at
    }

    task_reviews {
        uuid id PK
        uuid task_id FK
        uuid reviewer_id FK
        string status
        text submission_note
        text review_note
        timestamp reviewed_at
        timestamp created_at
        timestamp updated_at
    }

    task_comments {
        uuid id PK
        uuid task_id FK
        uuid user_id FK
        text content
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
    }

    task_activities {
        uuid id PK
        uuid task_id FK
        uuid user_id FK
        string action
        text old_value
        text new_value
        jsonb metadata
        timestamp created_at
    }
```

## Architecture

The project follows a modular layered architecture:

```mermaid
graph TD
    HTTP[HTTP Request] --> MW[Middleware Stack]
    MW --> R[Router chi]
    R --> H[Handler]
    H --> S[Service]
    S --> REP[Repository]
    REP --> DB[(PostgreSQL)]

    subgraph Middleware
        MW1[RequestID]
        MW2[Logging]
        MW3[Recovery]
        MW4[CORS]
        MW5[BodyLimit]
        MW6[ContentType]
        MW7[RateLimit]
        MW8[Authenticate]
    end

    subgraph Modules
        AUTH[Auth]
        USER[User]
        TL[TaskList]
        TASK[Task]
        PROG[Progress]
        REV[Review]
        COM[Comment]
        ACT[Activity]
        DASH[Dashboard]
    end
```

**Pattern**: Handler → Service → Repository → Database

Each module is self-contained with:

- `model.go` — Domain entity
- `dto.go` — Request/Response DTOs
- `repository.go` — Repository interface
- `repository_impl.go` — PostgreSQL implementation
- `service.go` — Business logic
- `handler.go` — HTTP handlers

## Security

- ✅ Password hashing with bcrypt
- ✅ JWT signature validation
- ✅ Access token expiration (15 min)
- ✅ Refresh token rotation
- ✅ Refresh token storage as hash (not plain text)
- ✅ Rate limiting on auth endpoints
- ✅ Generic login error (no email enumeration)
- ✅ Parameterized queries (SQL injection prevention)
- ✅ Input validation
- ✅ Ownership authorization (user-scoped queries)
- ✅ Soft delete
- ✅ Request body size limit (1MB)
- ✅ Panic recovery
- ✅ CORS configuration
- ✅ Request ID tracking
- ✅ Structured logging (no passwords/tokens)

## Makefile Commands

```bash
make run              # Run the application
make build            # Build binary
make test             # Run tests
make test-cover       # Run tests with coverage
make migrate-up       # Run database migrations up
make migrate-down     # Run database migrations down
make migrate-create name=xxx  # Create a new migration
make sqlc             # Generate sqlc code
make docker-up        # Start Docker Compose
make docker-down      # Stop Docker Compose
make lint             # Run linter
make fmt              # Format code
```

## Testing

```bash
# Run all tests
make test

# Run with coverage
make test-cover
```

Tests cover:

- Authentication (register, login, refresh, logout)
- Task list CRUD and ownership
- Task CRUD, filtering, and status transitions
- Password validation
- Status workflow rules
- Progress business rules
- User ownership enforcement

## License

MIT
