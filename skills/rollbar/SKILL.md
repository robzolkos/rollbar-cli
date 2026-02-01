---
name: rollbar
description: Query Rollbar error tracking to investigate errors, list recent issues, and get context for bug fixes. Use when the user asks about production errors, wants to investigate an error, asks "what errors happened overnight", or needs context to fix a bug from Rollbar.
user-invocable: false
allowed-tools: Bash(rollbar *)
---

# Rollbar Error Investigation

Use the `rollbar` CLI to query Rollbar for error information. This tool is optimized for investigating production errors and providing context for bug fixes.

## Prerequisites

Ensure `ROLLBAR_ACCESS_TOKEN` is set in the environment, or a `.rollbar.yaml` config file exists in the project.

## Common Workflows

### Check what errors happened recently

```bash
# Errors from the last 8 hours
rollbar items --since "8 hours ago" --level error,critical

# Errors from overnight
rollbar items --since "12 hours ago" --level error,critical --env production

# All active issues
rollbar items --status active --limit 10
```

### Investigate a specific error

```bash
# Get details for error #123
rollbar item 123

# Get full context for bug fixing (includes stack trace, request data, user info)
rollbar context 123

# Include multiple recent occurrences
rollbar context 123 --occurrences 5
```

### Search for specific errors

```bash
# Find TypeError issues
rollbar items --query "TypeError" --level error

# Find errors in a specific environment
rollbar items --env production --level error --since "24h"

# Find most frequent errors
rollbar items --sort occurrences --limit 10
```

### View error occurrences

```bash
# List recent occurrences for an item
rollbar occurrences --item 123 --limit 5

# Get details for a specific occurrence
rollbar occurrence 453568801204
```

## Output Formats

- `--output table` (default): Human-readable tables
- `--output json`: Full JSON for parsing
- `--output compact` or `--ai`: Token-efficient format for AI context
- `--output markdown`: Structured markdown

## Key Flags

| Flag | Description |
|------|-------------|
| `--level` | Filter by level: debug, info, warning, error, critical |
| `--status` | Filter by status: active, resolved, muted, any |
| `--env` | Filter by environment (e.g., production, staging) |
| `--since` | Time filter: "8 hours ago", "24h", "7 days" |
| `--query` | Text search in item titles |
| `--sort` | Sort by: recent, occurrences, first-seen, level |
| `--limit` | Limit number of results |

## Typical Investigation Flow

1. **List recent errors**: `rollbar items --since "8 hours ago" --level error,critical`
2. **Identify the issue**: Look at error titles and occurrence counts
3. **Get full context**: `rollbar context <item-number>` for the specific error
4. **Analyze stack trace**: The context output includes file paths and line numbers
5. **Check patterns**: Look at multiple occurrences to understand the trigger

## Example: Overnight Error Triage

When asked "what errors happened overnight?" or similar:

```bash
# Get overnight errors (last 12 hours of night)
rollbar items --since "12 hours ago" --level error,critical --env production --ai

# For the most critical one, get full context
rollbar context <item-number>
```

The `--ai` flag provides compact output that's token-efficient while preserving all essential information.
