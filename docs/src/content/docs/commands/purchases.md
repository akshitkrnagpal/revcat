---
title: purchases
description: Inspect non-renewing purchases
---

A purchase is a one-shot non-renewing transaction (lifetime grants,
consumables, in-app one-offs). For subscriptions see `revcat subscriptions`.

## Subcommands

| Command | Description |
| --- | --- |
| `purchases entitlements <id>` | List entitlements granted by a purchase |
| `purchases refund <id>` | Refund a non-renewing purchase (Web Billing) |
| `purchases search <store_id>` | Find purchases by store id |
| `purchases view <id>` | Show one purchase |

Aliases: `purchase`.

Full flag reference: see [the CLI reference](/reference/cli/).
<!-- AUTOGEN_END -->

## Examples

```sh
revcat purchases search ABC123XYZ
revcat purchases view purch_xxx
revcat purchases entitlements purch_xxx
revcat purchases refund purch_xxx -y
```
