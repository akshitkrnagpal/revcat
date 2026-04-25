---
title: events
description: Inspect RevenueCat events (lifecycle activity firehose).
---

Events is the firehose of subscription lifecycle activity in your project: purchases, renewals, cancellations, trials, refunds, etc.

Use `revcat events list` for a one-shot page, or `revcat events tail` to follow new events as they arrive (kubectl-logs-style).

## Subcommands

| Command | Description |
| --- | --- |
| `events list` | Print one page of recent events |
| `events tail` | Follow events as they arrive |

## Common flags

| Flag | Description |
| --- | --- |
| `--type <name>,<name>` | Filter by event type (repeatable). e.g. `INITIAL_PURCHASE,CANCELLATION` |
| `--since <when>` | RFC3339 (`2026-04-25T00:00:00Z`) or duration (`1h`, `30m`, `7d`) |
| `--limit <n>` | (`list` only) max page size |
| `--interval <dur>` | (`tail` only) poll interval, min `2s`, default `5s` |

## Examples

```sh
revcat events list --type INITIAL_PURCHASE --since 24h
revcat events tail --type INITIAL_PURCHASE,CANCELLATION
revcat events tail --since 1h --interval 5s

# ndjson when piped, perfect for jq
revcat events tail --output json | jq '.type'
```

The TTY view color-codes types: green for purchases, blue for trials, yellow for cancellations / expirations, red for refunds / billing issues.
