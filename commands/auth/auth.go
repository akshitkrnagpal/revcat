// Package auth holds the `revcat auth ...` subcommand tree.
package auth

import "github.com/spf13/cobra"

// Cmd is the parent of all auth subcommands. Mounted on root in commands/root.go.
var Cmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage RevenueCat authentication",
	Long: `revcat authenticates against RevenueCat via OAuth. One browser login
populates a global profile in your OS keychain; a per-repo
.revcat/config.json (written by ` + "`revcat init`" + `) carries that credential
into the directory so agents and sandboxes work without keychain access.

Most users only need:

    revcat auth login            # browser OAuth, saves to keychain
    cd ~/your/repo && revcat init   # bind this repo to a project
    revcat auth status

For Linux containers without secret-service, pass --bypass-keychain
(or set REVCAT_BYPASS_KEYCHAIN=1) to use ~/.revcat/config.json instead.

For CI / fresh sandboxes with no browser: set REVCAT_REFRESH_TOKEN
(and REVCAT_PROJECT_ID) to skip both keychain and login flow.`,
}

func init() {
	Cmd.AddCommand(loginCmd)
	Cmd.AddCommand(statusCmd)
	Cmd.AddCommand(doctorCmd)
	Cmd.AddCommand(useCmd)
	Cmd.AddCommand(logoutCmd)
	Cmd.AddCommand(listCmd)
}

// bypass is wired up in each subcommand so they can read the global flag
// from rootCmd via cmd.Root().PersistentFlags(). Helper centralises it.
func bypassKeychain(cmd *cobra.Command) bool {
	flag := cmd.Root().PersistentFlags().Lookup("bypass-keychain")
	if flag == nil {
		return false
	}
	return flag.Value.String() == "true"
}
