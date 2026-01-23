---
sidebar_position: 18
---

# booklife_best_way_to_read

Determine the best way to access a book, prioritizing free/library options.

## Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `book_id` | string | No* | Hardcover book ID |
| `isbn` | string | No* | ISBN-13 or ISBN-10 |
| `preferences` | string[] | No | Preferred formats: `ebook`, `audiobook`, `physical` |

*Either `book_id` or `isbn` is required.

## Examples

```json
{"isbn": "9780756404741"}
{"book_id": "123", "preferences": ["audiobook", "ebook"]}
```

## Response

Returns prioritized access options:
1. Library (free) — with availability status
2. Bookstore — local options
3. Online — purchase links

Includes actionable IDs for each option.
