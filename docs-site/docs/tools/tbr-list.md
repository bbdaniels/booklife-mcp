---
sidebar_position: 11
---

# tbr_list

List your unified TBR (to-be-read) list from all sources.

## Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `source` | string | No | Filter: `physical`, `hardcover`, `libby` |
| `page` | int | No | Page number (default: 1) |
| `page_size` | int | No | Items per page (default: 20, max: 100) |

## Examples

```json
{}
{"source": "libby"}
{"source": "hardcover", "page_size": 50}
```

## Sources

- **physical** — Manually added books (bookstore purchases, gifts)
- **hardcover** — Books with "want-to-read" status on Hardcover
- **libby** — Library holds and tagged books
