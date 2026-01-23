---
sidebar_position: 1
slug: /
---

# BookLife MCP

**Your reading life, unified.** BookLife is an [MCP (Model Context Protocol)](https://modelcontextprotocol.io) server that connects your library, reading tracker, and bookshelf into a seamless AI-powered reading assistant.

## What It Does

BookLife bridges three platforms into one conversational interface:

- **[Hardcover](https://hardcover.app)** — Reading tracker with library management, status updates, and community ratings
- **[Libby/OverDrive](https://libbyapp.com)** — Free library access to ebooks and audiobooks
- **[Open Library](https://openlibrary.org)** — Open metadata for enrichment and cover images

## Key Capabilities

| Capability | Description |
|-----------|-------------|
| **Library Search** | Search your library catalog, check availability, place holds |
| **Reading Tracker** | Update status, progress, ratings on Hardcover |
| **Unified TBR** | One list from all sources — Hardcover, Libby, physical books |
| **History Sync** | Sync returned Libby books as "read" in Hardcover |
| **Enrichment** | Add themes, topics, mood from Open Library/Google Books |
| **Recommendations** | Content-based similarity from your reading history |
| **Profile Analytics** | Format preferences, genres, cadence, streaks |

## How It Works

```
┌──────────────────────────────────────────┐
│           Claude / AI Assistant          │
└────────────────────┬─────────────────────┘
                     │ MCP (stdio)
┌────────────────────▼─────────────────────┐
│           BookLife MCP Server            │
│                                          │
│  ┌─────────┐ ┌─────────┐ ┌───────────┐  │
│  │Hardcover│ │  Libby  │ │OpenLibrary│  │
│  │ GraphQL │ │   API   │ │  REST API │  │
│  └─────────┘ └─────────┘ └───────────┘  │
│                                          │
│  ┌──────────────────────────────────┐    │
│  │        Local SQLite Store        │    │
│  │  History · TBR · Enrichment      │    │
│  └──────────────────────────────────┘    │
└──────────────────────────────────────────┘
```

## Quick Example

Ask Claude:

> "What am I currently reading, and are any library books due soon?"

BookLife will query both Hardcover and Libby to give you a unified reading snapshot with due dates and progress.

> "Find 'Project Hail Mary' at the library and place a hold if available."

BookLife searches your library catalog, checks availability, and places the hold — all in one conversation.

## Next Steps

- [Getting Started](/docs/getting-started) — Install and configure
- [Tool Reference](/docs/category/tool-reference) — All 27 tools
- [Workflows](/docs/category/workflows) — Step-by-step guides
- [Claude Code Plugin](/docs/claude-code-plugin) — Enhanced integration
