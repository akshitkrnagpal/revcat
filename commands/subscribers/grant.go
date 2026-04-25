package subscribers

import (
	"context"
	"errors"
	"fmt"
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
	grantReason   string
	grantConfirm  bool
)

var grantCmd = &cobra.Command{
	Use:   "grant <user_id> <entitlement>",
	Short: "Grant a promotional entitlement to a subscriber",
	Long: `Grant a promotional entitlement (audited) to a subscriber for a
specified duration.

Duration accepts:
    forever | lifetime
    7d, 30d, 90d
    1m, 3m, 6m, 1y, 2y
    daily, weekly, monthly, yearly  (RC built-in plans)

Examples:
    revcat subscribers grant app_user_123 premium --duration 7d --reason "support ticket #2241"
    revcat subscribers grant app_user_123 pro --duration lifetime --confirm`,
	Args: cobra.ExactArgs(2),
	RunE: runGrant,
}

func init() {
	grantCmd.Flags().StringVarP(&grantDuration, "duration", "d", "", "How long the grant lasts (required). e.g. 7d, 30d, 1m, 1y, lifetime")
	grantCmd.Flags().StringVarP(&grantReason, "reason", "r", "", "Audit reason recorded on the grant")
	grantCmd.Flags().BoolVarP(&grantConfirm, "confirm", "y", false, "Skip the confirmation prompt")
	_ = grantCmd.MarkFlagRequired("duration")
	Cmd.AddCommand(grantCmd)
}

func runGrant(cmd *cobra.Command, args []string) error {
	customerID, entitlementID := args[0], args[1]

	rcDuration, label, err := normalizeDuration(grantDuration)
	if err != nil {
		return err
	}

	client, _, err := cliutil.Client(cmd)
	if err != nil {
		return err
	}

	output.Info("plan: grant %q to %q for %s%s", entitlementID, customerID, label, reasonSuffix(grantReason))
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
	ent, err := client.GrantPromotionalEntitlement(ctx, customerID, entitlementID, api.GrantPromotionalRequest{
		Duration: rcDuration,
		Reason:   grantReason,
	})
	if err != nil {
		return err
	}
	output.Success("granted; expires %s", expiryString(ent))
	return nil
}

// normalizeDuration translates a user-friendly duration ("7d", "1y",
// "forever") into the literal string RC's API expects, plus a human
// label for the plan line.
func normalizeDuration(s string) (rc string, label string, err error) {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "forever", "lifetime":
		return "lifetime", "lifetime", nil
	case "daily", "weekly", "monthly", "yearly", "two_months", "three_months", "six_months":
		return s, s, nil
	}
	if len(s) < 2 {
		return "", "", fmt.Errorf("invalid duration %q", s)
	}
	unit := s[len(s)-1]
	num := s[:len(s)-1]
	switch unit {
	case 'd':
		// Map to RC's day-count promotional grants. RC accepts P{n}D
		// ISO-8601 form, which we pass through verbatim.
		return "P" + num + "D", num + " day(s)", nil
	case 'm':
		return "P" + num + "M", num + " month(s)", nil
	case 'y':
		return "P" + num + "Y", num + " year(s)", nil
	}
	return "", "", fmt.Errorf("invalid duration %q (try 7d, 1m, 1y, or lifetime)", s)
}

func expiryString(e *api.ActiveEntitlement) string {
	if e == nil || e.ExpiresAt == 0 {
		return "never"
	}
	t := time.UnixMilli(e.ExpiresAt).Local()
	if e.ExpiresAt < 9999999999 {
		t = time.Unix(e.ExpiresAt, 0).Local()
	}
	return t.Format("2006-01-02 15:04 MST")
}

func reasonSuffix(reason string) string {
	if reason == "" {
		return ""
	}
	return ` (reason: "` + reason + `")`
}
