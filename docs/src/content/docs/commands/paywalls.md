---
title: paywalls
description: Manage top-level paywall resources.
---

Manage paywall records in the project's paywall library. v2 ties every paywall to exactly one offering, and the create endpoint accepts only `{offering_id}` - the actual paywall content (template, copy, components) is set later via the offering-scoped paywall config.

For most workflows you want [`revcat publish offering --paywall <file>`](/commands/publish/), which composes the paywall PUT and offering activation in one shot.

## Subcommands

| Command | Description |
| --- | --- |
| `paywalls list` | List paywalls in the project |
| `paywalls view <id>` | Show one paywall (raw JSON) |
| `paywalls create` | Create a paywall record (`--offering <id>` or `--file`) |
| `paywalls delete <id>` | Delete a paywall |

## Examples

```sh
revcat paywalls list
revcat paywalls view pw_xxx | jq .
revcat paywalls create --offering ofr_xxx
revcat paywalls delete pw_xxx -y
```

## Body shape (create via `--file`)

```json
{ "offering_id": "ofr_xxx" }
```

That's it - other fields are rejected by v2. To populate the paywall, use `revcat publish offering <offering_id> --paywall ./paywall.json`.
