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

var refundConfirm bool

var refundCmd = &cobra.Command{
	Use:   "refund <subscription_id> <transaction_id>",
	Short: "Refund a transaction on a subscription",
	Long: `Issue a refund through the appropriate store (App Store, Play Store,
Stripe, ...). v2 scopes refunds under the subscription, so both ids
are required.

Find the subscription id with ` + "`revcat subscribers info <user_id>`" + ` -
each subscription row carries an internal id.

Refund availability depends on the store and the original purchase date.
RC returns the updated transaction status; we surface it in JSON if you
pipe the output.`,
	Args: cobra.ExactArgs(2),
	RunE: runRefund,
}

func init() {
	refundCmd.Flags().BoolVarP(&refundConfirm, "confirm", "y", false, "Skip the confirmation prompt")
	Cmd.AddCommand(refundCmd)
}

func runRefund(cmd *cobra.Command, args []string) error {
	subscriptionID, transactionID := args[0], args[1]

	client, _, err := cliutil.Client(cmd)
	if err != nil {
		return err
	}

	output.Info("plan: refund transaction %q on subscription %q", transactionID, subscriptionID)
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
	resp, err := client.RefundTransaction(ctx, subscriptionID, transactionID)
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
