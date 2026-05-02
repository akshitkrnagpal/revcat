---
title: Commands
description: Reference for every revcat subcommand.
---

revcat is organized by RevenueCat resource: customers, offerings, paywalls, etc. Each top-level group has read commands (`list`, `view`) and write commands (`create`, `update`, `delete`, plus action verbs like `attach`, `archive`, `refund`).

## Catalog

| Group | Reads | Writes |
| --- | --- | --- |
| [`projects`](/commands/projects/) | `list`, `view` | - |
| [`apps`](/commands/apps/) | `list`, `view`, `public-keys`, `storekit-config` | - |
| [`entitlements`](/commands/entitlements/) | `list`, `view`, `products` | `create`, `update`, `delete`, `archive`, `unarchive`, `attach`, `detach` |
| [`offerings`](/commands/offerings/) | `list`, `view`, `preview` | `create`, `update`, `delete`, `archive`, `unarchive`, `set-current` |
| [`packages`](/commands/packages/) | `list`, `view`, `products` | `create`, `update`, `delete`, `attach`, `detach` |
| [`products`](/commands/products/) | `list`, `view` | `create`, `update`, `delete`, `archive`, `unarchive`, `push-to-store` |
| [`paywalls`](/commands/paywalls/) | `list`, `view` | `create`, `delete` |

## Customers

| Group | Reads | Writes |
| --- | --- | --- |
| [`subscribers`](/commands/subscribers/) | `info`, `list`, `attributes`, `invoices` | `create`, `delete`, `grant`, `revoke`, `refund`, `transfer` |
| [`subscriptions`](/commands/subscriptions/) | `view`, `transactions`, `entitlements`, `management-url`, `search` | `cancel`, `refund` |
| [`purchases`](/commands/purchases/) | `view`, `entitlements`, `search` | `refund` |
| [`invoices`](/commands/invoices/) | `view` | - |

## Activity

| Group | Reads | Writes |
| --- | --- | --- |
| [`metrics`](/commands/metrics/) | `overview` | - |
| [`charts`](/commands/charts/) | `get`, `options` | - |
| [`audit-logs`](/commands/audit-logs/) | `list` | - |
| [`collaborators`](/commands/collaborators/) | `list` | - |

## Integrations

| Group | Reads | Writes |
| --- | --- | --- |
| [`webhooks`](/commands/webhooks/) | `list`, `view` | `create`, `update`, `delete` |
| [`virtual-currencies`](/commands/virtual-currencies/) | `list`, `view` | `create`, `update`, `delete`, `archive`, `unarchive` |

## Verb-orchestrators

| Command | Description |
| --- | --- |
| [`publish offering`](/commands/publish/) | Set an offering as current and / or push a paywall config in one shot |

## Auth + housekeeping

[`auth`](/commands/auth/) (login, status, doctor, use, list, logout), [`init`](/commands/init/), [`doctor`](/commands/doctor/), `completion`, `version`. See the individual pages for details.

## Global flags

Available on every command:

| Flag | Description |
| --- | --- |
| `--profile <name>` | Global auth profile to use (default: `REVCAT_PROFILE` env or `default`). Ignored when a `.revcat/config.json` is walked up from cwd. |
| `--project-id <id>` | RevenueCat project id (default: `REVCAT_PROJECT_ID`, walked-up `.revcat/config.json`, or `revcat.toml`) |
| `--bypass-keychain` | Use `~/.revcat/config.json` (file backend) instead of the OS keychain |
| `--output table\|json\|csv\|markdown` | Force an output format. Auto-detected when omitted (table on TTY, JSON when piped) |
| `--pretty` | Indent JSON output |
| `-v, --verbose` | Show detailed output |
| `-q, --quiet` | Suppress non-essential output |
| `--no-color` | Disable color |
| `--debug` | Show stack traces |
