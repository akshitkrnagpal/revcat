#####################################################################
# revcat ops demo - audit logs, webhooks, and headline metrics
#
# Shows operational checks that are awkward to hop through the dashboard
# for during support or launch monitoring.
#
# Regenerate with:
#   brew install vhs
#   cd demo
#   vhs ops.tape           # produces ops.gif
#
# Hermetic by default: uses ./mock-bin/revcat and never calls RevenueCat.
#####################################################################

Output ops.gif

Set Shell "zsh"
Set FontSize 17
Set Width 1280
Set Height 600
Set Theme "Catppuccin Mocha"
Set Padding 24
Set TypingSpeed 28ms
Set PlaybackSpeed 1.0

Hide
Type 'export PS1="%F{12}$ %f"' Enter
Type 'export PATH="$PWD/mock-bin:$PATH"' Enter
Ctrl+L
Show

Type "revcat webhooks list"
Sleep 200ms
Enter
Sleep 1700ms

Type "revcat audit-logs list"
Sleep 200ms
Enter
Sleep 1700ms

Type "revcat metrics overview"
Sleep 200ms
Enter
Sleep 1700ms

Sleep 600ms
