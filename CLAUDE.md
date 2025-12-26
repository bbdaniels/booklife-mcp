# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

BookLife is an MCP (Model Context Protocol) server for managing a personal reading life across multiple platforms. It unifies Hardcover (reading tracker), Libby/OverDrive (library access), Open Library (metadata), and local bookstores into a cohesive assistant that prioritizes free/local options.

## Build and Run Commands

```bash
# Build the server
cd booklife-mcp && go build -o booklife ./cmd/booklife

# Run the server (requires booklife.kdl config)
./booklife --config booklife.kdl

# Run tests
go test ./...

# Run a single test
go test -run TestName ./internal/...
```

## Environment Variables

Required:
- `HARDCOVER_API_KEY` - Hardcover GraphQL API token
- `LIBBY_CLONE_CODE` - 8-digit clone code from Libby app (Settings > Copy To Another Device)

Optional:
- `YOUTUBE_API_KEY` - For BookTube integration
- `BOOKLIFE_CONFIG_PATH` - Config file path (default: `./booklife.kdl`)
- `BOOKLIFE_LOG_LEVEL` - debug/info/warn/error

## Architecture

```
booklife-mcp/
├── cmd/booklife/main.go     # Entry point, signal handling
├── internal/
│   ├── config/              # KDL configuration parsing with env var resolution
│   ├── server/              # MCP server setup
│   │   ├── server.go        # Server struct, provider initialization
│   │   ├── tools.go         # MCP tool handlers (search_books, get_my_library, etc.)
│   │   ├── resources.go     # MCP resource handlers (booklife://library/*, etc.)
│   │   └── prompts.go       # MCP prompt templates (what_should_i_read, etc.)
│   ├── providers/           # External API clients
│   │   ├── hardcover/       # GraphQL client for Hardcover
│   │   ├── libby/           # Reverse-engineered Libby/OverDrive API
│   │   └── openlibrary/     # Open Library REST API with rate limiting
│   └── models/              # Shared types (Book, LibbyLoan, LibraryAvailability, etc.)
```

### Key Patterns

**Provider Initialization**: Providers are conditionally initialized in `server.initProviders()` based on config `Enabled` flags. Tools check for nil providers before executing.

**MCP SDK Usage**: Uses `github.com/modelcontextprotocol/go-sdk` with typed tool handlers:
```go
mcp.AddTool(s.mcpServer, &mcp.Tool{Name: "tool_name", Description: "..."}, s.handler)
```

**Configuration**: KDL v1 format with environment variable resolution (`api-key env="VAR_NAME"` syntax). See SPEC.md for full config structure.

**Rate Limiting**: Open Library client uses `golang.org/x/time/rate` for 10 req/sec limiting.

## MCP Interface

### Tools
- **Library Management**: `search_books`, `get_my_library`, `update_reading_status`, `add_to_library`
- **Libby**: `search_library`, `check_availability`, `get_loans`, `get_holds`, `place_hold`
- **Unified**: `find_book_everywhere`, `best_way_to_read`, `add_to_tbr`

### Resources
- `booklife://library/current` - Currently reading books
- `booklife://library/tbr` - Want-to-read list
- `booklife://loans` - Active Libby loans
- `booklife://holds` - Library hold queue
- `booklife://stats` - Reading statistics

### Prompts
- `what_should_i_read` - Personalized recommendations
- `book_summary` - Book summaries with spoiler control
- `reading_wrap_up` - Period reading summaries
- `pick_from_tbr` - TBR decision helper

## Implementation Status

See SPEC.md for full implementation phases. Current state:
- Phase 1-2: Core infrastructure and Hardcover integration (implemented)
- Phase 3: Libby integration (implemented)
- Phase 4: Unified actions (partially implemented)
- Phase 5-6: Community/semantic features (not started)

## Claude Desktop Integration

```json
{
  "mcpServers": {
    "booklife": {
      "command": "/path/to/booklife",
      "args": ["--config", "/path/to/booklife.kdl"],
      "env": {
        "HARDCOVER_API_KEY": "...",
        "LIBBY_CLONE_CODE": "..."
      }
    }
  }
}
```
