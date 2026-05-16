#####################################################################
# revcat customer-debug demo - the support-ticket flow
#
# A user pings support: "I bought premium but the app says I don't
# have access." The store gave them a transaction id but no app_user_id.
# Walks the diagnostic from store id back to the customer, lists
# entitlements, and shows the grant command commented out so this gif
# stays read-only.
#
# Regenerate with:
#   brew install vhs
#   cd demo
#   vhs customer-debug.tape    # produces customer-debug.gif
#
# Hermetic by default: uses ./mock-bin/revcat and never calls RevenueCat.
#####################################################################

Output customer-debug.gif

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

# 1. Support hands you a store transaction id. Resolve it to a sub.
#    This is the bridge: the dashboard requires an app_user_id, but
#    your support tooling only has the store id.
Type "revcat subscriptions search 1000000123456789"
Sleep 200ms
Enter
Sleep 1600ms

# 2. Now you have the customer. Pull their state.
Type "revcat subscribers info app_user_demo"
Sleep 200ms
Enter
Sleep 1800ms

# 3. Read-only entitlement list - confirms what you'd attach to.
Type "revcat entitlements list"
Sleep 200ms
Enter
Sleep 1600ms

# 4. Fix the user in one command. This is fixture output, not live state.
Type "revcat subscribers grant app_user_demo premium -d 7d --confirm"
Sleep 200ms
Enter
Sleep 1600ms

Sleep 600ms
