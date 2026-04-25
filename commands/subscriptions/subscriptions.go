// Package subscriptions holds `revcat subscriptions ...`.
package subscriptions

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
	Use:     "subscriptions",
	Aliases: []string{"sub"},
	Short:   "Inspect and manage subscriptions",
	Long: `A subscription is one customer's ongoing purchase relationship with one
product. Find it via ` + "`revcat subscriptions search <store_id>`" + ` or by
listing it under a customer (` + "`revcat subscribers info`" + `).`,
}

func init() {
	Cmd.AddCommand(viewCmd, transactionsCmd, entitlementsCmd, cancelCmd, refundCmd, mgmtURLCmd, searchCmd)
}

var viewCmd = &cobra.Command{
	Use:   "view <id>",
	Short: "Show one subscription",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		s, err := client.GetSubscription(ctx, args[0])
		if err != nil {
			return err
		}
		if output.IsJSON() {
			return output.JSON(s)
		}
		rows := [][]any{
			{"id", s.ID},
			{"customer", cliutil.Dash(s.CustomerID)},
			{"product", s.ProductID},
			{"store", s.Store},
			{"status", s.Status},
			{"trial", s.IsTrial},
			{"sandbox", s.IsSandbox},
			{"will_renew", s.WillRenew},
			{"starts", cliutil.FormatTime(s.StartsAt)},
			{"current_ends", cliutil.FormatTime(s.CurrentEndsAt)},
		}
		return output.Table([]string{"field", "value"}, rows)
	},
}

var transactionsCmd = &cobra.Command{
	Use:   "transactions <id>",
	Short: "List billing transactions for a subscription",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		txs, err := client.ListSubscriptionTransactions(ctx, args[0])
		if err != nil {
			return err
		}
		rows := make([][]any, 0, len(txs))
		for _, t := range txs {
			rows = append(rows, []any{t.ID, t.Status, t.Amount, t.Currency, cliutil.FormatTime(t.OccurredAt)})
		}
		return output.Table([]string{"id", "status", "amount", "currency", "when"}, rows)
	},
}

var entitlementsCmd = &cobra.Command{
	Use:   "entitlements <id>",
	Short: "List entitlements granted by a subscription",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		ents, err := client.ListSubscriptionEntitlements(ctx, args[0])
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

var cancelConfirm bool

var cancelCmd = &cobra.Command{
	Use:   "cancel <id>",
	Short: "Cancel a subscription (Web Billing)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !cancelConfirm {
			var ok bool
			if err := survey.AskOne(&survey.Confirm{
				Message: "cancel subscription " + args[0] + "?",
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
		resp, err := client.CancelSubscription(ctx, args[0])
		if err != nil {
			return err
		}
		output.Success("cancelled %s", args[0])
		if output.IsJSON() {
			return output.JSON(resp)
		}
		return nil
	},
}

func init() {
	cancelCmd.Flags().BoolVarP(&cancelConfirm, "confirm", "y", false, "Skip the prompt")
}

var refundConfirm bool

var refundCmd = &cobra.Command{
	Use:   "refund <id>",
	Short: "Refund the entire subscription (Web Billing)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !refundConfirm {
			var ok bool
			if err := survey.AskOne(&survey.Confirm{
				Message: "refund subscription " + args[0] + "? this is irreversible.",
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
		resp, err := client.RefundSubscription(ctx, args[0])
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

var mgmtURLCmd = &cobra.Command{
	Use:   "management-url <id>",
	Short: "Print the store-specific manage/cancel URL",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		url, err := client.SubscriptionManagementURL(ctx, args[0])
		if err != nil {
			return err
		}
		output.Info("%s", url)
		return nil
	},
}

var searchCmd = &cobra.Command{
	Use:   "search <store_id>",
	Short: "Find subscriptions by store id (App Store / Play / Stripe / ...)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		subs, err := client.SearchSubscriptions(ctx, args[0])
		if err != nil {
			return err
		}
		if output.IsJSON() {
			return output.JSON(subs)
		}
		rows := make([][]any, 0, len(subs))
		for _, s := range subs {
			rows = append(rows, []any{s.ID, cliutil.Dash(s.CustomerID), s.ProductID, s.Status, s.Store, cliutil.FormatTime(s.StartsAt)})
		}
		return output.Table([]string{"id", "customer", "product", "status", "store", "started"}, rows)
	},
}
