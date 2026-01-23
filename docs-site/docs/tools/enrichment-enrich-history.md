---
sidebar_position: 24
---

# enrichment_enrich_history

Enrich reading history with metadata from Open Library and Google Books.

## Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `force` | bool | No | Re-enrich all books, even if already enriched (default: false) |

## Examples

```json
{}
{"force": true}
```

## Purpose

Fetches and stores:
- Book descriptions
- Themes and topics
- Mood classifications
- Series information
- Subject categories

This data is required for content-based recommendations (`book_find_similar`).

## Behavior

- Runs asynchronously as a background job
- Processes entire library (~1-2 seconds per book)
- Already-enriched books are skipped unless `force=true`
- Returns job ID for progress monitoring

## Monitoring

Use `enrichment_status` to check progress:

```json
{"job_id": "returned-job-id"}
```
