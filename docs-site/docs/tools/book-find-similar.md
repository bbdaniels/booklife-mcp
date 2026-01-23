---
sidebar_position: 26
---

# book_find_similar

Find books similar to a given book based on themes, topics, and mood.

## Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `title` | string | Yes | Book title to find similar to |
| `author` | string | Yes | Author of the reference book |
| `limit` | int | No | Max results (default: 10) |

## Examples

```json
{"title": "Project Hail Mary", "author": "Andy Weir", "limit": 10}
{"title": "The Name of the Wind", "author": "Patrick Rothfuss"}
```

## Prerequisites

Requires enrichment data. Run `enrichment_enrich_history` first to populate themes and topics.

## How It Works

Uses content-based matching on:
- Themes and topics (from enrichment)
- Mood and complexity
- Genre overlap
- Writing style indicators

Results are ranked by similarity score from your reading history.
