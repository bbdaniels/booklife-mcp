---
sidebar_position: 17
---

# booklife_find_book_everywhere

Search all sources for a book and show comprehensive availability.

## Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `query` | string | Yes | Search query (title, author, or ISBN) |

## Examples

```json
{"query": "The Name of the Wind"}
{"query": "9780756404741"}
{"query": "Andy Weir"}
```

## Response

Returns unified results from Hardcover and Libby including:
- Book metadata (title, author, genres, ratings)
- All cross-tool IDs (`book_id`, `isbn`, `media_id`)
- Library availability with format options
- Access recommendation (library, bookstore, etc.)

## Use Cases

This is the primary discovery tool — start here when looking for a book. The response includes all IDs needed for follow-up actions:
- `book_id` → `hardcover_add_to_library`
- `media_id` → `libby_place_hold`
- `isbn` → `hardcover_add_to_library`
