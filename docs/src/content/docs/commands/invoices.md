---
title: invoices
description: Inspect invoices
---

Inspect invoices

## Subcommands

| Command | Description |
| --- | --- |
| `invoices view <id>` | Show one invoice (raw JSON) |

Full flag reference: see [the CLI reference](/reference/cli/).
<!-- AUTOGEN_END -->

## Example

```sh
revcat invoices view inv_xxx | jq .
```
