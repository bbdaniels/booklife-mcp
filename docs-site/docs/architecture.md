---
sidebar_position: 9
---

# Architecture

## Project Structure

```
booklife-mcp/
├── cmd/booklife/main.go         # Entry point, CLI commands
├── internal/
│   ├── config/                  # KDL configuration parsing
│   ├── models/                  # Shared types (Book, LibbyLoan, etc.)
│   ├── providers/               # External API clients
│   │   ├── interfaces.go        # Provider interfaces
│   │   ├── hardcover/           # Hardcover GraphQL client
│   │   ├── libby/               # Libby/OverDrive reverse-engineered API
│   │   ├── openlibrary/         # Open Library REST client
│   │   └── mocks/               # Test mocks
│   ├── server/                  # MCP server implementation
│   │   ├── server.go            # Server struct, initialization
│   │   ├── tools.go             # Tool registrations + shared helpers
│   │   ├── hardcover_handlers.go
│   │   ├── libby_handlers.go
│   │   ├── libby_tag_sync_handlers.go
│   │   ├── unified_handlers.go
│   │   ├── tbr_handlers.go
│   │   ├── history_handlers.go
│   │   ├── sync_handlers.go
│   │   ├── recommendation_handlers.go
│   │   ├── info_handlers.go
│   │   ├── resources.go         # MCP resources
│   │   └── prompts.go           # MCP prompts
│   ├── history/                 # Local SQLite history store
│   ├── tbr/                     # Local SQLite TBR store
│   ├── sync/                    # Sync engine (Libby → Hardcover)
│   ├── enrichment/              # Metadata enrichment service
│   ├── graph/                   # Book relationship graph
│   └── analytics/               # Reading profile computation
└── booklife.kdl                 # Example configuration
```

## Design Patterns

### Provider Interfaces

All external services are abstracted behind interfaces for testability:

```go
type HardcoverProvider interface {
    SearchBooks(ctx, query, offset, limit) ([]Book, int, error)
    GetUserBooks(ctx, status, offset, limit) ([]Book, int, error)
    UpdateBookStatus(ctx, bookID, status, progress, rating) error
    AddBook(ctx, isbn, title, author, status) (string, error)
    GetBook(ctx, bookID) (*Book, error)
}

type LibbyProvider interface {
    Search(ctx, query, formats, available, offset, limit) ([]Book, int, error)
    GetLoans(ctx) ([]LibbyLoan, error)
    GetHolds(ctx) ([]LibbyHold, error)
    PlaceHold(ctx, mediaID, format, autoBorrow) (string, error)
    GetTags(ctx) (map[string][]string, error)
}
```

### Conditional Provider Initialization

Providers are initialized only when enabled in config. Tools check for `nil` before execution:

```go
if s.hardcover == nil {
    return nil, nil, NewHardcoverNotConfiguredError()
}
```

### Progressive Disclosure

Complex tools use progressive disclosure to minimize token usage:

1. **Summary** — Quick stats (minimal tokens)
2. **List** — Items with IDs (moderate tokens)
3. **Detail** — Full information (maximum tokens)

### Cross-Tool ID Format

All tools output standardized IDs for chaining:

```
IDs: { book_id: 123, isbn: 9780593135204, media_id: 12345 }
```

- `book_id` → use with `hardcover_update_reading_status`
- `isbn` → use with `hardcover_add_to_library`
- `media_id` → use with `libby_place_hold`

### Automation Metadata

Responses include `_meta` for AI agent consumption:

```json
{
  "_meta": {
    "automation": {
      "next_actions": ["libby_place_hold", "hardcover_add_to_library"],
      "result_count": 5,
      "truncated": false
    }
  }
}
```

### Dual Store Strategy

- **Live APIs** — Hardcover and Libby for current state
- **Local SQLite** — History and TBR for offline access, enrichment cache, sync state

### Enrichment Chain

Book metadata enrichment follows a fallback chain:

1. **Hardcover** — Primary source (genres, tags, series, ratings)
2. **Open Library** — Subjects, descriptions, work IDs
3. **Google Books** — Categories, additional descriptions

### Sync Engine

The sync engine handles Libby → Hardcover reconciliation:

1. Find unsynced returned books in local history
2. Match to Hardcover by ISBN (primary) or title+author (fallback)
3. Cross-format lookup (audiobook ISBN → ebook ISBN via Libby search)
4. Cache book identity mappings for future syncs
5. Mark matched books as "read" in Hardcover
6. Track sync state per entry for retry/debugging

## MCP SDK Usage

BookLife uses the official [Go MCP SDK](https://github.com/modelcontextprotocol/go-sdk):

```go
mcp.AddTool(s.mcpServer, &mcp.Tool{
    Name: "tool_name",
    Description: "...",
}, s.handler)
```

Handlers return `(*mcp.CallToolResult, any, error)` where the second value is structured data for programmatic consumers.
