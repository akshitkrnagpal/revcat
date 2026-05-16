#####################################################################
# revcat catalog demo - inspect RevenueCat catalog relationships
#
# Shows offerings, products, and package membership without opening the
# RevenueCat dashboard.
#
# Regenerate with:
#   brew install vhs
#   cd demo
#   vhs catalog.tape       # produces catalog.gif
#
# Hermetic by default: uses ./mock-bin/revcat and never calls RevenueCat.
#####################################################################

Output catalog.gif

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

Type "revcat offerings list"
Sleep 200ms
Enter
Sleep 1600ms

Type "revcat products list"
Sleep 200ms
Enter
Sleep 1600ms

Type "revcat packages list --offering default"
Sleep 200ms
Enter
Sleep 1600ms

Sleep 600ms
