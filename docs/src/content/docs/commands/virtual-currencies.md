---
title: virtual-currencies
description: Manage virtual currencies (coins / credits / tokens).
---

Project-level virtual currencies (in-game coins, credits, tokens). For per-customer balances and transactions see [`revcat subscribers vc-balance`](/commands/subscribers/) / `vc-tx` / `vc-set-balance`.

## Subcommands

| Command | Description |
| --- | --- |
| `virtual-currencies list` | List virtual currencies |
| `virtual-currencies view <id>` | Show one virtual currency |
| `virtual-currencies create` | Create (`--file`, required) |
| `virtual-currencies update <id>` | Update (`--file`, required) |
| `virtual-currencies delete <id>` | Delete |
| `virtual-currencies archive <id>` | Archive |
| `virtual-currencies unarchive <id>` | Unarchive |

## Examples

```sh
revcat virtual-currencies list
revcat virtual-currencies create --file ./vc/coins.json
revcat virtual-currencies archive coins -y
```

Aliases: `vc`.

## Body shape (create)

```json
{
  "lookup_key": "coins",
  "display_name": "Coins",
  "code": "COIN"
}
```
