# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Running the Application
```bash
# Local development
go run cmd/app/main.go

# With custom config
go run cmd/app/main.go -config config.yaml

# Build binary
go build -o b1cron cmd/app/main.go

# Docker development
docker-compose up -d
docker-compose logs -f

# Stop services
docker-compose down
```

### Database Operations
- Database file: `data/b1cron.db` (SQLite)
- GORM automatically handles migrations on startup
- Default user: admin/admin (created on first run)
- **Script Files**: Stored in `data/shell/` and `data/python/` directories
- **Script Naming**: `task_<ID>_<timestamp>.<ext>` format

## Architecture Overview

### Core Architecture
B1Cron is a web-based cron job management system built with Go, following a clean architecture pattern:

```
cmd/app/main.go           → Application entry point & routing setup
internal/
├── models/               → Data models (User, Task, TaskExecution)
├── database/             → Database connection & initialization
├── scheduler/            → gocron v2 integration for job scheduling
├── service/              → Business logic layer
├── handler/              → HTTP handlers (Gin framework)
├── auth/                 → JWT authentication middleware
└── config/               → Configuration management
```

### Key Design Patterns

**Service Layer Pattern**: Business logic is separated into services (`TaskService`, `ScriptFileService`)
- `TaskService`: Manages task CRUD operations and scheduler integration
- `ScriptFileService`: Handles script file creation and execution command generation

**Scheduler Integration**: Uses gocron v2 with UUID job tracking
- Tasks store `GocronJobID` for scheduler management
- Automatic task recovery on startup via `loadExistingTasks()`
- Task execution creates `TaskExecution` records with status tracking

**Script Execution Model**: Supports multiple script types
- `command`: Direct shell commands (executed via `/bin/sh -c` on Unix, `cmd /C` on Windows)
- `shell`: Shell script files (`.sh`)  
- `python`: Python script files (`.py`)
- Scripts are written to `data/{shell|python}/` directories
- All commands executed through shell to support redirection, pipes, and other shell features

### Frontend Architecture

**Technology Stack**: HTMX + Modular JavaScript + Tailwind CSS
- Templates use Go's `html/template` with custom functions (`js`, `jsRaw`, `toFloat64`, `div`)
- Modular component-based JavaScript architecture (see JavaScript Architecture section)
- Uses `dashboard-simplified.js` (not `dashboard.js` - that's been removed)
- SimpleToast for notifications (lightweight toast system)

**Template Structure**:
- `base.html`: Base template with common layout
- `dashboard.html`: Main dashboard view
- `login.html`: Authentication page
- `_*.html`: Modal and component templates
- HTMX powers dynamic UI interactions (toggle tasks, delete, etc.)

### Authentication & Security

**JWT-based Authentication**: Uses `gin-jwt/v2`
- JWT tokens stored in HTTP-only cookies
- Middleware protects `/dashboard` and `/api/*` routes
- Password hashing with bcrypt

**Configuration**: YAML-based config with sensible defaults
- Default credentials: admin/admin (should be changed in production)
- JWT secret should be updated for production deployment

## Database Schema

### Core Models
- **User**: Basic user authentication (username, password_hash)
- **Task**: Scheduled jobs with script support and multiple execution types
  - Basic fields: name, command, script_type, is_enabled, gocron_job_id
  - Scheduling: schedule_spec, schedule_type ("cron", "once", "range", "dynamic")  
  - One-time execution: execute_at (*time.Time) for schedule_type="once"
- **TaskExecution**: Execution history (task_id, status, started_at, completed_at, duration, output, error_msg)

### Important GORM Patterns
- Soft deletion enabled on User and Task models
- Use `Unscoped()` when querying soft-deleted tasks (e.g., for execution history)
- Foreign key relationships: TaskExecution → Task

## Frontend Development Notes

### JavaScript Architecture

**Modular Component System**: Component-based JavaScript architecture with separated concerns
- **components.js**: Component registry and loading system with dependency management
- **task-form.js**: TaskForm component for creating/editing tasks with schedule type switching
- **tab-manager.js**: TabManager for UI tab switching with Tailwind CSS class management
- **schedule-type-manager.js**: Handles switching between "cron" and "once" schedule types
- **schedule-manager.js**: ScheduleManager for cron expression generation and parsing
- **code-editor.js**: CodeMirror wrapper for syntax highlighting and multi-line editing
- **dashboard-simplified.js**: Main dashboard functionality and HTMX integration
- **modern.js**: Core application functionality and global utilities
- **toast-libraries.js**: SimpleToast implementation (preferred over other toast systems)

**Component Loading Order**: Components must be loaded in dependency order via base.html:
1. Individual component files (tab-manager.js, schedule-*.js, etc.)
2. components.js (registry system)
3. Application-specific files (dashboard-simplified.js)

### CSS & Styling
- **Light Mode Only**: Dark mode has been completely removed from templates and CSS
- **Tailwind CSS**: CDN-based with custom color palette defined in templates
- **Custom CSS**: `static/css/tailwind-custom.css` for animations, CodeMirror styling, and component overrides

### Key Frontend Functions
- `editTaskFromElement()`: Edit task from dashboard table (fetches via API for complete data)
- `showExecutionDetailFromElement()`: View execution details
- Global `window.b1cron` object for modal and toast management

### Task Display Compatibility
Dashboard template intelligently displays schedule information based on task type:
- **Periodic tasks**: Shows schedule_spec with 🔄 icon in gray background
- **One-time tasks**: Shows execute_at datetime with ⏰ icon in blue background
- Edit buttons include all necessary data attributes: schedule_type, execute_at, script_type

## Development Guidelines

### Cron Expression Support
The system supports three schedule types:
```bash
# Periodic execution - Standard cron (5 fields: minute hour day month weekday)
30 9 * * *          # Daily at 9:30 AM
0 */2 * * *         # Every 2 hours

# Periodic execution - @every interval format (gocron extension)
@every 5m           # Every 5 minutes
@every 1h           # Every 1 hour

# One-time execution - via datetime picker in UI
# Uses execute_at field instead of schedule_spec
# Tasks execute once at specified time then auto-complete
```

### Error Handling Patterns
- Services return `(result, error)` tuples
- HTTP handlers use JSON error responses
- Task execution errors are captured in `TaskExecution.ErrorMsg`
- Scheduler errors are logged but don't stop the service

### Testing Notes
- No formal test suite currently exists
- Manual testing via web interface
- Database operations can be tested with temporary SQLite files

## Production Deployment

### Docker Deployment (Recommended)
```bash
docker-compose up -d
```
- Builds multi-stage Docker image (~60MB final size)  
- Includes volume persistence for `data/` directory
- Configured for production with `gin.ReleaseMode`
- **Runtime Environment**: Alpine Linux with Python 3, pip, and bash support
- **Script Execution**: Supports shell scripts (`/bin/bash`) and Python scripts (`python3`)

### Configuration for Production
- Change JWT secret in `config.yaml`
- Update default admin credentials
- Set `server.mode: "release"`
- Consider setting `jwt.secure: true` for HTTPS