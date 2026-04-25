---
title: charts
description: Project charts (revenue, active subs, conversion, ...).
---

Charts mirror the dashboard graphs: revenue, active subscribers, conversion, MRR, churn, etc. Run `revcat charts options <name>` for the supported filters before requesting data.

## Subcommands

| Command | Description |
| --- | --- |
| `charts get <name>` | Fetch chart data (raw JSON) |
| `charts options <name>` | Show the available filters / dimensions for a chart |

## `charts get` flags

| Flag | Description |
| --- | --- |
| `--start YYYY-MM-DD` | Range start |
| `--end YYYY-MM-DD` | Range end |
| `--period day\|week\|month` | Bucket granularity |
| `--filter key=value` | Filter on a dimension (repeatable) |

## Examples

```sh
revcat charts options revenue
revcat charts get revenue --start 2026-04-01 --end 2026-04-30 --period day
revcat charts get active_subscribers --filter store=app_store
```

`revcat charts get` returns the raw RC v2 chart payload.
