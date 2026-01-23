---
sidebar_position: 25
---

# enrichment_status

Query progress and status of an enrichment job.

## Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `job_id` | string | No | Specific job ID (defaults to most recent) |

## Examples

```json
{}
{"job_id": "abc-123"}
```

## Response

Returns:
- Job status: `pending`, `running`, `completed`, `failed`, `cancelled`
- Processed / successful / failed counts
- Current book being processed
- Elapsed time and estimated remaining
- Average time per book
- Recent errors (last 10)
