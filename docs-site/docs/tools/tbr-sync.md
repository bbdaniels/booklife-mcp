---
sidebar_position: 15
---

# tbr_sync

Sync TBR from external sources (Hardcover, Libby).

## Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `action` | string | Yes | Sync action (see below) |

### Actions

| Action | Description |
|--------|-------------|
| `sync_hardcover` | Import Hardcover "want-to-read" books to TBR |
| `sync_libby_holds` | Import current Libby holds to TBR |
| `sync_libby_tags` | Import Libby tagged books with full metadata |
| `sync_all` | Run all syncs |

## Examples

```json
{"action": "sync_hardcover"}
{"action": "sync_libby_tags"}
{"action": "sync_all"}
```

## Notes

- `sync_libby_tags` fetches full book information and stores locally for offline access
- Existing entries are updated, not duplicated
- Run periodically to keep TBR current with external sources
