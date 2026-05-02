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
Set Height 560
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

# 4. Project context is now automatic. No --project-id, no env, just
#    works from cwd.
Type "revcat auth status"
Sleep 200ms
Enter
Sleep 2000ms

Sleep 600ms
