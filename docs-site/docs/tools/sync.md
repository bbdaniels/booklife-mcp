---
sidebar_position: 23
---

# sync

Universal sync tool for reading history, enrichment, and Libby tag metadata.

## Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `action` | string | No | Action to perform (default: `status`) |
| `target` | string | No | Target system (default: `hardcover`) |
| `entry_id` | string | No | For `details` action |
| `limit` | int | No | For `run`/`sync_all` action: max entries |
| `dry_run` | bool | No | For `run`/`sync_all` action: preview only |
| `unmatched_type` | string | No | For `unmatched` action: `isbn`, `no_isbn`, `all` |

## Actions (Progressive Disclosure)

| Action | Description |
|--------|-------------|
| `status` | Show pending count, last sync, error summary |
| `preview` | List books that will be synced |
| `run` | Execute history sync only |
| `sync_all` | Comprehensive: history + enrichment + tag metadata |
| `details` | Show sync state for specific entry |
| `unmatched` | Show books that failed to match |

## Examples

```json
{}
{"action": "status"}
{"action": "preview"}
{"action": "run"}
{"action": "sync_all"}
{"action": "sync_all", "dry_run": true}
{"action": "details", "entry_id": "abc123"}
{"action": "unmatched", "unmatched_type": "isbn"}
```

## sync_all Flow

The comprehensive sync runs four steps:

1. **Import current loans** — Fetches active Libby checkouts to local history
2. **Sync history** — Marks returned books as "read" in Hardcover
3. **Enrich metadata** — Starts background enrichment job for unenriched books
4. **Cache tag metadata** — Syncs Libby tagged books with full book info

## Typical Workflow

```
1. sync action="status"      → Check what's pending
2. sync action="preview"     → See what will sync
3. sync action="run"         → Execute (or sync_all for everything)
4. sync action="unmatched"   → Review failures
```
