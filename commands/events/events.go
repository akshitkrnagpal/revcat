// Package events holds `revcat events ...`.
package events

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
	Use:   "events",
	Short: "Inspect RevenueCat events",
	Long: `Events is the firehose of subscription lifecycle activity in your project:
purchases, renewals, cancellations, trials, refunds, etc.

Use ` + "`revcat events list`" + ` for a one-shot page, or ` + "`revcat events tail`" + `
to follow new events as they arrive (kubectl-logs-style).`,
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(tailCmd)
}
