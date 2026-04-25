---
title: audit-logs
description: Inspect the project's audit log.
---

## Subcommands

| Command | Description |
| --- | --- |
| `audit-logs list` | List audit log entries (TTY shows `when / action / actor / resource`; JSON returns the full payload) |

## Examples

```sh
revcat audit-logs list
revcat audit-logs list --output json | jq '.[] | select(.action == "delete")'
```

Aliases: `audit`.
