---
sidebar_position: 4
---

# Getting Recommendations

How to get personalized book recommendations based on your reading history.

## Setup (One Time)

### 1. Enrich your history

```json
// enrichment_enrich_history
{}
```

This background job fetches descriptions, themes, topics, and mood data for all books in your history. Required for content-based recommendations.

### 2. Monitor progress

```json
// enrichment_status
{}
```

Check the job status until complete.

## Get Recommendations

### View your reading profile

```json
// profile_get
{}
```

Shows format preferences, top genres, favorite authors, reading cadence, and streaks.

### Find similar books

```json
// book_find_similar
{"title": "Project Hail Mary", "author": "Andy Weir", "limit": 10}
```

Returns books from your history that share themes, topics, and mood. Use this to discover patterns in what you enjoy.

### Find at library

Once you have a recommendation, check availability:

```json
// booklife_find_book_everywhere
{"query": "recommended book title"}
```

## Tips

- Enrichment processes ~1-2 seconds per book
- Use `force: true` to re-enrich after new books are added
- The more books in your history, the better recommendations get
- Enrichment also runs as part of `sync` with `action="sync_all"`
- Profile updates automatically as you read more
