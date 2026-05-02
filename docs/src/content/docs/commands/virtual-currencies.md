---
title: virtual-currencies
description: Manage virtual currencies (coins / credits)
---

Project-level virtual currencies (in-game coins, credits, tokens).
v2 keys VCs by their uppercase code (e.g., COIN, GEM) - that's the
identifier you pass to view/update/delete/archive.

Per-customer balances and transactions are NOT exposed by v2 REST.

## Subcommands

| Command | Description |
| --- | --- |
| `virtual-currencies archive <code>` | Archive a virtual currency |
| `virtual-currencies create` | Create a virtual currency |
| `virtual-currencies delete <code>` | Delete a virtual currency |
| `virtual-currencies list` | List virtual currencies |
| `virtual-currencies unarchive <code>` | Unarchive a virtual currency |
| `virtual-currencies update <code>` | Update a virtual currency |
| `virtual-currencies view <code>` | Show one virtual currency by uppercase code (e.g. COIN) |

Aliases: `vc`.

Full flag reference: see [the CLI reference](/reference/cli/).
<!-- AUTOGEN_END -->

## Examples

```sh
revcat virtual-currencies create --name Coins --code COIN --description "in-game currency"
revcat virtual-currencies list
revcat virtual-currencies view COIN
revcat virtual-currencies update COIN --name "Gold Coins"
revcat virtual-currencies archive COIN -y
```

Aliases: `vc`.

## Body shape (create)

```json
{
  "name": "Coins",
  "code": "COIN",
  "description": "in-game currency"
}
```
