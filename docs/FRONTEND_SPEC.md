# GoTask Backend Review & Flutter Frontend Specification

## A. Backend Review

### 1. Endpoint Map (Complete)

| # | Method | Path | Auth | Request Body | Response Key Fields |
|---|--------|------|------|-------------|---------------------|
| 1 | POST | `/api/v1/auth/register` | Public | `name*`, `email*`, `password*` | `id, name, email, avatar_url, created_at` |
| 2 | POST | `/api/v1/auth/login` | Public | `email*`, `password*` | `access_token, refresh_token, expires_in, token_type` |
| 3 | POST | `/api/v1/auth/refresh` | Public | `refresh_token*` | `access_token, refresh_token, expires_in` |
| 4 | POST | `/api/v1/auth/logout` | Public | `refresh_token*` | - |
| 5 | GET | `/api/v1/users/me` | Bearer | - | `id, name, email, avatar_url, created_at` |
| 6 | PATCH | `/api/v1/users/me` | Bearer | `name*`, `avatar_url?` | `id, name, email, avatar_url` |
| 7 | PATCH | `/api/v1/users/me/password` | Bearer | `current_password*`, `new_password*` | - |
| 8 | POST | `/api/v1/lists` | Bearer | `name*`, `description?` | `id, name, description, is_archived` |
| 9 | GET | `/api/v1/lists` | Bearer | - | Array of `{id, name, description, is_archived}` |
| 10 | GET | `/api/v1/lists/{listId}` | Bearer | - | Same as #8 |
| 11 | PATCH | `/api/v1/lists/{listId}` | Bearer | `name*`, `description?` | Same as #8 |
| 12 | DELETE | `/api/v1/lists/{listId}` | Bearer | - | - |
| 13 | PATCH | `/api/v1/lists/{listId}/archive` | Bearer | - | Same as #8 (is_archived=true) |
| 14 | PATCH | `/api/v1/lists/{listId}/restore` | Bearer | - | Same as #8 (is_archived=false) |
| 15 | GET | `/api/v1/lists/{listId}/board` | Bearer | - | `{backlog:[], todo:[], in_progress:[], review:[], done:[]}` |
| 16 | POST | `/api/v1/lists/{listId}/tasks` | Bearer | `title*`, `description?`, `priority?`, `status?`, `due_date?`, `estimated_minutes?` | Full task object |
| 17 | GET | `/api/v1/lists/{listId}/tasks` | Bearer | Query: `status, priority, search, due_date_from, due_date_to, is_overdue, sort_by, sort_order, page, limit` | `{data:[tasks], meta:{page,limit,total,total_pages}}` |
| 18 | GET | `/api/v1/tasks/{taskId}` | Bearer | - | Full task object |
| 19 | PATCH | `/api/v1/tasks/{taskId}` | Bearer | `title*`, `description?`, `priority?`, `due_date?`, `estimated_minutes?` | Full task object |
| 20 | DELETE | `/api/v1/tasks/{taskId}` | Bearer | - | Soft delete |
| 21 | PATCH | `/api/v1/tasks/{taskId}/status` | Bearer | `status*` | Full task object |
| 22 | PATCH | `/api/v1/tasks/{taskId}/priority` | Bearer | `priority*` | Full task object |
| 23 | POST | `/api/v1/tasks/{taskId}/reopen` | Bearer | - | Full task object |
| 24 | POST | `/api/v1/tasks/{taskId}/progress` | Bearer | `progress*(0-100)`, `note?`, `allow_rollback?` | `{id, task_id, progress, note}` |
| 25 | GET | `/api/v1/tasks/{taskId}/progress` | Bearer | - | Array of progress objects |
| 26 | GET | `/api/v1/tasks/{taskId}/progress/{progressId}` | Bearer | - | Single progress object |
| 27 | PATCH | `/api/v1/tasks/{taskId}/progress/{progressId}` | Bearer | `note*` | Single progress object |
| 28 | DELETE | `/api/v1/tasks/{taskId}/progress/{progressId}` | Bearer | - | - |
| 29 | POST | `/api/v1/tasks/{taskId}/submit-review` | Bearer | `submission_note` | `{id, task_id, status:"pending"}` |
| 30 | GET | `/api/v1/tasks/{taskId}/reviews` | Bearer | - | Array of review objects |
| 31 | GET | `/api/v1/tasks/{taskId}/reviews/{reviewId}` | Bearer | - | Single review object |
| 32 | POST | `/api/v1/tasks/{taskId}/reviews/{reviewId}/approve` | Bearer | `review_note?` | Review object (status=approved) |
| 33 | POST | `/api/v1/tasks/{taskId}/reviews/{reviewId}/request-changes` | Bearer | `review_note*` | Review object (status=changes_requested) |
| 34 | POST | `/api/v1/tasks/{taskId}/comments` | Bearer | `content*` | `{id, task_id, content}` |
| 35 | GET | `/api/v1/tasks/{taskId}/comments` | Bearer | - | Array of comment objects |
| 36 | PATCH | `/api/v1/tasks/{taskId}/comments/{commentId}` | Bearer | `content*` | Comment object |
| 37 | DELETE | `/api/v1/tasks/{taskId}/comments/{commentId}` | Bearer | - | - |
| 38 | GET | `/api/v1/tasks/{taskId}/activities` | Bearer | - | Array of activity objects |
| 39 | GET | `/api/v1/activities` | Bearer | Query: `page, limit` | `{data:[activities], meta:{...}}` |
| 40 | GET | `/api/v1/dashboard/summary` | Bearer | Query: `list_id?` | `{total_tasks, backlog, todo, in_progress, review, done, overdue, average_progress}` |
| 41 | GET | `/api/v1/dashboard/progress` | Bearer | Query: `period` (day/week/month/year) | Array of `{period, avg_progress, tasks_count}` |
| 42 | GET | `/api/v1/dashboard/upcoming-deadlines` | Bearer | Query: `limit?` | Array of `{id, title, due_date, status, priority, progress, list_name}` |
| 43 | GET | `/api/v1/dashboard/overdue-tasks` | Bearer | - | Array of `{id, title, due_date, status, priority, progress, list_name}` |
| 44 | GET | `/api/v1/dashboard/priority-distribution` | Bearer | Query: `list_id?` | Array of `{priority, count}` |
| 0 | GET | `/health` | Public | - | `{status:"ok"}` |

