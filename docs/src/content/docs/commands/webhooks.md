---
title: webhooks
description: Manage webhook integrations.
---

Webhooks are project integrations that receive event POSTs (purchases, renewals, cancellations, refunds, ...). Each webhook has a target URL, a list of events it subscribes to, and a `disabled` flag.

## Subcommands

| Command | Description |
| --- | --- |
| `webhooks list` | List webhook integrations |
| `webhooks view <id>` | Show one webhook |
| `webhooks create` | Create (shortcut flags, or `--file`) |
| `webhooks update <id>` | Update (`--url`, `--events`, `--disabled`, or `--file`) |
| `webhooks delete <id>` | Delete |

## `webhooks create` flags

| Flag | Description |
| --- | --- |
| `--url <url>` | Target URL |
| `--events <a>,<b>` | Events to subscribe to (comma-separated) |
| `--description <s>` | Optional description |
| `--file <path>` | JSON body (overrides shortcuts) |

## Examples

```sh
revcat webhooks list
revcat webhooks create --url https://hooks.example.com/rc \
  --events INITIAL_PURCHASE,RENEWAL,CANCELLATION
revcat webhooks update wh_xxx --disabled true
revcat webhooks delete wh_xxx -y
```
