---
sidebar_position: 21
---

# history_get

Get reading history from local store with pagination and search.

## Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `query` | string | No | Search term (matches title and author) |
| `page` | int | No | Page number (default: 1) |
| `page_size` | int | No | Items per page (default: 20, max: 100) |

## Examples

```json
{"page": 1, "page_size": 20}
{"query": "Sanderson", "page": 1}
```

## Response

Returns timeline entries with:
- Title and author
- Checkout/return dates
- Format (ebook/audiobook)
- Library name
- ISBN (when available)