`*` = required, `?` = optional

### 2. Valid Status Transitions

```
backlog     → todo
todo        → backlog, in_progress
in_progress → todo, review
review      → in_progress, done
done        → in_progress   (via reopen)
```

### 3. Business Rules Summary

| Rule | Condition |
|------|-----------|
| **Progress only in_progress** | Progress hanya bisa ditambah saat task status = `in_progress` |
| **No backward progress** | Progress baru tidak boleh < progress lama (kecuali `allow_rollback: true` + note) |
| **Review needs 100%** | Submitting review requires progress = 100% |
| **Done via review** | Status `done` hanya bisa dicapai lewat review approval |
| **Reopen done** | Task `done` bisa dibuka kembali ke `in_progress` |
| **Auto started_at** | Server otomatis set `started_at` saat pertama ke `in_progress` |
| **Auto completed_at** | Review approved → `completed_at` otomatis terisi |
| **Soft delete** | Task & comment pakai soft delete (deleted_at) |
| **Ownership** | Semua query di-scope ke user_id pemilik |
| **Activity auto-log** | Semua perubahan status, progress, review, comment dicatat otomatis |

### 4. Standard API Response Format

**Success:**
```json
{"success": true, "message": "...", "data": {}}
```

**Success + Pagination:**
```json
{"success": true, "message": "...", "data": [], "meta": {"page":1, "limit":20, "total":100, "total_pages":5}}
```

**Error:**
```json
{"success": false, "message": "...", "errors": {"field": "..."}}
```

### 5. Gaps & Recommendations

| Gap | Severity | Recommendation |
|-----|----------|----------------|
| **No task reorder/drag-drop API** | Medium | Tambah `PATCH /tasks/{taskId}/reorder` jika butuh drag-drop |
| **No bulk task move** | Low | Opsional — untuk "select all → move to done" |
| **No user avatar upload** | Low | Endpoint PATCH users/me hanya terima URL, bukan file upload |
| **Due date in response is string** | Low | Frontend perlu parse `"2026-07-30"` manual |
| **Refresh token rotation** | Info | Saat refresh, token lama dicabut, dapat token baru — frontend harus handle ini |
| **Rate limiting** | Info | Hanya di auth endpoint (5 req/sec) |
| **Activity user_name missing** | Medium | Response activity tidak include `user_name` — frontend harus fetch terpisah atau backend tambah join query |
| **Task list board tidak paginate** | Low | Board endpoint return semua task tanpa pagination |

---

## B. Flutter Frontend Prompt

Copy teks di bawah ini dan berikan ke AI Agent untuk membangun frontend Flutter GoTask.

---

### PROMPT START

You are a Senior Flutter Developer. Build a complete mobile frontend for **GoTask** — a task management app with workflow system like a mini-Jira.

## Technology Requirements

