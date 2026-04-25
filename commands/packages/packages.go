// Package packages holds `revcat packages ...`.
//
// Packages live within offerings; this command list flattens them by
// fetching every offering's packages, optionally narrowed with --offering.
package packages

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/akshitkrnagpal/revcat/internal/api"
	"github.com/akshitkrnagpal/revcat/internal/cliutil"
	"github.com/akshitkrnagpal/revcat/internal/output"
)

var Cmd = &cobra.Command{
	Use:     "packages",
	Aliases: []string{"pkg"},
	Short:   "Inspect RevenueCat packages",
	Long: `Packages are the purchasable units inside an offering. Identifiers
follow RC's $rc_monthly / $rc_annual / custom convention.`,
}

func init() {
	Cmd.AddCommand(listCmd)
}

var listOffering string

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List packages across one offering or all offerings",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()

		offerings := []string{}
		if listOffering != "" {
			offerings = append(offerings, listOffering)
		} else {
			all, err := client.ListOfferings(ctx)
			if err != nil {
				return err
			}
			for _, o := range all {
				offerings = append(offerings, o.LookupKey)
			}
		}

		var rows [][]any
		var pkgs []api.Package
		for _, off := range offerings {
			ps, err := client.ListPackages(ctx, off)
			if err != nil {
				output.Warn("offering %q: %v", off, err)
				continue
			}
			for _, p := range ps {
				pkgs = append(pkgs, p)
				rows = append(rows, []any{off, p.Position, p.Identifier, dash(p.ProductID), formatTime(p.CreatedAt)})
			}
		}

		if output.IsJSON() {
			return output.JSON(pkgs)
		}
		return output.Table([]string{"offering", "#", "identifier", "product", "created"}, rows)
	},
}

func init() {
	listCmd.Flags().StringVarP(&listOffering, "offering", "o", "", "Restrict to a single offering by lookup_key")
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
