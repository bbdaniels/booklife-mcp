---
sidebar_position: 5
---

# libby_search

Search your library catalog via Libby for ebooks and audiobooks.

## Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `query` | string | Yes | Search query (title, author, ISBN). Max 500 chars. |
| `format` | string[] | No | Filter: `ebook`, `audiobook`, `magazine` |
| `available` | bool | No | Only show immediately available items |
| `language` | string | No | Language filter (e.g., `eng`, `spa`) |
| `sort_by` | string | No | Sort: `relevance`, `title`, `author`, `date` |
| `page` | int | No | Page number (default: 1) |
| `page_size` | int | No | Items per page (default: 20, max: 100) |

## Examples

```json
{"query": "Project Hail Mary"}
{"query": "Brandon Sanderson", "available": true}
{"query": "mystery", "format": ["audiobook"], "page_size": 5}
```

## Response

Returns books with:
- `media_id` → for `libby_place_hold`
- Availability status (available now or waitlist)
- Format options (ebook, audiobook)
- Estimated wait times
- ISBN for cross-referencing
