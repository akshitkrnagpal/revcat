---
title: packages
description: Manage RevenueCat packages (purchasables inside an offering)
---

Packages are the purchasable units inside an offering. Identifiers
follow RC's $rc_monthly / $rc_annual / custom convention.

## Subcommands

| Command | Description |
| --- | --- |
| `packages attach <pkg_id> <product_id> [<product_id> ...]` | Attach product(s) to a package |
| `packages create <offering>` | Create a package under an offering |
| `packages delete <id>` | Delete a package |
| `packages detach <id> <product_id> [<product_id> ...]` | Detach product(s) from a package |
| `packages list` | List packages across one offering or all offerings |
| `packages products <id>` | List products attached to a package |
| `packages update <id>` | Update a package |
| `packages view <id>` | Show one package by internal id |

Aliases: `pkg`.

Full flag reference: see [the CLI reference](/reference/cli/).
<!-- AUTOGEN_END -->

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
