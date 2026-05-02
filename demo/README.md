# revcat demos

Three [vhs](https://github.com/charmbracelet/vhs) tapes, each rendering one user-facing flow. `demo.gif` is the README hero; the other two anchor specific guide pages.

| Tape                    | Output                | What it shows |
| ---                     | ---                   | --- |
| `demo.tape`             | `demo.gif`            | Hero: ship a paywall update + set offering current with one orchestrator command. |
| `init.tape`             | `init.gif`            | Bootstrap: `revcat init` materializes `revcat.toml` + `.revcat/config.json` so subsequent commands inherit project context. |
| `customer-debug.tape`   | `customer-debug.gif`  | Support flow: resolve a store transaction id back to a customer, pull their state, surface the entitlements catalog. Read-only; the grant command is shown as a comment. |

## Regenerate

```sh
brew install vhs

# revcat must be on $PATH. Either bind a project (`revcat init` in this
# directory) so revcat picks up project context from ./.revcat/config.json,
# OR set the env hatch for a one-off:
export REVCAT_REFRESH_TOKEN=rtk_...
export REVCAT_PROJECT_ID=proj_...

cd demo
vhs demo.tape            # produces demo.gif
vhs init.tape            # produces init.gif
vhs customer-debug.tape  # produces customer-debug.gif
```

The `customer-debug.tape` ids are placeholders (`1000000123456789`, `app_user_demo`, `premium`). Edit them to match real ids in your test project before recording. The other two tapes work against any project that has an offering called `default` (for the hero) or just any authed credential (for init).

`paywall.json` is a sample paywall body kept in this folder so the hero tape is self-contained.
