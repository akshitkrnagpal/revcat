// Package auth holds the `revcat auth ...` subcommand tree.
package auth

import "github.com/spf13/cobra"

// Cmd is the parent of all auth subcommands. Mounted on root in commands/root.go.
var Cmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage RevenueCat authentication",
	Long: `revcat stores credentials in your OS keychain by default. Each set of
credentials is a "profile" with a name, secret key, and project id.

Most users only need:

    revcat auth login --name my-app --secret-key sk_...
    revcat auth status

For CI, pass --bypass-keychain (or set REVCAT_BYPASS_KEYCHAIN=1) to use a
local file instead, or pass REVCAT_API_KEY directly.`,
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
