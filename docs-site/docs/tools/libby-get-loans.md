---
sidebar_position: 6
---

# libby_get_loans

Get your current Libby loans with due dates and progress.

## Parameters

None required.

```json
{}
```

## Response

Returns active loans with:
- Title and author
- `media_id` for reference
- Format (ebook/audiobook)
- Due date with days remaining
- Overdue and "due soon" warnings
- Reading progress percentage
