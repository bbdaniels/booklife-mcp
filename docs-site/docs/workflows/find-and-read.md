---
sidebar_position: 1
---

# Find and Read a Book

Step-by-step workflow for discovering a book, checking availability, and starting to read.

## Steps

### 1. Search for the book

```json
// booklife_find_book_everywhere
{"query": "The Name of the Wind"}
```

This searches Hardcover and your library simultaneously. Look for:
- `book_id` — for Hardcover operations
- `media_id` — for library holds
- `isbn` — for cross-referencing

### 2. Check library availability

The search results include availability. Look for:
- "Available now" — borrow immediately
- "Wait list" — place a hold

### 3a. If available at library

```json
// libby_place_hold
{"media_id": "12345", "format": "ebook"}
```

Or with auto-borrow:
```json
{"media_id": "12345", "format": "audiobook", "auto_borrow": true}
```

### 3b. If not available, add to TBR

```json
// tbr_add
{"title": "The Name of the Wind", "author": "Patrick Rothfuss", "isbn13": "9780756404741"}
```

### 4. Track in Hardcover

```json
// hardcover_add_to_library
{"isbn": "9780756404741", "status": "want-to-read"}
```

Or combine with a library hold:
```json
{"isbn": "9780756404741", "status": "want-to-read", "place_hold": true}
```

### 5. Start reading

Once you have the book, update your status:
```json
// hardcover_update_reading_status
{"book_id": "123", "status": "reading"}
```

## Tips

- Always check library first — it's free
- Use `booklife_find_book_everywhere` to search all sources at once
- `place_hold` with `auto_borrow: true` means you don't have to check back
- Add to TBR as a reminder even if you're placing a hold
