---
title: subscriptions
description: Inspect and manage subscriptions
---

A subscription is one customer's ongoing purchase relationship with one
product. Find it via `revcat subscriptions search <store_id>` or by
listing it under a customer (`revcat subscribers info`).

## Subcommands

| Command | Description |
| --- | --- |
| `subscriptions cancel <id>` | Cancel a subscription (Web Billing) |
| `subscriptions entitlements <id>` | List entitlements granted by a subscription |
| `subscriptions management-url <id>` | Print the store-specific manage/cancel URL |
| `subscriptions refund <id>` | Refund the entire subscription (Web Billing) |
| `subscriptions search <store_id>` | Find subscriptions by store id (App Store / Play / Stripe / ...) |
| `subscriptions transactions <id>` | List billing transactions for a subscription |
| `subscriptions view <id>` | Show one subscription |

Aliases: `sub`.

Full flag reference: see [the CLI reference](/reference/cli/).
<!-- AUTOGEN_END -->

## Examples

```sh
revcat subscriptions search ABC123XYZ
revcat subscriptions view sub_xxx
revcat subscriptions transactions sub_xxx
revcat subscriptions cancel sub_xxx -y
revcat subscriptions refund sub_xxx -y
```

Aliases: `sub`.
