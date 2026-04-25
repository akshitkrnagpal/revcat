---
title: invoices
description: Inspect invoices.
---

## Subcommands

| Command | Description |
| --- | --- |
| `invoices view <id>` | Show one invoice (raw JSON) |

To list invoices for a specific customer use [`revcat subscribers invoices <user_id>`](/commands/subscribers/).

## Example

```sh
revcat invoices view inv_xxx | jq .
```
