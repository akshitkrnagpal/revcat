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

var revokeConfirm bool

var revokeCmd = &cobra.Command{
	Use:   "revoke <user_id> <entitlement>",
	Short: "Revoke a promotional entitlement from a subscriber",
	Long: `Remove a promotional entitlement that was previously granted via
` + "`revcat subscribers grant`" + ` or the dashboard. Does not affect
entitlements granted by an active store subscription.`,
	Args: cobra.ExactArgs(2),
	RunE: runRevoke,
}

func init() {
	revokeCmd.Flags().BoolVarP(&revokeConfirm, "confirm", "y", false, "Skip the confirmation prompt")
	Cmd.AddCommand(revokeCmd)
}

func runRevoke(cmd *cobra.Command, args []string) error {
	customerID, entitlementID := args[0], args[1]

	client, _, err := cliutil.Client(cmd)
	if err != nil {
		return err
	}

	output.Info("plan: revoke %q from %q", entitlementID, customerID)
	if !revokeConfirm {
		var ok bool
		if err := survey.AskOne(&survey.Confirm{Message: "apply?", Default: false}, &ok); err != nil {
			return err
		}
		if !ok {
			return errors.New("aborted")
		}
	}

	ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
	defer cancel()
	if err := client.RevokePromotionalEntitlement(ctx, customerID, entitlementID); err != nil {
		return err
	}
	output.Success("revoked")
	return nil
}
