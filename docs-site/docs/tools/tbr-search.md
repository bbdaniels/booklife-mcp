---
sidebar_position: 12
---

# tbr_search

Search your TBR list by title or author.

## Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `query` | string | Yes | Search term (matches title and author) |
| `source` | string | No | Filter by source: `physical`, `hardcover`, `libby` |
| `page` | int | No | Page number (default: 1) |
| `page_size` | int | No | Items per page (default: 20, max: 100) |

## Examples

```json
{"query": "Sanderson"}
{"query": "Mistborn", "source": "libby"}
```