- Flutter latest stable
- State management: **GetX** (dependency injection, routing, reactive state)
- Responsive: **flutter_screenutil**
- Architecture: **Clean Architecture** (feature-based)
- HTTP: **dio** for API calls
- Platforms: Android + iOS

## Design Language: Apple Human Interface

- Clean, minimal, elegant, premium feel
- Material 3 base with Apple-like visual touches
- Soft colors, rounded corners (12-16dp), subtle shadows
- Smooth animations and transitions (200-300ms duration)
- Comfortable spacing (16dp standard, 24dp section gaps)
- Typography: clear hierarchy (headline, body, caption)
- Light mode as default, dark mode optional
- Responsive for phone and tablet

## API Integration

**Base URL:** `http://YOUR_SERVER_URL/api/v1` (configurable via `.env`)
**Auth:** Bearer token in `Authorization` header
**Content-Type:** `application/json` or `multipart/form-data` (both supported)
**Response format:** `{"success": bool, "message": string, "data": any, "errors": map?, "meta": {page, limit, total, total_pages}?}`

### All Endpoints (Reference)

```
POST   /auth/register       {name*, email*, password*}
POST   /auth/login           {email*, password*} → {access_token, refresh_token, expires_in, token_type}
POST   /auth/refresh         {refresh_token*} → new tokens
POST   /auth/logout          {refresh_token*}

GET    /users/me             → User profile
PATCH  /users/me             {name*, avatar_url?}
PATCH  /users/me/password    {current_password*, new_password*}

POST   /lists                {name*, description?}
GET    /lists                → Array
GET    /lists/{id}           → Single
PATCH  /lists/{id}           {name*, description?}
DELETE /lists/{id}
PATCH  /lists/{id}/archive
PATCH  /lists/{id}/restore
GET    /lists/{id}/board     → {backlog:[], todo:[], in_progress:[], review:[], done:[]}

POST   /lists/{id}/tasks     {title*, description?, priority?, status?, due_date?, estimated_minutes?}
GET    /lists/{id}/tasks     ?status&priority&search&sort_by&sort_order&page&limit → paginated
GET    /tasks/{id}           → Single
PATCH  /tasks/{id}           {title*, description?, priority?, due_date?, estimated_minutes?}
DELETE /tasks/{id}           Soft delete
PATCH  /tasks/{id}/status    {status*}
PATCH  /tasks/{id}/priority  {priority*}
POST   /tasks/{id}/reopen    Reopen from done

POST   /tasks/{id}/progress  {progress*(0-100), note?, allow_rollback?}
GET    /tasks/{id}/progress  → Array
GET    /tasks/{id}/progress/{pid}
PATCH  /tasks/{id}/progress/{pid}  {note*}
DELETE /tasks/{id}/progress/{pid}

POST   /tasks/{id}/submit-review  {submission_note?}
GET    /tasks/{id}/reviews         → Array
GET    /tasks/{id}/reviews/{rid}
POST   /tasks/{id}/reviews/{rid}/approve         {review_note?}
POST   /tasks/{id}/reviews/{rid}/request-changes {review_note*}

POST   /tasks/{id}/comments   {content*}
GET    /tasks/{id}/comments   → Array
PATCH  /tasks/{id}/comments/{cid}  {content*}
DELETE /tasks/{id}/comments/{cid}

GET    /tasks/{id}/activities  → Array
GET    /activities             ?page&limit → paginated

GET    /dashboard/summary                ?list_id → summary counts
GET    /dashboard/progress               ?period(day|week|month|year)
GET    /dashboard/upcoming-deadlines     ?limit
GET    /dashboard/overdue-tasks
GET    /dashboard/priority-distribution  ?list_id
```

### Valid Status Values & Transitions

**Statuses:** `backlog`, `todo`, `in_progress`, `review`, `done`
**Priorities:** `low`, `medium`, `high`, `urgent`

```
backlog → todo
todo → backlog, in_progress
in_progress → todo, review  (review only if progress=100%)
review → in_progress, done  (done only via approve)
done → in_progress          (via reopen)
```

### Business Rules (MUST be implemented in Frontend Logic)

1. "Mark as In Progress" button only appears if status != `in_progress` AND transition is allowed
2. "Submit for Review" button disabled if progress < 100%
3. "Approve" button only for `pending` reviews
4. "Request Changes" requires review note (validate non-empty)
5. Progress slider/text field: validate 0-100, warn if less than current progress
6. Reopen button only visible for `done` tasks
7. "Delete" task shows confirmation dialog
8. Cannot edit/delete other users' comments (hide edit/delete if not owner — track user ID from login)

