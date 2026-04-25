package subscribers

import (
	"context"
	"errors"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"

	"github.com/akshitkrnagpal/revcat/internal/cliutil"
	"github.com/akshitkrnagpal/revcat/internal/output"
)

var (
	refundTransaction string
	refundConfirm     bool
)

var refundCmd = &cobra.Command{
	Use:   "refund <user_id>",
	Short: "Refund a transaction for a subscriber",
	Long: `Issue a refund through the appropriate store (App Store, Play Store,
Stripe, ...). Pass --transaction <id> to identify which one; if the
subscriber only has one refundable transaction we'll find it.

Refund availability depends on the store and the original purchase date.
RC returns the updated transaction status; we surface it in JSON if you
pipe the output.`,
	Args: cobra.ExactArgs(1),
	RunE: runRefund,
}

func init() {
	refundCmd.Flags().StringVarP(&refundTransaction, "transaction", "t", "", "Transaction id (required)")
	refundCmd.Flags().BoolVarP(&refundConfirm, "confirm", "y", false, "Skip the confirmation prompt")
	_ = refundCmd.MarkFlagRequired("transaction")
	Cmd.AddCommand(refundCmd)
}

func runRefund(cmd *cobra.Command, args []string) error {
	customerID := args[0]

	client, _, err := cliutil.Client(cmd)
	if err != nil {
		return err
	}

	output.Info("plan: refund transaction %q for %q", refundTransaction, customerID)
	if !refundConfirm {
		var ok bool
		if err := survey.AskOne(&survey.Confirm{
			Message: "this is irreversible. apply?",
			Default: false,
		}, &ok); err != nil {
			return err
		}
		if !ok {
			return errors.New("aborted")
		}
	}

	ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
	defer cancel()
	resp, err := client.RefundTransaction(ctx, customerID, refundTransaction)
	if err != nil {
		return err
	}
	if output.IsJSON() {
		return output.JSON(resp)
	}
	output.Success("refund issued")
	if status, ok := resp["status"].(string); ok {
		output.Info("  status: %s", status)
	}
	return nil
}
