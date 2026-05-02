---
title: charts
description: Project charts (revenue, active subs, conversion, etc.)
---

Charts mirror the dashboard graphs: revenue, active subscribers,
conversion, MRR, churn, etc. Run `revcat charts options <name>` for
the supported filters before requesting data.

## Subcommands

| Command | Description |
| --- | --- |
| `charts get <chart_name>` | Fetch chart data (raw JSON) |
| `charts options <chart_name>` | Show the available filters/dimensions for a chart |

Full flag reference: see [the CLI reference](/reference/cli/).
<!-- AUTOGEN_END -->

## `charts get` flags

| Flag | Description |
| --- | --- |
| `--start YYYY-MM-DD` | Range start |
| `--end YYYY-MM-DD` | Range end |
| `--period day\|week\|month` | Bucket granularity |
| `--filter key=value` | Filter on a dimension (repeatable) |

## Available chart names

`actives`, `actives_movement`, `actives_new`, `arr`, `churn`, `cohort_explorer`, `conversion_to_paying`, `customers_new`, `ltv_per_customer`, `ltv_per_paying_customer`, `mrr`, `mrr_movement`, `refund_rate`, `revenue`, `subscription_retention`, `subscription_status`, `trials`, `trials_movement`.

## Examples

```sh
revcat charts options actives
revcat charts get revenue --start 2026-04-01 --end 2026-04-30 --period day
revcat charts get actives --filter store=app_store
revcat charts get mrr --period month
```

`revcat charts get` returns the raw RC v2 chart payload.
