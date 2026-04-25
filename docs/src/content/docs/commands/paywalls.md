---
title: paywalls
description: Manage top-level paywall resources.
---

Manage paywall records in the project's paywall library. To deploy a paywall config to an offering use [`revcat publish offering --paywall <file>`](/commands/publish/).

## Subcommands

| Command | Description |
| --- | --- |
| `paywalls list` | List paywalls in the project |
| `paywalls view <id>` | Show one paywall (raw JSON) |
| `paywalls create` | Create a paywall from a JSON body (`--file`) |
| `paywalls delete <id>` | Delete a paywall |

## Example

```sh
revcat paywalls list
revcat paywalls view pw_xxx | jq .
revcat paywalls create --file ./paywalls/pro.json
```
