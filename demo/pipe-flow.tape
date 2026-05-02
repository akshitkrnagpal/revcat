#####################################################################
# revcat pipe-flow demo - the TTY-aware output flip
#
# Same command, two contexts. On a TTY revcat renders a colored table.
# The moment you pipe stdout, output flips to JSON automatically. No
# --json flag, no env knob, no opt-in. Pipe straight into jq, awk, or
# anything that speaks JSON.
#
# Anchor for the "the dashboard cannot" beat in the launch tweet.
#
# Regenerate with:
#   brew install vhs
#   export REVCAT_REFRESH_TOKEN=rtk_...
#   export REVCAT_PROJECT_ID=proj_...
#   cd demo
#   vhs pipe-flow.tape       # produces pipe-flow.gif
#####################################################################

Output pipe-flow.gif

Set Shell "zsh"
Set FontSize 17
Set Width 1280
Set Height 520
Set Theme "Catppuccin Mocha"
Set Padding 24
Set TypingSpeed 35ms
Set PlaybackSpeed 1.0

Hide
Type 'export PS1="%F{12}$ %f"' Enter
Ctrl+L
Show

# 1. On a TTY: tables, colors, lipgloss. The default human view.
Type "revcat offerings list"
Sleep 200ms
Enter
Sleep 2200ms

# 2. Pipe stdout. Same command, same flags - JSON, automatically.
#    No --json ceremony. Compose with jq from the same prompt.
Type "revcat offerings list | jq '.[].lookup_key'"
Sleep 200ms
Enter
Sleep 2200ms

Sleep 600ms
