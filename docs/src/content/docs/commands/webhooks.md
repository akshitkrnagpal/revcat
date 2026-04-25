---
title: webhooks
description: Manage webhook integrations.
---

Webhooks are project integrations that receive event POSTs (purchases, renewals, cancellations, refunds, ...). Each webhook has a name, target URL, and a list of `event_types` it subscribes to.

Event values are LOWERCASE in the API config (`initial_purchase`, `renewal`, ...) - even though the webhook payload itself uses screaming case (`INITIAL_PURCHASE`). revcat lowercases values passed via `--events` for you, so either form works on the CLI.

## Subcommands

| Command | Description |
| --- | --- |
| `webhooks list` | List webhook integrations |
| `webhooks view <id>` | Show one webhook |
| `webhooks create` | Create (`--name --url --events` or `--file`) |
| `webhooks update <id>` | Update (`--name --url --events` or `--file`) |
| `webhooks delete <id>` | Delete |

## `webhooks create` flags

| Flag | Description |
| --- | --- |
| `--name <s>` | Webhook name (required) |
| `--url <url>` | Target URL (required, must pass RC's reachability check) |
| `--events <a>,<b>` | Event types (comma-separated, lowercased automatically) |
| `--file <path>` | JSON body (overrides shortcuts) |

## Examples

```sh
revcat webhooks list
revcat webhooks create \
  --name "Production hook" \
  --url https://hooks.example.com/rc \
  --events initial_purchase,renewal,cancellation
revcat webhooks update wh_xxx --events INITIAL_PURCHASE,RENEWAL,CANCELLATION,EXPIRATION
revcat webhooks delete wh_xxx -y
```

## Notes on URL validation

RC validates the URL is reachable when you create / update. `localhost`, `example.com`, and unresolvable domains will fail. For testing use a real https endpoint (e.g., webhook.site).

## Common event types

`initial_purchase`, `renewal`, `cancellation`, `expiration`, `billing_issue`, `non_renewing_purchase`, `uncancellation`, `transfer`, `subscription_paused`, `product_change`, `subscription_extended`, `temporary_entitlement_grant`. The full v2 list is enforced server-side; an invalid event will be returned with the valid set.
