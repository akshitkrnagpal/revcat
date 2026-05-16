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
# Hermetic by default:
#   - prepends ./mock-bin to PATH so this recording never calls RevenueCat
#   - shows real revcat command syntax against deterministic fixture output
#####################################################################

Output demo.gif

Set Shell "zsh"
Set FontSize 17
Set Width 1280
Set Height 560
Set Theme "Catppuccin Mocha"
Set Padding 24
Set TypingSpeed 28ms
Set PlaybackSpeed 1.0

# Custom prompt: simple $, no zsh continuation `>`.
Hide
Type 'export PS1="%F{12}$ %f"' Enter
Type 'export PATH="$PWD/mock-bin:$PATH"' Enter
Ctrl+L
Show

# 1. Show a compact summary of the paywall body that will ship.
Type "jq '{title: .config.title, cta: .config.cta, packages: .config.package_ids}' paywall.json"
Sleep 200ms
Enter
Sleep 1400ms

# 2. Ship it. The orchestrator does both steps - PUT paywall + set
#    offering as current - in one shot. --dry-run prints the plan so
#    the demo is reproducible regardless of repo state.
Type "revcat publish offering default --paywall ./paywall.json --dry-run --confirm"
Sleep 200ms
Enter
Sleep 2200ms

# 3. Confirm. Read-only, just shows that the offering is what the SDK
#    will see.
Type "revcat offerings view default"
Sleep 200ms
Enter
Sleep 1600ms

Sleep 600ms
