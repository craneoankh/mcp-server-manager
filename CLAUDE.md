# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

MCP Server Manager is a Go web application that centralizes management of Model Context Protocol (MCP) servers across multiple AI clients. It solves the problem of manually editing various JSON files for different MCP clients by providing a single YAML configuration file and web interface.

**Key Features:**
- Single binary with embedded assets (no external dependencies)
- Cross-platform support (Linux, macOS, Windows)
- Real-time web interface with HTMX
- Automatic client config synchronization
- Systemd integration for auto-start

## Development Guidelines

### Important Development Notes
- **DO NOT run `make run`** - The development server is already running in the background
- **DO NOT include Claude Code attribution in git commits** - Use semantic commit messages instead
- Use semantic git commit format: `feat:`, `fix:`, `docs:`, `refactor:`, `chore:`, etc.
- Example: `feat: add syntax highlighting to config viewer` instead of generic attribution

### Core Commands

- `make build` - Build the binary to `bin/mcp-server-manager` (single binary with embedded assets)
- `make install-deps` - Download and organize Go dependencies
- `make setup` - Complete production setup (build, install systemd user service, enable, start)
- `make logs-service` - View systemd user service logs in real-time
- `make status-service` - Check systemd user service status
- `make sync-assets` - Sync web assets from web/ to internal/assets/web/ for embedding
- `make test-release` - Build local .deb package, install, and restart service for testing
- `make release VERSION=x.x.x` - Create git tag and trigger GitHub Actions release

### Release Commands

- `make test-release` - Build and test .deb package locally before release
- `make release VERSION=v1.1.0` - Create official release with git tag and trigger GitHub Actions
- `make sync-assets` - Sync web assets to embedded location (auto-included in build/release)
- This triggers GitHub Actions to build cross-platform binaries via GoReleaser
- Produces releases for Linux, macOS, Windows (amd64 + arm64)

## Detailed Architecture

### File Structure & Key Components

```
├── cmd/server/main.go           # Application entry point with embedded assets setup
├── internal/
│   ├── assets/                  # Embedded web assets (templates, CSS, JS)
│   │   └── web/                 # Mirror of web/ directory for embedding
│   ├── config/loader.go         # YAML config loading/saving with validation
│   ├── handlers/                # HTTP request handlers
│   │   ├── api.go              # REST API endpoints for programmatic access
│   │   ├── web.go              # HTMX endpoints returning HTML fragments
│   │   └── config_viewer.go    # Configuration display handlers
│   ├── models/config.go         # Data structures for app and client configs
│   └── services/                # Core business logic
│       ├── mcp_manager.go      # Central orchestration service
│       ├── client_config.go    # Client JSON config manipulation
│       └── validator.go        # Configuration validation logic
├── web/                         # Source templates and static files
│   ├── templates/*.html         # Go templates with HTMX integration
│   └── static/                  # CSS, JS, Prism.js syntax highlighting
├── configs/config.yaml          # Example configuration file
└── systemd/mcp-server-manager.service  # User systemd service definition
```

### Core Services Architecture

**MCPManagerService** (`internal/services/mcp_manager.go`):
- Central orchestrator coordinating between central config and client configs
- Handles global/per-client server toggling and sync operations
- Manages the two-layer configuration model (central YAML → client JSONs)

**ClientConfigService** (`internal/services/client_config.go`):
- Reads/writes individual AI client configuration files
- Handles heterogeneous client config formats (Claude vs Gemini structures)
- Creates automatic timestamped backups before modifications
- Preserves existing client settings (theme, auth, etc.) during updates

**ValidatorService** (`internal/services/validator.go`):
- Validates YAML configuration and client JSON structures
- Checks command availability in PATH for MCP servers
- Handles different server types (command-based vs HTTP-based)

### Configuration Flow Architecture

The application operates on a **two-layer configuration model**:

1. **Central Configuration** (`configs/config.yaml`):
   - Defines all available MCP servers with commands, args, environment variables
   - Specifies which clients exist and their config file paths
   - Controls global enable/disable states and per-client overrides

2. **Client Configurations** (e.g., `~/.claude/settings.json`, `~/.gemini/settings.json`):
   - Individual AI client settings that get automatically updated
   - Preserves client-specific settings (themes, auth) while updating MCP sections
   - Supports different formats: Claude (command-based) vs Gemini (command + HTTP)

### Web Interface & HTMX Patterns

**Handler Types**:
- **APIHandler**: REST endpoints (`/api/*`) for programmatic access, returns JSON
- **WebHandler**: HTMX endpoints (`/htmx/*`) returning HTML fragments for reactive UI
- **ConfigViewerHandler**: Read-only configuration display with syntax highlighting

