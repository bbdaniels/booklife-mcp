---
sidebar_position: 14
---

# tbr_remove

Remove a book from your TBR list.

## Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | int | No* | TBR entry ID (from `tbr_list` or `tbr_search`) |
| `title` | string | No* | Book title (used with `author`) |
| `author` | string | No* | Author name |

*Either `id` or `title`+`author` is required.

## Examples

```json
{"id": 42}
{"title": "The Name of the Wind", "author": "Patrick Rothfuss"}
```
