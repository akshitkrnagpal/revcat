---
title: virtual-currencies
description: Manage virtual currencies (coins / credits / tokens).
---

Project-level virtual currencies (in-game coins, credits, tokens). v2 keys virtual currencies by their **uppercase code** (e.g., `COIN`, `GEM`) - that's what you pass to view / update / delete / archive.

Per-customer balances and transactions are NOT exposed by v2 REST. Manage those in the dashboard or via the SDK.

## Subcommands

| Command | Description |
| --- | --- |
| `virtual-currencies list` | List virtual currencies |
| `virtual-currencies view <code>` | Show one virtual currency |
| `virtual-currencies create` | Create (`--name --code` or `--file`) |
| `virtual-currencies update <code>` | Update (`--name --description` or `--file`) |
| `virtual-currencies delete <code>` | Delete |
| `virtual-currencies archive <code>` | Archive |
| `virtual-currencies unarchive <code>` | Unarchive |

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