**HTMX Integration Patterns**:
- Toggle checkboxes post to `/htmx/...` endpoints, update specific DOM elements
- Config viewers use `hx-trigger="revealed, configChanged from:body"` for lazy loading + auto-refresh
- Custom `configChanged` events trigger config viewer refreshes after any changes
- Hyperscript handles event triggering after HTMX requests complete

### Web Interface Features

**Add New Servers via Web UI**:
- Form-based server addition with JSON validation
- Real-time validation for MCP client configuration format
- Example configurations for STDIO, HTTP, SSE, and Context7 servers
- Client-side validation before submission

**Interactive Configuration Examples**:
- Pre-built examples for common server types
- One-click loading of example configurations
- Supports all transport types (command, httpUrl, url)

**Configuration Display**:
- YAML config displayed in original format (not JSON conversion)
- Real-time syntax highlighting with Prism.js
- Auto-refresh after configuration changes

### Embedded Assets System

**Production Build**:
- Uses `//go:embed` to bundle all web assets into the binary
- `internal/assets/assets.go` defines embedded filesystems for templates and static files
- No external file dependencies in production - completely self-contained binary
- Template parsing uses `template.ParseFS()` with custom `dict` function for HTMX data passing

**Development vs Production**:
- Development: Templates loaded from `web/templates/`, static files from `web/static/`
- Production: Everything embedded, served from memory via `http.FS()`
- Asset sync: Changes in `web/` must be copied to `internal/assets/web/` for embedding

### Client Configuration Complexity

**Heterogeneous Config Formats**:
- Claude: `mcpServers` with command/args structure
- Gemini: `mcpServers` supporting both command-based AND HTTP servers with headers
- Solution: `ClientConfig.MCPServers` uses `map[string]interface{}` for flexibility

**Synchronization Process**:
1. Read current client config (create empty if missing)
2. Update `mcpServers` section based on central config + client-specific overrides
3. Preserve non-MCP settings (authentication, themes, editor preferences)
4. Create automatic `.backup.TIMESTAMP` files before writing
5. Validate structure before writing to prevent corruption

### Key Technical Constraints & Solutions

**Development Environment**:
- Server runs on port 6543 (not 6543)
- Development server runs in background - never run `make run` during development
- Changes to templates/static files require `make sync-assets` or are auto-synced during build
- Use `make test-release` to test complete .deb package installation locally

**Git Workflow**:
- Use semantic commit messages: `feat:`, `fix:`, `docs:`, `refactor:`, `chore:`
- Never include Claude Code attribution in commits
- Example: `feat: add syntax highlighting to config viewer`

**Production Deployment**:
- Single binary with embedded assets (no external dependencies)
- Cross-platform releases via GoReleaser (Linux/macOS/Windows, amd64/arm64)
- Systemd user service integration for Linux auto-start
- Release triggered by Git tag push, builds via GitHub Actions

### Syntax Highlighting Implementation

**Prism.js Integration**:
- Embedded Prism.js core, YAML, and JSON language modules
- Config viewers automatically detect content type (JSON for client configs)
- HTMX auto-refresh triggers `Prism.highlightAllUnder()` for new content
- CSS and JS files embedded in binary via assets system

## TODO

- Remove MCP servers via Web UI
- Test on macOS and Windows
- Allow users to edit files in the Web UI
- When the app config is changed manually on disk, the web UI does not reflect the changes until the server is restarted.
- Investigate and document the requirement for enabling user lingering (`loginctl enable-linger <username>`) to ensure the systemd user service starts automatically on system boot.
- Add 'profiles' functionality (low priority) to save and restore different application configurations. This would allow users to switch between predefined sets of server states. Implementation could involve a new `profiles` key in the main config file or a small local database.
- Refactor the 'Add New Server' functionality. The current implementation uses a full page reload, which prevents the success notification from being displayed. Change it to use HTMX to asynchronously add the server and update the server list in place, without a page reload. This will also allow the success notification to be displayed correctly.
- Allow adding new clients and their config paths via the web UI. As part of this, consider refactoring the `clients` section in the main config from a map to an array of objects (e.g., `{name: 'client-name', path: '/path/to/config.json'}`) for easier manipulation.
- Investigate adding a feature to open configuration files in the user's default text editor directly from the web UI.
- Update the documentation to explain the benefits of using a dotfile manager (e.g., `chezmoi`) to manage the central `config.yaml`. Mention that this allows for version control, synchronization, and managing secrets like API keys through the dotfile manager's encryption features. [https://www.chezmoi.io/](https://www.chezmoi.io/)
- Consider removing the 'global' enable/disable flag for servers. It adds complexity, may be buggy, and per-client toggles may be sufficient, which would simplify the UI and configuration logic.