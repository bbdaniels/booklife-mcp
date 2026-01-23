---
sidebar_position: 4
---

# hardcover_add_to_library

Add a book to your Hardcover library.

## Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `isbn` | string | No* | ISBN-13 or ISBN-10 |
| `title` | string | No* | Book title (used with `author` for matching) |
| `author` | string | No* | Author name |
| `status` | string | No | Initial status (default: `want-to-read`) |
| `place_hold` | bool | No | Also place a library hold via Libby |

*Either `isbn` or `title`+`author` is required.

## Examples

```json
{"isbn": "9780593135204", "status": "want-to-read"}
{"title": "Project Hail Mary", "author": "Andy Weir"}
{"isbn": "9780756404741", "status": "want-to-read", "place_hold": true}
```

## Notes

- When `place_hold` is `true` and Libby is configured, BookLife will search the library catalog and place a hold automatically
- ISBN lookup is more reliable than title+author matching
- Returns the Hardcover `book_id` for subsequent operations
