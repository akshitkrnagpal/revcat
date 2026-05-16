# revcat demos

Six [vhs](https://github.com/charmbracelet/vhs) tapes, each rendering one user-facing flow. `demo.gif` is the README hero; the other GIFs are marketing/support cuts for specific use cases.

All tapes are hermetic by default: they prepend `./mock-bin` to `PATH`, so they show real revcat command syntax without calling RevenueCat or requiring credentials.

| Tape                    | Output                | What it shows |
| ---                     | ---                   | --- |
| `demo.tape`             | `demo.gif`            | Hero: ship a paywall update + set offering current with one orchestrator command. |
| `agent-first.tape`      | `agent-first.gif`     | Agent-first: repo-local context, parseable JSON, and a replayable health gate. |
| `init.tape`             | `init.gif`            | Bootstrap: `revcat init` materializes `revcat.toml` + `.revcat/config.json` so subsequent commands inherit project context. |
| `customer-debug.tape`   | `customer-debug.gif`  | Support flow: resolve a store transaction id back to a customer, inspect access, then grant goodwill access. |
| `catalog.tape`          | `catalog.gif`         | Catalog inspection: offerings, products, and package membership. |
| `ops.tape`              | `ops.gif`             | Ops checks: webhooks, audit logs, and headline metrics. |

## Regenerate

```sh
brew install vhs

cd demo
vhs demo.tape            # produces demo.gif
vhs agent-first.tape     # produces agent-first.gif
vhs init.tape            # produces init.gif
vhs customer-debug.tape  # produces customer-debug.gif
vhs catalog.tape         # produces catalog.gif
vhs ops.tape             # produces ops.gif
```

For social video uploads, duplicate any tape and change `Output something.gif` to `Output something.mp4`; VHS will render an MP4 with the same scene.

`paywall.json` is a sample paywall body kept in this folder so the hero tape is self-contained. `mock-bin/revcat` owns the deterministic fixture output; update it when a tape adds a new command.
