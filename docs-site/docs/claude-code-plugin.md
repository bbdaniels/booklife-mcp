---
sidebar_position: 7
---

# Claude Code Plugin

BookLife has an enhanced Claude Code plugin available in the [andy-marketplace](https://github.com/andylbrummer/andy-marketplace/tree/main/plugins/booklife) that provides intelligent skills and convenient slash commands.

## Installation

```bash
# Add the marketplace (if not already added)
/plugin marketplace add https://github.com/andylbrummer/andy-marketplace

# Install the plugin
/plugin install booklife@andy-marketplace

# Verify it's enabled
/plugin list
```

## Slash Commands

Quick-action commands for common tasks:

| Command | Description |
|---------|-------------|
| `/booklife:reading-status` | Current reading snapshot — books, loans, holds, due dates |
| `/booklife:sync-now` | Comprehensive sync — history, enrichment, tags |
| `/booklife:whats-next` | Personalized recommendation based on TBR and profile |
| `/booklife:find-at-library [query]` | Search library catalog and place holds |
| `/booklife:tbr-review` | Review TBR stats, availability, and priorities |

## Skills

Skills trigger automatically based on conversation context:

### Reading Sync & Management
Triggers when discussing reading progress, syncing libraries, or updating book status.

### Library Search & Availability
Triggers when searching for books, checking library availability, or placing holds.

### TBR Management
Triggers when discussing reading lists, adding books to TBR, or organizing what to read.

### Book Discovery & Recommendations
Triggers when asking for recommendations, "books like X", or what to read next.

## Natural Language Examples

The plugin's skills activate on natural language — no commands needed:

- "What books are due soon at the library?"
- "Add The Name of the Wind to my TBR"
- "Show me books similar to Piranesi"
- "Sync my Libby history to Hardcover"
- "What should I read next? I'm in the mood for something light."

## Prerequisites

The BookLife MCP server must be running and configured in Claude Desktop/Code. See [Getting Started](./getting-started) for setup.

## Source

Full plugin source and documentation: [andy-marketplace/plugins/booklife](https://github.com/andylbrummer/andy-marketplace/tree/main/plugins/booklife)
