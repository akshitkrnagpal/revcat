---
title: packages
description: Manage RevenueCat packages (purchasables inside an offering).
---

Packages are the purchasable units inside an offering. Lookup_keys follow RC's `$rc_monthly` / `$rc_annual` / custom convention.

Packages live under offerings, so create takes `<offering>` as a positional argument. Once created, a package has its own `pkg...` system id that other commands use.

## Subcommands

| Command | Description |
| --- | --- |
| `packages list [--offering <id>]` | List packages across one offering or all |
| `packages view <pkg_id>` | Show one package by system id |
| `packages create <offering>` | Create a package under an offering (`--id --display-name --position` or `--file`) |
| `packages update <pkg_id>` | Update a package (`--file`) |
| `packages delete <pkg_id>` | Delete a package |
| `packages products <pkg_id>` | List products attached to a package |
| `packages attach <pkg_id> <product_id> [...]` | Attach product(s) to a package |
| `packages detach <pkg_id> <product_id> [...]` | Detach product(s) from a package |

## Examples

```sh
revcat packages list                                      # flat across all offerings
revcat packages list --offering default                   # one offering only
revcat packages create default --id '$rc_monthly' --display-name "Monthly" --position 0
revcat packages view pkg_xxx
revcat packages attach pkg_xxx prod_app_monthly --eligibility all
```

Aliases: `pkg`.

## Eligibility on attach

Each product attachment carries an `eligibility_criteria` value that gates when RC offers the product. `--eligibility` defaults to `all`. Other valid values:

- `all` - every customer (default)
- `google_sdk_lt_6` - clients on Google Play SDK <6
- `google_sdk_ge_6` - clients on Google Play SDK >=6

## Body shape (create via `--file`)

```json
{
  "lookup_key": "$rc_monthly",
  "display_name": "Monthly",
  "position": 0
}
```

`offering_id` is determined by the `<offering>` positional argument; including it in the body is ignored.
