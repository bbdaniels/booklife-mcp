---
sidebar_position: 2
---

# Sync Reading History

How to sync your Libby reading history to Hardcover so returned books are tracked.

## Initial Setup (One Time)

### 1. Import your Libby timeline

Get your timeline export:
1. Open Libby app → **Settings** → **Reading History**
2. Tap **Export Timeline** → Copy the JSON link

```json
// history_import_timeline
{"url": "https://share.libbyapp.com/data/{uuid}/libbytimeline-all-loans.json"}
```

### 2. Run initial sync

```json
// sync
{"action": "sync_all"}
```

This will:
- Import current loans
- Match returned books to Hardcover
- Start metadata enrichment
- Cache Libby tag data

## Ongoing Sync

### Check status

```json
// sync
{"action": "status"}
```

Shows pending count and last sync time.

### Preview changes

```json
// sync
{"action": "preview"}
```

Lists books that will be marked as "read" in Hardcover.

### Execute sync

```json
// sync
{"action": "run"}
```

Or test without changes:
```json
{"action": "run", "dry_run": true}
```

### Comprehensive sync

```json
// sync
{"action": "sync_all"}
```

Runs all four steps: import → sync → enrich → cache tags.

## Troubleshooting

### Check unmatched books

```json
// sync
{"action": "unmatched"}
```

Shows books that couldn't be found in Hardcover. Filter by type:
```json
{"action": "unmatched", "unmatched_type": "isbn"}
{"action": "unmatched", "unmatched_type": "no_isbn"}
```

### Check specific entry

```json
// sync
{"action": "details", "entry_id": "abc123"}
```

Shows the full activity history and sync state for one book.

## How Matching Works

1. **ISBN match** — Most reliable. Searches Hardcover by ISBN.
2. **Title + Author** — Fallback. Fuzzy matching on title and author name.
3. **Cross-format lookup** — If an audiobook ISBN isn't found, searches Libby for the ebook edition's ISBN and tries that.
4. **Identity cache** — Successful matches are cached for instant lookup on future syncs.
