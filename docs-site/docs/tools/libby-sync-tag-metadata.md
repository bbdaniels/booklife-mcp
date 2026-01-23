---
sidebar_position: 9
---

# libby_sync_tag_metadata

Sync full book information for all Libby tagged books to local cache.

## Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `tag` | string | No | Sync only books with this specific tag |

## Examples

```json
{}
{"tag": "favorites"}
```

## Purpose

Fetches complete metadata (title, author, ISBN, cover, format, availability) for all books you've tagged in Libby and stores it locally for:

- Browsing tagged books offline
- Building reading lists and recommendations
- Cross-referencing with Hardcover
- Organizing your library collection

## Notes

- Data comes from current loans and holds
- Returned books may have limited metadata
- Use `libby_tag_metadata_list` to browse cached data after syncing
- Also runs as part of `sync` with `action="sync_all"`
