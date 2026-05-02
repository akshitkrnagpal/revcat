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
#   export REVCAT_REFRESH_TOKEN=rtk_...
#   export REVCAT_PROJECT_ID=proj_...
#   cd demo
#   vhs init.tape          # produces init.gif
#####################################################################

Output init.gif

Set Shell "zsh"
Set FontSize 18
Set Width 1280
Set Height 700
Set Theme "Catppuccin Mocha"
Set Padding 24
Set TypingSpeed 35ms
Set PlaybackSpeed 1.0

# Setup: throwaway sandbox dir, simple prompt, project_id pre-resolved
# from the env so the survey picker doesn't gate the recording.
Hide
Type 'export PS1="%F{12}$ %f"' Enter
Type 'cd $(mktemp -d)' Enter
Ctrl+L
Show

# 1. Empty directory. Same shape every reader's repo starts as.
Type "ls -la"
Sleep 200ms
Enter
Sleep 800ms

# 2. Bind. --no-apps keeps the demo deterministic; the interactive flow
#    would also offer to tag specific apps from the project.
Type "revcat init --no-apps"
Sleep 200ms
Enter
Sleep 2500ms

# 3. Two files now exist. revcat.toml is the committed half.
Type "ls -la"
Sleep 200ms
Enter
Sleep 1200ms

Type "cat revcat.toml"
Sleep 200ms
Enter
Sleep 1500ms

# 4. The agent / sandbox proof. Wipe the env vars revcat used to auth
#    in the first place, then run a project-scoped command. It still
#    works - the walked-up .revcat/config.json carries credentials
#    plus project_id, so any process started in this directory inherits
#    both without touching the global file.
Type "unset REVCAT_REFRESH_TOKEN REVCAT_PROJECT_ID"
Sleep 200ms
Enter
Sleep 600ms

Type "revcat offerings list"
Sleep 200ms
Enter
Sleep 2200ms

Sleep 600ms
