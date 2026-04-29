package offerings

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/akshitkrnagpal/revcat/internal/api"
	"github.com/akshitkrnagpal/revcat/internal/cliutil"
	"github.com/akshitkrnagpal/revcat/internal/output"
)

var (
	previewAs       string
	previewPlatform string
	previewAppID    string
)

var previewCmd = &cobra.Command{
	Use:   "preview [<offering_id>]",
	Short: "Show what the SDK will receive from /v1/subscribers/{user}/offerings",
	Long: `Hit the v1 SDK-facing endpoint that ` + "`Purchases.getOfferings()`" + ` calls and
render the response. Useful when the dashboard looks healthy but the SDK
reports zero packages.

Auth is auto-handled: revcat fetches the public SDK key for the project's
app on the chosen platform and sends it as the bearer. ` + "`--as`" + ` defaults to a
synthetic user id (` + "`revcat_preview_<unix_ms>`" + `) so you don't have to think
about it.

Pass an optional <offering_id> to filter the rendered output to a single
offering (the request still fetches all offerings; v1 has no per-offering
endpoint).`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()

		platform := strings.ToLower(strings.TrimSpace(previewPlatform))
		if platform == "" {
			platform = "ios"
		}
		if !validPlatform(platform) {
			return fmt.Errorf("unsupported --platform %q (want ios|android|web)", platform)
		}

		appID := previewAppID
		if appID == "" {
			id, err := pickAppForPlatform(ctx, client, platform)
			if err != nil {
				return err
			}
			appID = id
		}

		key, err := pickPublicKey(ctx, client, appID)
		if err != nil {
			return err
		}

		userID := previewAs
		if userID == "" {
			userID = fmt.Sprintf("revcat_preview_%d", time.Now().UnixMilli())
		}

		resp, err := client.PreviewOfferings(ctx, userID, key, headerPlatform(platform))
		if err != nil {
			return err
		}

		if len(args) == 1 && args[0] != "" {
			filtered := make([]api.PreviewOffering, 0, 1)
			for _, o := range resp.Offerings {
				if o.Identifier == args[0] {
					filtered = append(filtered, o)
				}
			}
			resp.Offerings = filtered
		}

		if output.IsJSON() {
			return output.JSON(resp)
		}

		output.Info("user: %s   platform: %s   app: %s", userID, platform, appID)
		output.Info("current_offering_id: %s", cliutil.Dash(resp.CurrentOfferingID))

		if len(resp.Offerings) == 0 {
			output.Warn("no offerings returned")
			return nil
		}

		header := [][]any{}
		for _, o := range resp.Offerings {
			cur := ""
			if o.Identifier == resp.CurrentOfferingID {
				cur = "*"
			}
			header = append(header, []any{cur, o.Identifier, len(o.Packages)})
		}
		if err := output.Table([]string{"", "offering", "packages"}, header); err != nil {
			return err
		}

		for _, o := range resp.Offerings {
			if len(o.Packages) == 0 {
				continue
			}
			fmt.Fprintln(cmd.OutOrStdout())
			output.Info("packages in %q:", o.Identifier)
			rows := make([][]any, 0, len(o.Packages))
			for _, p := range o.Packages {
				rows = append(rows, []any{p.Identifier, cliutil.Dash(p.PlatformProductIdentifier)})
			}
			if err := output.Table([]string{"identifier", "platform_product_identifier"}, rows); err != nil {
				return err
			}
		}
		return nil
	},
}

func init() {
	previewCmd.Flags().StringVar(&previewAs, "as", "", "User id to query as (default: revcat_preview_<unix_ms>)")
	previewCmd.Flags().StringVar(&previewPlatform, "platform", "ios", "Storefront platform to preview (ios|android|web)")
	previewCmd.Flags().StringVar(&previewAppID, "app-id", "", "App id to derive the public SDK key from (default: auto-detect)")
	Cmd.AddCommand(previewCmd)
}

func validPlatform(p string) bool {
	switch p {
	case "ios", "android", "web":
		return true
	}
	return false
}

// headerPlatform returns the value to send in the X-Platform header. The v1
// API accepts the SDK platform identifiers; we mirror what the official
// SDKs send.
func headerPlatform(p string) string {
	switch p {
	case "ios":
		return "iOS"
	case "android":
		return "android"
	case "web":
		return "web"
	}
	return p
}

// appTypesFor maps a CLI --platform value to the set of v2 App.Type values
// that count as that storefront. iOS covers App Store + Mac App Store; web
// covers Stripe + RC web billing.
func appTypesFor(platform string) []string {
	switch platform {
	case "ios":
		return []string{"app_store", "mac_app_store"}
	case "android":
		return []string{"play_store", "amazon"}
	case "web":
		return []string{"stripe", "rc_billing", "web_billing"}
	}
	return nil
}

// pickAppForPlatform finds the single app in the active project whose type
// matches the requested platform. If 0 or >1 apps match, we surface the
// list and ask the user to pass --app-id explicitly.
func pickAppForPlatform(ctx context.Context, client *api.Client, platform string) (string, error) {
	apps, err := client.ListApps(ctx)
	if err != nil {
		return "", err
	}
	wanted := appTypesFor(platform)
	matches := make([]api.App, 0, 2)
	for _, a := range apps {
		for _, t := range wanted {
			if a.Type == t {
				matches = append(matches, a)
				break
			}
		}
	}
	if len(matches) == 1 {
		return matches[0].ID, nil
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("no app in this project matches --platform %s; pass --app-id explicitly (run `revcat apps list`)", platform)
	}
	names := make([]string, 0, len(matches))
	for _, m := range matches {
		names = append(names, fmt.Sprintf("%s (%s, %s)", m.ID, m.Name, m.Type))
	}
	return "", fmt.Errorf("multiple apps match --platform %s, pass --app-id explicitly:\n  - %s", platform, strings.Join(names, "\n  - "))
}

// pickPublicKey returns the first public SDK key for the given app. RC
// projects usually expose one production key per app; if there are
// multiple, we take the first listed (callers can pass --app-id to scope).
func pickPublicKey(ctx context.Context, client *api.Client, appID string) (string, error) {
	keys, err := client.ListPublicAPIKeys(ctx, appID)
	if err != nil {
		return "", fmt.Errorf("fetch public keys for app %s: %w", appID, err)
	}
	for _, k := range keys {
		if k.Key != "" {
			return k.Key, nil
		}
	}
	return "", fmt.Errorf("app %s has no public SDK keys; create one in the RevenueCat dashboard", appID)
}
