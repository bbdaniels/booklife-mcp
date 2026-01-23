---
sidebar_position: 13
---

# tbr_add

Add a book to your TBR list manually (for physical books or custom entries).

## Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `title` | string | Yes | Book title |
| `author` | string | Yes | Author name |
| `subtitle` | string | No | Book subtitle |
| `isbn10` | string | No | ISBN-10 |
| `isbn13` | string | No | ISBN-13 |
| `publisher` | string | No | Publisher name |
| `published_date` | string | No | Publication date |
| `page_count` | int | No | Number of pages |
| `description` | string | No | Book description |
| `cover_url` | string | No | Cover image URL |
| `genres` | string[] | No | Genre list |
| `series_name` | string | No | Series name |
| `series_position` | float | No | Position in series |
| `notes` | string | No | Personal notes |
| `priority` | int | No | Priority 0-10 (higher = more important) |
| `source` | string | No | Source label (default: `physical`) |

## Examples

```json
{"title": "The Name of the Wind", "author": "Patrick Rothfuss"}
{"title": "Mistborn", "author": "Sanderson", "priority": 5, "notes": "Recommended by friend"}
{"title": "Project Hail Mary", "author": "Andy Weir", "isbn13": "9780593135204", "source": "physical"}
```
