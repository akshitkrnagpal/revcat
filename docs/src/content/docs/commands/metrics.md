---
title: metrics
description: Project-level revenue + subscription metrics.
---

## Subcommands

| Command | Description |
| --- | --- |
| `metrics overview` | Headline metrics for the active project (active subs, MRR, lifetime revenue, ...) |

For per-chart drill-downs see [`revcat charts get`](/commands/charts/).

## Example

```sh
revcat metrics overview
revcat metrics overview --output json | jq .
```
