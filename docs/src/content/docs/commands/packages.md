---
title: packages
description: Manage RevenueCat packages (purchasables inside an offering).
---

Packages are the purchasable units inside an offering. Identifiers follow RC's `$rc_monthly` / `$rc_annual` / custom convention.

## Subcommands

| Command | Description |
| --- | --- |
| `packages list` | List packages across one offering or all offerings |
| `packages view <id>` | Show one package by internal id |
| `packages create` | Create a package from a JSON body (`--file`) |
| `packages update <id>` | Update a package (`--file`) |
| `packages delete <id>` | Delete a package |
| `packages products <id>` | List products attached to a package |
| `packages attach <id> <product_id> ...` | Attach product(s) to a package |
| `packages detach <id> <product_id> ...` | Detach product(s) from a package |

## Examples

```sh
revcat packages list                       # flat across all offerings
revcat packages list --offering default    # one offering only
revcat packages view pkg_xxx
revcat packages attach pkg_xxx prod_app_monthly
```

Aliases: `pkg`.

## Body shape (create)

```json
{
  "identifier": "$rc_monthly",
  "offering_id": "ofr_xxx",
  "position": 1
}
```
