---
sidebar_position: 20
---

# history_sync_current_loans

Sync current Libby loans to local history store.

## Parameters

None required.

```json
{}
```

## Purpose

Captures your current Libby checkouts into the local history database. This is useful for:
- Building a complete reading history
- Tracking books before they're returned
- Feeding the sync engine with current activity

## Notes

- Also runs automatically as part of `sync` with `action="sync_all"`
- Idempotent — safe to run multiple times
