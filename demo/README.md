# revcat demo

`demo.gif` is the README hero. Generated from `demo.tape` via [vhs](https://github.com/charmbracelet/vhs).

## Regenerate

```sh
brew install vhs

# revcat must be on $PATH and a profile must exist that points at a
# project with at least one offering called "default".
export REVCAT_API_KEY=sk_...
export REVCAT_PROJECT_ID=proj_...

cd demo
vhs demo.tape          # produces demo.gif
```

## What's in the demo

1. `revcat doctor --output table` - quick health check, all green.
2. `revcat offerings list` - catalog readback.
3. `revcat entitlements list` - catalog readback.
4. `revcat publish offering default --paywall ./paywall.json --dry-run --confirm` - the verb-orchestrator, prints the plan without mutating.

`paywall.json` is a sample paywall body kept in this folder so the demo is self-contained.
