---
sidebar_position: 7
---

# libby_get_holds

Get your current library holds and queue positions.

## Parameters

None required.

```json
{}
```

## Response

Returns active holds with:
- Title and author
- `media_id` for reference
- Format (ebook/audiobook)
- Queue position
- "Ready to borrow" status
- Estimated wait in days
- Auto-borrow setting
