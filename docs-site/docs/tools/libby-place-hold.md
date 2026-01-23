---
sidebar_position: 8
---

# libby_place_hold

Place a hold on a library ebook or audiobook.

## Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `media_id` | string | Yes | Media ID from `libby_search` or `libby_get_holds` |
| `format` | string | Yes | Format: `ebook` or `audiobook` |
| `auto_borrow` | bool | No | Automatically borrow when available (default: false) |

## Examples

```json
{"media_id": "12345", "format": "ebook"}
{"media_id": "12345", "format": "audiobook", "auto_borrow": true}
```

## Response

Returns:
- Hold confirmation with hold ID
- Queue position
- Estimated wait time

## Prerequisites

Get `media_id` from:
- `libby_search` results
- `booklife_find_book_everywhere` results
- `libby_get_holds` (for existing holds)
