---
title: paywalls
description: Manage top-level paywall resources
---

Manage paywall records in the project's paywall library. To deploy a
paywall config to an offering use `revcat publish offering --paywall <file>`.

## Subcommands

| Command | Description |
| --- | --- |
| `paywalls create` | Create a paywall record scoped to an offering |
| `paywalls delete <id>` | Delete a paywall |
| `paywalls list` | List paywalls in the project |
| `paywalls view <id>` | Show one paywall (raw JSON) |

Full flag reference: see [the CLI reference](/reference/cli/).
<!-- AUTOGEN_END -->

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
