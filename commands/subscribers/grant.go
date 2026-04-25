package subscribers

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"

	"github.com/akshitkrnagpal/revcat/internal/api"
	"github.com/akshitkrnagpal/revcat/internal/cliutil"
	"github.com/akshitkrnagpal/revcat/internal/output"
)

var (
	grantDuration string
	grantConfirm  bool
)

var grantCmd = &cobra.Command{
	Use:   "grant <user_id> <entitlement>",
	Short: "Grant a promotional entitlement to a subscriber",
	Long: `Grant a promotional entitlement (audited) to a subscriber for a
specified duration. v2 stores absolute expiry timestamps; revcat
translates --duration to the right "expires_at".

Duration accepts:
    forever | lifetime         (~100 years)
    7d, 30d, 90d               (days)
    1m, 3m, 6m                 (months ~ 30 days each)
    1y, 2y, 5y                 (years ~ 365 days each)

Examples:
    revcat subscribers grant app_user_123 premium --duration 7d
    revcat subscribers grant app_user_123 pro --duration lifetime --confirm`,
	Args: cobra.ExactArgs(2),
	RunE: runGrant,
}

func init() {
	grantCmd.Flags().StringVarP(&grantDuration, "duration", "d", "", "How long the grant lasts (required)")
	grantCmd.Flags().BoolVarP(&grantConfirm, "confirm", "y", false, "Skip the confirmation prompt")
	_ = grantCmd.MarkFlagRequired("duration")
	Cmd.AddCommand(grantCmd)
}

func runGrant(cmd *cobra.Command, args []string) error {
	customerID, entitlementID := args[0], args[1]

	expiresAt, label, err := durationToExpiresAt(grantDuration)
	if err != nil {
		return err
	}

	client, _, err := cliutil.Client(cmd)
	if err != nil {
		return err
	}

	output.Info("plan: grant %q to %q for %s (expires %s)", entitlementID, customerID, label, time.UnixMilli(expiresAt).Local().Format("2006-01-02 15:04 MST"))
	if !grantConfirm {
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
	if _, err := client.GrantEntitlement(ctx, customerID, api.GrantEntitlementRequest{
		EntitlementID: entitlementID,
		ExpiresAt:     expiresAt,
	}); err != nil {
		return err
	}
	output.Success("granted")
	return nil
}

// durationToExpiresAt converts a human-friendly duration string into an
// absolute Unix-millisecond expiry. Returns the expiry, a human label
// for the plan line, and any parse error.
func durationToExpiresAt(s string) (int64, string, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	now := time.Now()
	switch s {
	case "forever", "lifetime":
		return now.AddDate(100, 0, 0).UnixMilli(), "lifetime", nil
	}
	if len(s) < 2 {
		return 0, "", fmt.Errorf("invalid duration %q", s)
	}
	unit := s[len(s)-1]
	num := s[:len(s)-1]
	n, err := strconv.Atoi(num)
	if err != nil || n <= 0 {
		return 0, "", fmt.Errorf("invalid duration %q (try 7d, 1m, 1y, or lifetime)", s)
	}
	switch unit {
	case 'd':
		return now.AddDate(0, 0, n).UnixMilli(), fmt.Sprintf("%d day(s)", n), nil
	case 'm':
		return now.AddDate(0, n, 0).UnixMilli(), fmt.Sprintf("%d month(s)", n), nil
	case 'y':
		return now.AddDate(n, 0, 0).UnixMilli(), fmt.Sprintf("%d year(s)", n), nil
	}
	return 0, "", fmt.Errorf("invalid duration %q (try 7d, 1m, 1y, or lifetime)", s)
}
