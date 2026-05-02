// Package collaborators holds `revcat collaborators ...` - read access
// to project membership.
package collaborators

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/akshitkrnagpal/revcat/internal/cliutil"
	"github.com/akshitkrnagpal/revcat/internal/output"
)

// Cmd is the parent of the collaborators subcommands. Mounted at root
// in commands/root.go.
var Cmd = &cobra.Command{
	Use:     "collaborators",
	Aliases: []string{"members"},
	Short:   "Inspect project collaborators (members)",
	Long: `List the people with access to the active RevenueCat project.

Read-only: v2 doesn't expose invite / role-change / remove via REST.
Manage membership in the dashboard.`,
}

func init() {
	Cmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List collaborators on the active project",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		members, err := client.ListCollaborators(ctx)
		if err != nil {
			return err
		}
		rows := make([][]any, 0, len(members))
		for _, m := range members {
			rows = append(rows, []any{
				m.ID,
				dash(m.Name),
				m.Email,
				dash(m.Role),
				accepted(m.AcceptedAt),
				yesNo(m.HasMFA),
			})
		}
		return output.Table([]string{"id", "name", "email", "role", "accepted", "mfa"}, rows)
	},
}

func dash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

// accepted converts a nullable accepted_at unix-ms into a date string
// or "pending" if the invite has not been accepted yet.
func accepted(ms int64) string {
	if ms == 0 {
		return "pending"
	}
	return time.UnixMilli(ms).UTC().Format("2006-01-02")
}

func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
