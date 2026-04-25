// Package publish holds the verb-orchestrators - top-level commands that
// compose multiple API calls behind one ergonomic flag set. The flagship
// command is `revcat publish offering`.
package publish

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
	Use:   "publish",
	Short: "One-shot deploy verbs (offering, paywall, ...)",
	Long: `Publish-style verbs compose several API calls behind a single command.
The intent is to mirror the dashboard's higher-level actions ("set as
current", "deploy paywall") rather than mirror REST endpoints.`,
}

func init() {
	Cmd.AddCommand(offeringCmd)
}
