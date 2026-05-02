// Package auth holds the `revcat auth ...` subcommand tree.
package auth

import "github.com/spf13/cobra"

// Cmd is the parent of all auth subcommands. Mounted on root in commands/root.go.
var Cmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage RevenueCat authentication",
	Long: `revcat authenticates against RevenueCat via OAuth. One browser login
writes the credential to ~/.revcat/config.json (mode 0600). A per-repo
.revcat/config.json (written by ` + "`revcat init`" + `) carries that credential
into the directory so agents and sandboxes inside the directory inherit
it without touching the global file.

Most users only need:

    revcat auth login                # browser OAuth
    cd ~/your/repo && revcat init    # bind this repo to a project
    revcat auth status

For CI / fresh sandboxes with no browser: set REVCAT_REFRESH_TOKEN
(and REVCAT_PROJECT_ID) to skip both file and login flow.`,
}

func init() {
	Cmd.AddCommand(loginCmd)
	Cmd.AddCommand(statusCmd)
	Cmd.AddCommand(doctorCmd)
	Cmd.AddCommand(useCmd)
	Cmd.AddCommand(logoutCmd)
	Cmd.AddCommand(listCmd)
}
