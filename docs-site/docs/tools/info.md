---
sidebar_position: 1
---

# info

Progressive discovery system for BookLife capabilities.

## Usage

```json
{}
{"category": "hardcover"}
{"tool": "libby_place_hold"}
{"workflow": "find_and_read"}
```

## Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `category` | string | Show tools in category: `hardcover`, `libby`, `tbr`, `booklife`, `history`, `enrichment`, `sync`, `profile`, `recommendation` |
| `tool` | string | Show detailed help for a specific tool |
| `workflow` | string | Show step-by-step workflow guide: `find_and_read`, `sync_history`, `tbr_management`, `recommendations` |

## Behavior

- **No arguments**: Shows overview with all categories, workflows, and quick start
- **Category**: Lists tools in the category with brief descriptions
- **Tool**: Shows full parameter reference and examples
- **Workflow**: Provides step-by-step guide with tool calls

## Examples

### Overview
```json
{}
```
Returns category list, available workflows, and quick start instructions.

### Category browsing
```json
{"category": "libby"}
```
Returns: `libby_search`, `libby_get_loans`, `libby_get_holds`, `libby_place_hold`, `libby_sync_tag_metadata`, `libby_tag_metadata_list`

### Tool detail
```json
{"tool": "sync"}
```
Returns full parameter reference, typical flow, and examples.

### Workflow guide
```json
{"workflow": "tbr_management"}
```
Returns step-by-step TBR management workflow with tool calls.
