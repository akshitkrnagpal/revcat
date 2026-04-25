// Package offerings holds `revcat offerings ...`.
package offerings

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/akshitkrnagpal/revcat/internal/cliutil"
	"github.com/akshitkrnagpal/revcat/internal/output"
)

var Cmd = &cobra.Command{
	Use:     "offerings",
	Aliases: []string{"offer"},
	Short:   "Manage RevenueCat offerings",
	Long: `An offering is a presentation grouping of packages displayed on a
paywall. Each project has 0..N offerings; exactly one is "current" and is
returned by SDKs that ask for the current offering.`,
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(viewCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all offerings in the active project",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		offers, err := client.ListOfferings(ctx)
		if err != nil {
			return err
		}
		rows := make([][]any, 0, len(offers))
		for _, o := range offers {
			marker := ""
			if o.IsCurrent {
				marker = "*"
			}
			rows = append(rows, []any{marker, o.LookupKey, o.DisplayName, formatTime(o.CreatedAt)})
		}
		return output.Table([]string{"", "id", "display_name", "created"}, rows)
	},
}

var viewWithPackages bool

var viewCmd = &cobra.Command{
	Use:   "view <id>",
	Short: "Show one offering by lookup_key",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		o, err := client.GetOffering(ctx, args[0], viewWithPackages || !output.IsJSON())
		if err != nil {
			return err
		}
		if output.IsJSON() {
			return output.JSON(o)
		}
		current := "no"
		if o.IsCurrent {
			current = "yes (current)"
		}
		head := [][]any{
			{"id", o.LookupKey},
			{"display_name", o.DisplayName},
			{"is_current", current},
			{"created", formatTime(o.CreatedAt)},
			{"package count", len(o.Packages)},
		}
		if err := output.Table([]string{"field", "value"}, head); err != nil {
			return err
		}
		if len(o.Packages) > 0 {
			pkgRows := make([][]any, 0, len(o.Packages))
			for _, p := range o.Packages {
				pkgRows = append(pkgRows, []any{p.Position, p.Identifier, dash(p.ProductID)})
			}
			return output.Table([]string{"#", "identifier", "product"}, pkgRows)
		}
		return nil
	},
}

func init() {
	viewCmd.Flags().BoolVar(&viewWithPackages, "packages", true, "Include packages in the rendered card (always on for table output)")
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

func dash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
