// Package entitlements holds `revcat entitlements ...`.
package entitlements

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/akshitkrnagpal/revcat/internal/cliutil"
	"github.com/akshitkrnagpal/revcat/internal/output"
)

var Cmd = &cobra.Command{
	Use:     "entitlements",
	Aliases: []string{"ent"},
	Short:   "Manage RevenueCat entitlements",
	Long: `Entitlements are project-level access flags (e.g., "premium", "pro").
Customers gain entitlements via products attached on offerings, or via
promotional grants. Use ` + "`revcat subscribers info <user_id>`" + ` to see what
a specific customer has.`,
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(viewCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all entitlements in the active project",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		ents, err := client.ListEntitlements(ctx)
		if err != nil {
			return err
		}
		rows := make([][]any, 0, len(ents))
		for _, e := range ents {
			rows = append(rows, []any{e.LookupKey, e.DisplayName, formatTime(e.CreatedAt)})
		}
		return output.Table([]string{"id", "display_name", "created"}, rows)
	},
}

var viewCmd = &cobra.Command{
	Use:   "view <id>",
	Short: "Show one entitlement by lookup_key",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		e, err := client.GetEntitlement(ctx, args[0])
		if err != nil {
			return err
		}
		if output.IsJSON() {
			return output.JSON(e)
		}
		rows := [][]any{
			{"id", e.LookupKey},
			{"display_name", e.DisplayName},
			{"created", formatTime(e.CreatedAt)},
			{"internal_id", e.ID},
		}
		return output.Table([]string{"field", "value"}, rows)
	},
}

func formatTime(unix int64) string {
	if unix == 0 {
		return "-"
	}
	t := time.Unix(unix, 0).UTC()
	if unix > 9999999999 {
		t = time.UnixMilli(unix).UTC()
	}
	return t.Format("2006-01-02")
}
