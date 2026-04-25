// Package purchases holds `revcat purchases ...`.
package purchases

import (
	"context"
	"errors"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"

	"github.com/akshitkrnagpal/revcat/internal/cliutil"
	"github.com/akshitkrnagpal/revcat/internal/output"
)

var Cmd = &cobra.Command{
	Use:     "purchases",
	Aliases: []string{"purchase"},
	Short:   "Inspect non-renewing purchases",
	Long: `A purchase is a one-shot non-renewing transaction (lifetime grants,
consumables, in-app one-offs). For subscriptions see ` + "`revcat subscriptions`" + `.`,
}

func init() {
	Cmd.AddCommand(viewCmd, entitlementsCmd, refundCmd, searchCmd)
}

var viewCmd = &cobra.Command{
	Use:   "view <id>",
	Short: "Show one purchase",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		p, err := client.GetPurchase(ctx, args[0])
		if err != nil {
			return err
		}
		if output.IsJSON() {
			return output.JSON(p)
		}
		rows := [][]any{
			{"id", p.ID},
			{"customer", cliutil.Dash(p.CustomerID)},
			{"product", p.ProductID},
			{"store", p.Store},
			{"sandbox", p.IsSandbox},
			{"purchased", cliutil.FormatTime(p.PurchaseAt)},
		}
		if p.Amount > 0 {
			rows = append(rows, []any{"amount", p.Amount})
			rows = append(rows, []any{"currency", p.Currency})
		}
		return output.Table([]string{"field", "value"}, rows)
	},
}

var entitlementsCmd = &cobra.Command{
	Use:   "entitlements <id>",
	Short: "List entitlements granted by a purchase",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		ents, err := client.ListPurchaseEntitlements(ctx, args[0])
		if err != nil {
			return err
		}
		rows := make([][]any, 0, len(ents))
		for _, e := range ents {
			rows = append(rows, []any{e.LookupKey, e.DisplayName})
		}
		return output.Table([]string{"id", "display_name"}, rows)
	},
}

var refundConfirm bool

var refundCmd = &cobra.Command{
	Use:   "refund <id>",
	Short: "Refund a non-renewing purchase (Web Billing)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !refundConfirm {
			var ok bool
			if err := survey.AskOne(&survey.Confirm{
				Message: "refund purchase " + args[0] + "? this is irreversible.",
				Default: false,
			}, &ok); err != nil {
				return err
			}
			if !ok {
				return errors.New("aborted")
			}
		}
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		resp, err := client.RefundPurchase(ctx, args[0])
		if err != nil {
			return err
		}
		output.Success("refund issued for %s", args[0])
		if output.IsJSON() {
			return output.JSON(resp)
		}
		return nil
	},
}

func init() {
	refundCmd.Flags().BoolVarP(&refundConfirm, "confirm", "y", false, "Skip the prompt")
}

var searchCmd = &cobra.Command{
	Use:   "search <store_id>",
	Short: "Find purchases by store id",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		ps, err := client.SearchPurchases(ctx, args[0])
		if err != nil {
			return err
		}
		if output.IsJSON() {
			return output.JSON(ps)
		}
		rows := make([][]any, 0, len(ps))
		for _, p := range ps {
			rows = append(rows, []any{p.ID, cliutil.Dash(p.CustomerID), p.ProductID, p.Store, cliutil.FormatTime(p.PurchaseAt)})
		}
		return output.Table([]string{"id", "customer", "product", "store", "purchased"}, rows)
	},
}