## Project Structure (Clean Architecture)

```
lib/
├── main.dart                          # App entry + GetMaterialApp
├── app/
│   ├── routes.dart                    # GetX named routes
│   ├── bindings.dart                  # Global bindings
│   └── theme.dart                     # Apple-inspired theme
├── core/
│   ├── constants.dart                 # API URL, app constants
│   ├── network/
│   │   ├── api_client.dart            # Dio instance + interceptors
│   │   └── api_interceptor.dart       # Auth token injection, error handling
│   ├── storage/
│   │   └── secure_storage.dart        # Token persistence
│   └── widgets/                       # Reusable widgets
│       ├── loading_widget.dart
│       ├── empty_state_widget.dart
│       ├── error_state_widget.dart
│       ├── status_badge.dart
│       ├── priority_badge.dart
│       └── task_card.dart
├── features/
│   ├── auth/
│   │   ├── controllers/auth_controller.dart
│   │   ├── screens/login_screen.dart
│   │   ├── screens/register_screen.dart
│   │   ├── screens/profile_screen.dart
│   │   └── widgets/
│   ├── dashboard/
│   │   ├── controllers/dashboard_controller.dart
│   │   └── screens/dashboard_screen.dart
│   ├── tasklist/
│   │   ├── controllers/tasklist_controller.dart
│   │   ├── screens/tasklist_screen.dart
│   │   ├── screens/tasklist_detail_screen.dart
│   │   └── widgets/
│   ├── task/
│   │   ├── controllers/task_controller.dart
│   │   ├── screens/task_board_screen.dart
│   │   ├── screens/task_detail_screen.dart
│   │   ├── screens/task_create_screen.dart
│   │   ├── screens/task_edit_screen.dart
│   │   └── widgets/
│   ├── progress/
│   │   ├── controllers/progress_controller.dart
│   │   └── widgets/progress_card.dart
│   ├── review/
│   │   ├── controllers/review_controller.dart
│   │   └── widgets/review_card.dart
│   ├── comment/
│   │   ├── controllers/comment_controller.dart
│   │   └── widgets/comment_card.dart
│   └── activity/
│       ├── controllers/activity_controller.dart
│       └── screens/activity_screen.dart
├── data/
│   ├── models/                        # JSON serializable models
│   │   ├── user_model.dart
│   │   ├── task_list_model.dart
│   │   ├── task_model.dart
│   │   ├── progress_model.dart
│   │   ├── review_model.dart
│   │   ├── comment_model.dart
│   │   ├── activity_model.dart
│   │   └── dashboard_model.dart
│   ├── repositories/                  # Repository implementations
│   │   ├── auth_repository.dart
│   │   ├── user_repository.dart
│   │   ├── tasklist_repository.dart
│   │   ├── task_repository.dart
│   │   ├── progress_repository.dart
│   │   ├── review_repository.dart
│   │   ├── comment_repository.dart
│   │   ├── activity_repository.dart
│   │   └── dashboard_repository.dart
│   └── datasources/                   # Individual API service classes
│       ├── auth_datasource.dart
│       ├── user_datasource.dart
│       └── ... (one per feature)
└── gen/                               # If using code generation
```

## Screens Required

### 1. Splash Screen
- Logo + app name
- Auto-check token → dashboard or login
- 2 second duration

### 2. Login Screen
- Email field with keyboard type: email
- Password field with show/hide toggle
- "Login" button (primary color, full width)
- "Don't have account? Register" link
- Loading state on button
- Error snackbar for wrong credentials

### 3. Register Screen
- Name, Email, Password fields
- Password requirements hint (8+ chars, uppercase, lowercase, number)
- "Register" button
- "Already have account? Login" link

### 4. Dashboard Screen (Home)
- **Header:** User greeting + avatar
- **Summary Cards Row:** Total, In Progress, Done (horizontal scroll)
- **Progress Chart:** Weekly progress bar chart
- **Priority Distribution:** Horizontal bar chart or pie chart
- **Upcoming Deadlines:** Horizontal list of task cards
- **Overdue Tasks:** Red-highlighted list (collapsible)
- Pull-to-refresh on entire screen

### 5. Task List Screen
- List of task lists with name, description, task count
- Swipe to archive/delete
- FAB to create new list
- Bottom sheet modal for create/edit list

### 6. Task Board Screen (Kanban-style)
- 5 columns: Backlog, Todo, In Progress, Review, Done
- Horizontal scrollable columns
- Each column: header with count + list of task cards
- Task card: title, priority badge, due date, progress bar, assignee
- Tap card → Task Detail
- Long press → quick action menu (change status, delete)
- Drag & drop between columns (optional, v2)
- FAB to create new task
- Filter chips at top (status, priority)

