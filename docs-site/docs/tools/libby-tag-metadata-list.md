---
sidebar_position: 10
---

# libby_tag_metadata_list

List cached Libby tag metadata with full book information.

## Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `tag` | string | No | Filter by specific tag |
| `page` | int | No | Page number (default: 1) |
| `page_size` | int | No | Items per page (default: 20, max: 100) |

## Examples

```json
{}
{"tag": "favorites"}
{"tag": "sci-fi", "page_size": 50}
```

## Prerequisites

Run `libby_sync_tag_metadata` first to populate the cache.

## Response

Returns cached book details for tagged items including title, author, ISBN, format, and availability status.
