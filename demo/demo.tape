#####################################################################
# revcat demo gif - the 5-second hero shot
#
# One moment of joy: ship a paywall update from the terminal.
#
# Regenerate with:
#   brew install vhs
#   cd demo
#   vhs demo.tape          # produces demo.gif
#
# Assumes:
#   - revcat is on $PATH (brew install akshitkrnagpal/tap/revcat)
#   - this directory has been bound to a project (cd demo && revcat init)
#     so revcat picks up project context from ./.revcat/config.json,
#     OR REVCAT_REFRESH_TOKEN + REVCAT_PROJECT_ID are set in the env
#   - the bound project has an offering called "pro" (or edit below)
#####################################################################

Output demo.gif

Set Shell "zsh"
Set FontSize 18
Set Width 1280
Set Height 480
Set Theme "Catppuccin Mocha"
Set Padding 24
Set TypingSpeed 35ms
Set PlaybackSpeed 1.0

# Custom prompt: simple $, no zsh continuation `>`.
Hide
Type 'export PS1="%F{12}$ %f"' Enter
Ctrl+L
Show

# 1. Frame the input. Show the paywall body that will ship.
Type "cat paywall.json | jq"
Sleep 200ms
Enter
Sleep 1500ms

# 2. Ship it. The orchestrator does both steps - PUT paywall + set
#    offering as current - in one shot. --dry-run prints the plan so
#    the demo is reproducible regardless of repo state.
Type "revcat publish offering pro --paywall ./paywall.json --dry-run --confirm"
Sleep 200ms
Enter
Sleep 2500ms

# 3. Confirm. Read-only, just shows that the offering is what the SDK
#    will see.
Type "revcat offerings view pro"
Sleep 200ms
Enter
Sleep 1500ms

Sleep 600ms
