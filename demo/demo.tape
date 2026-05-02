#####################################################################
# revcat demo gif
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
#   - the bound project has at least one offering called "default"
#####################################################################

Output demo.gif

Set Shell "zsh"
Set FontSize 18
Set Width 1280
Set Height 720
Set Theme "Catppuccin Mocha"
Set Padding 24
Set TypingSpeed 35ms
Set PlaybackSpeed 1.0

# Custom prompt: simple $, no zsh continuation `>`.
Hide
Type 'export PS1="%F{12}$ %f"' Enter
Ctrl+L
Show

Type "revcat doctor --output table"
Sleep 200ms
Enter
Sleep 3s

Type "clear"
Enter
Sleep 200ms

# Catalog read
Type "revcat offerings list"
Sleep 200ms
Enter
Sleep 3s

Type "clear"
Enter
Sleep 200ms

# Catalog: entitlements
Type "revcat entitlements list"
Sleep 200ms
Enter
Sleep 2.5s

Type "clear"
Enter
Sleep 200ms

# The verb-orchestrator: dry-run a paywall publish.
Type "revcat publish offering default --paywall ./paywall.json --dry-run --confirm"
Sleep 300ms
Enter
Sleep 4s

Type "clear"
Enter
Sleep 200ms

Sleep 800ms
