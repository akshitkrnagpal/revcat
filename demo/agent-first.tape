#####################################################################
# revcat agent-first demo - deterministic RevenueCat ops for agents
#
# Shows the core launch positioning: an agent can run exact commands from
# inside a repo, consume JSON, and leave replayable logs.
#
# Regenerate with:
#   brew install vhs
#   cd demo
#   vhs agent-first.tape     # produces agent-first.gif
#
# Hermetic by default: uses ./mock-bin/revcat and never calls RevenueCat.
#####################################################################

Output agent-first.gif

Set Shell "zsh"
Set FontSize 17
Set Width 1280
Set Height 680
Set Theme "Catppuccin Mocha"
Set Padding 24
Set TypingSpeed 24ms
Set PlaybackSpeed 1.0

Hide
Type 'export PS1="%F{12}$ %f"' Enter
Type 'export PATH="$PWD/mock-bin:$PATH"' Enter
Type 'setopt interactivecomments' Enter
Ctrl+L
Show

# 1. Frame the agent task without using a fake UI.
Type "# agent task: customer paid, but premium is missing"
Sleep 150ms
Enter
Sleep 450ms
Type "# repo: ~/your-app"
Sleep 150ms
Enter
Sleep 900ms

# 2. Confirm repo-local project context.
Type "revcat auth status"
Sleep 150ms
Enter
Sleep 1400ms

# 3. Resolve the store transaction id in a parseable form.
Type "revcat subscriptions search 1000000123456789 --output json"
Sleep 150ms
Enter
Sleep 1200ms

# 4. Pull the customer state for the next agent step.
Type "revcat subscribers info app_user_demo --output json"
Sleep 150ms
Enter
Sleep 1300ms

# 5. Leave a launch/debug gate that can fail CI.
Type "revcat doctor"
Sleep 150ms
Enter
Sleep 1500ms

Sleep 800ms
