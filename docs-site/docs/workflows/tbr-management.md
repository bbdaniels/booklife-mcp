---
sidebar_position: 3
---

# TBR Management

Managing your unified to-be-read list across all sources.

## Initial Setup

### Sync from all sources

```json
// tbr_sync
{"action": "sync_all"}
```

This pulls in:
- Hardcover "want-to-read" list
- Libby holds
- Libby tagged books (with full metadata)

## Daily Usage

### View your TBR

```json
// tbr_stats
{}
```

Quick overview: total books, breakdown by source, availability stats.

### Browse by source

```json
// tbr_list
{"source": "libby", "page_size": 20}
```

Options: `physical`, `hardcover`, `libby`, or omit for all.

### Search your TBR

```json
// tbr_search
{"query": "science fiction"}
{"query": "Sanderson", "source": "libby"}
```

### Add books manually

For physical books or recommendations:

```json
// tbr_add
{"title": "Project Hail Mary", "author": "Andy Weir", "priority": 8}
{"title": "Piranesi", "author": "Susanna Clarke", "notes": "Book club pick", "source": "physical"}
```

### Remove finished books

```json
// tbr_remove
{"id": 42}
{"title": "The Name of the Wind", "author": "Patrick Rothfuss"}
```

## Keeping in Sync

Run periodically to stay current:

```json
// tbr_sync
{"action": "sync_all"}
```

Or sync specific sources:
```json
{"action": "sync_hardcover"}
{"action": "sync_libby_holds"}
{"action": "sync_libby_tags"}
```

## Tips

- Use `priority` (0-10) to rank what to read next
- `sync_libby_tags` fetches full metadata for offline browsing
- Check `tbr_stats` to see how many TBR books are available at the library now
- After finishing a book, remove from TBR and update Hardcover status
