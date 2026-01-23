---
sidebar_position: 3
---

# hardcover_update_reading_status

Update a book's status, progress, or rating in Hardcover.

## Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `book_id` | string | Yes | Hardcover book ID (from `hardcover_get_my_library` or search) |
| `status` | string | Yes | New status: `reading`, `read`, `want-to-read`, `dnf` |
| `progress` | int | No | Reading progress 0-100 (percentage) |
| `rating` | float | No | Rating 0-5 (half-star increments) |

## Examples

```json
{"book_id": "123", "status": "reading", "progress": 50}
{"book_id": "123", "status": "read", "rating": 4.5}
{"book_id": "456", "status": "dnf", "progress": 20}
```

## Notes

- `book_id` is obtained from `hardcover_get_my_library` or `booklife_find_book_everywhere`
- Progress is only meaningful for `reading` status
- Rating persists across status changes
