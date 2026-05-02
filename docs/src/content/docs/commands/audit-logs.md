---
title: audit-logs
description: Inspect the project's audit log
---

Inspect the project's audit log

## Subcommands

| Command | Description |
| --- | --- |
| `audit-logs list` | List audit log entries |

Aliases: `audit`.

Full flag reference: see [the CLI reference](/reference/cli/).
<!-- AUTOGEN_END -->

## Examples

```sh
revcat audit-logs list
revcat audit-logs list --output json | jq '.[] | select(.action == "delete")'
```

Aliases: `audit`.