### 7. Task Detail Screen
- Title (editable inline)
- Description (expandable)
- Status badge (with dropdown to change)
- Priority badge (with dropdown to change)
- Progress bar + percentage
- Due date with calendar picker
- Estimated minutes
- **Progress Section:** List of progress updates + "Add Progress" button
- **Review Section:** Review status + Submit/Approve/Request Changes buttons
- **Comments Section:** Comment list + input field
- **Activity Section:** Timeline of activities
- Delete button (with confirmation)

### 8. Task Create/Edit Screen
- Title field
- Description field (multiline)
- Priority dropdown (low/medium/high/urgent)
- Status dropdown (backlog/todo/in_progress)
- Due date picker
- Estimated minutes (number input)
- "Save" button
- Form validation

### 9. Profile Screen
- Avatar (with edit)
- Name (editable)
- Email (read-only)
- Change Password section
- Logout button (red)

### 10. Activity Screen
- Full timeline of all activities
- Grouped by date
- Pagination (infinite scroll)

## GetX Controllers

Each controller should:
- Extend `GetxController`
- Use `.obs` reactive variables for UI state
- Have clear state management: `RxBool isLoading`, `RxString errorMessage`
- Call repository methods
- Handle loading, error, empty states

### Controller List

1. `AuthController` — login, register, logout, token storage, auto-login
2. `DashboardController` — summary, charts data, deadlines, overdue
3. `TaskListController` — CRUD task lists, archive/restore
4. `TaskBoardController` — board view, grouped tasks, filter
5. `TaskController` — CRUD task, status change, priority change, reopen
6. `ProgressController` — add/update/delete progress, validation
7. `ReviewController` — submit, approve, request changes
8. `CommentController` — CRUD comments, ownership check
9. `ActivityController` — list activities, pagination
10. `UserController` — profile, update profile, change password

## Routes (GetX named routes)

```dart
Routes.initial  → SplashScreen
Routes.login    → LoginScreen
Routes.register → RegisterScreen
Routes.dashboard → DashboardScreen
Routes.taskLists → TaskListScreen
Routes.taskBoard → TaskBoardScreen (param: listId)
Routes.taskDetail → TaskDetailScreen (param: taskId)
Routes.taskCreate → TaskCreateScreen (param: listId)
Routes.taskEdit   → TaskEditScreen (param: taskId)
Routes.profile    → ProfileScreen
Routes.activities → ActivityScreen
```

## Data Models (JSON Serializable)

Each model must have:
- `fromJson(Map<String, dynamic> json)` factory
- `toJson()` method
- All fields matching backend response exactly

## API Client (Dio)

- Base URL from environment config
- Interceptor: auto-attach Bearer token
- Interceptor: on 401 → try refresh token → retry original request → if refresh fails → logout
- Interceptor: log requests in debug mode
- Timeout: 30 seconds

## States (Every Screen MUST Implement)

| State | Widget | Trigger |
|-------|--------|---------|
| **Loading** | `LoadingWidget` (centered spinner) | Initial data fetch |
| **Empty** | `EmptyStateWidget` (icon + message) | List returns 0 items |
| **Error** | `ErrorStateWidget` (message + retry button) | API error/network error |
| **Success** | Normal content | Data loaded |

## Form Validation Rules

| Field | Rules |
|-------|-------|
| Email | Required, valid email format |
| Password | Required, min 8 chars, uppercase + lowercase + digit |
| Name | Required, min 2 chars |
| Task Title | Required |
| Progress | Required, 0-100 range |
| Comment Content | Required |
| Review Note (changes) | Required |

## Key Dependencies (pubspec.yaml)

```yaml
dependencies:
  get: latest
  dio: latest
  flutter_screenutil: latest
  json_annotation: latest
  flutter_secure_storage: latest
  intl: latest
  fl_chart: latest        # For dashboard charts
  shimmer: latest         # Loading skeleton
  cached_network_image: latest

dev_dependencies:
  json_serializable: latest
  build_runner: latest
```

## Implementation Priority

1. **Foundation:** Project structure, theme, routes, API client, token storage
2. **Auth:** Login, Register, auto-login, token refresh
3. **Dashboard:** Summary, charts
4. **Task List + Board:** Kanban board, task CRUD
5. **Task Detail:** Full detail with all sections
6. **Progress + Review:** Workflow actions
7. **Comments + Activity:** Social features
8. **Profile:** User settings
9. **Polish:** Animations, empty/error states, responsive

### PROMPT END
