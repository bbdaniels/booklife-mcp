---
sidebar_position: 2
---

# hardcover_get_my_library

Get your reading list from Hardcover with progressive detail levels.

## Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `status` | string | No | Filter: `reading`, `read`, `want-to-read`, `dnf`, `all` (default: `all`) |
| `detail` | string | No | Detail level: `summary`, `list` (default), `full` |
| `sort_by` | string | No | Sort: `date_added`, `title`, `author`, `rating`, `progress` |
| `page` | int | No | Page number (default: 1) |
| `page_size` | int | No | Items per page (default: 20, max: 100) |

## Detail Levels

- **`summary`** — Quick stats: counts per status, average progress/rating (~200 tokens)
- **`list`** — Book titles with IDs, status, and progress (default)
- **`full`** — Complete metadata including genres, series, ratings

## Examples

```json
{"detail": "summary"}
{"status": "reading"}
{"status": "want-to-read", "page_size": 10}
{"status": "read", "sort_by": "rating"}
```

## Response

Returns books with cross-tool IDs:
- `book_id` → for `hardcover_update_reading_status`
- `isbn` → for `libby_search`, `hardcover_add_to_library`
