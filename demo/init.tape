#####################################################################
# revcat init demo - the per-repo bootstrap flow
#
# Shows binding a fresh repo to a RevenueCat project. Two files
# materialize: revcat.toml (committed, project_id) and
# .revcat/config.json (gitignored, mode 0600, credentials + project_id).
# After this, every revcat command run in the directory inherits the
# project context automatically.
#
# Regenerate with:
#   brew install vhs
#   cd demo
#   vhs init.tape          # produces init.gif
#
# Hermetic by default: uses ./mock-bin/revcat and never calls RevenueCat.
#####################################################################

Output init.gif

Set Shell "zsh"
Set FontSize 17
Set Width 1280
Set Height 600
Set Theme "Catppuccin Mocha"
Set Padding 24
Set TypingSpeed 28ms
Set PlaybackSpeed 1.0

# Setup: throwaway sandbox dir, simple prompt, mock revcat on PATH.
Hide
Type 'export PS1="%F{12}$ %f"' Enter
Type 'export DEMO_DIR="$PWD"; export PATH="$DEMO_DIR/mock-bin:$PATH"' Enter
Type 'cd $(mktemp -d)' Enter
Ctrl+L
Show

# 1. Empty directory. Same shape every reader's repo starts as.
Type "find . -maxdepth 2 -type f | sort"
Sleep 200ms
Enter
Sleep 800ms

# 2. Bind. --no-apps keeps the demo deterministic; the interactive flow
#    would also offer to tag specific apps from the project.
Type "revcat init --no-apps"
Sleep 200ms
Enter
Sleep 2200ms

# 3. Two files now exist. revcat.toml is the committed half.
Type "find . -maxdepth 2 -type f | sort"
Sleep 200ms
Enter
Sleep 1000ms

Type "cat revcat.toml"
Sleep 200ms
Enter
Sleep 1200ms

# 4. Project context is now automatic. No --project-id, no env, just
#    works from cwd.
Type "revcat auth status"
Sleep 200ms
Enter
Sleep 1800ms

Sleep 600ms
