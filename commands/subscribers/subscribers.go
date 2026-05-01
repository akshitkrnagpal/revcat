// Package subscribers holds the `revcat subscribers ...` subcommand tree.
package subscribers

import "github.com/spf13/cobra"

// Cmd is the parent of all subscribers subcommands.
var Cmd = &cobra.Command{
	Use:     "subscribers",
	Aliases: []string{"customers", "subs"},
	Short:   "Inspect and manage RevenueCat subscribers",
	Long: `Subscribers (a.k.a. customers, app users) are the end-users of your app.
revcat treats them as the unit of debugging - one command surfaces their
entitlements, subscriptions, purchases, and aliases in a single card.`,
}

func init() {
	Cmd.AddCommand(infoCmd)
}
