// Package apps holds `revcat apps ...`.
package apps

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

var Cmd = &cobra.Command{
	Use:   "apps",
	Short: "Manage RevenueCat apps (per-platform inside a project)",
	Long: `Each project has one app per platform/storefront (one for iOS, one for
Android, etc.).

Read commands ` + "`list`, `view`, `public-keys`, `storekit-config`" + ` work on
any app. Write commands ` + "`create`, `update`, `delete`" + ` use the v2 app
endpoints; pass ` + "`--file <path>`" + ` for any non-trivial body since the
schema is wide and storefront-specific.`,
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(viewCmd)
	Cmd.AddCommand(keysCmd)
	Cmd.AddCommand(storeKitCmd)
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(updateCmd)
	Cmd.AddCommand(deleteCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List apps in the active project",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		apps, err := client.ListApps(ctx)
		if err != nil {
			return err
		}
		rows := make([][]any, 0, len(apps))
		for _, a := range apps {
			id := a.BundleID
			if id == "" {
				id = a.PackageName
			}
			rows = append(rows, []any{a.ID, a.Name, a.Type, dash(id), formatTime(a.CreatedAt)})
		}
		return output.Table([]string{"id", "name", "platform", "bundle/package", "created"}, rows)
	},
}

var viewCmd = &cobra.Command{
	Use:   "view <id>",
	Short: "Show one app",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		a, raw, err := client.GetAppRaw(ctx, args[0])
		if err != nil {
			return err
		}
		if output.IsJSON() {
			return output.JSON(raw)
		}
		bundleOrPkg := a.BundleID
		bundleLabel := "bundle_id"
		if bundleOrPkg == "" {
			bundleOrPkg = a.PackageName
			bundleLabel = "package_name"
		}
		rows := [][]any{
			{"id", a.ID},
			{"name", a.Name},
			{"platform", a.Type},
			{bundleLabel, dash(bundleOrPkg)},
			{"created", formatTime(a.CreatedAt)},
		}
		return output.Table([]string{"field", "value"}, rows)
	},
}

var keysCmd = &cobra.Command{
	Use:   "public-keys <app_id>",
	Short: "List the public SDK keys for an app",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		keys, err := client.ListPublicAPIKeys(ctx, args[0])
		if err != nil {
			return err
		}
		rows := make([][]any, 0, len(keys))
		for _, k := range keys {
			rows = append(rows, []any{k.ID, k.Label, k.Key})
		}
		return output.Table([]string{"id", "label", "key"}, rows)
	},
}

var storeKitCmd = &cobra.Command{
	Use:   "storekit-config <app_id>",
	Short: "Print the StoreKit configuration for an iOS app",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		cfg, err := client.GetStoreKitConfig(ctx, args[0])
		if err != nil {
			return err
		}
		return output.JSON(cfg)
	},
}

var (
	createName    string
	createType    string
	createBundle  string
	createPackage string
	createFile    string

	updateName string
	updateFile string

	deleteYes bool
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an app under the active project",
	Long: `Create a new app. v2's app body is discriminated by --type and the
shape varies per storefront. For the common cases revcat takes
shortcut flags:

  --type app_store    --bundle <com.acme.app>
  --type play_store   --package <com.acme.app>
  --type amazon       --package <com.acme.app>

Anything more (Stripe, rc_billing, paddle, roku, mac_app_store, or any
optional fields like a shared_secret) needs --file <path.json>.

Examples:
    revcat apps create --type app_store --bundle com.acme.app --name "Acme iOS"
    revcat apps create --file ./apps/new-stripe.json
    revcat apps create --file - <<< '{"name":"Web","type":"stripe","stripe":{"stripe_account_id":"acct_..."}}'`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var body map[string]any
		switch {
		case createFile != "":
			b, err := cliutil.LoadJSON(createFile)
			if err != nil {
				return err
			}
			body = b
		case createType != "":
			name := strings.TrimSpace(createName)
			if name == "" {
				return errors.New("--name is required when using shortcut flags (or pass --file)")
			}
			body = map[string]any{"name": name, "type": createType}
			switch createType {
			case "app_store", "mac_app_store":
				if createBundle == "" {
					return fmt.Errorf("--bundle is required for type %q", createType)
				}
				body[createType] = map[string]any{"bundle_id": createBundle}
			case "play_store", "amazon":
				if createPackage == "" {
					return fmt.Errorf("--package is required for type %q", createType)
				}
				body[createType] = map[string]any{"package_name": createPackage}
			default:
				return fmt.Errorf("--type %q has no shortcut; pass --file with the full body", createType)
			}
		default:
			return errors.New("pass --type with --bundle/--package, or --file with the full v2 body")
		}

		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		a, err := client.CreateApp(ctx, body)
		if err != nil {
			return err
		}
		if output.IsJSON() {
			return output.JSON(a)
		}
		return printApp(a)
	},
}

var updateCmd = &cobra.Command{
	Use:   "update <app_id>",
	Short: "Update an app",
	Long: `Update an existing app. Pass --name to rename, or --file <path.json>
for a full v2 body (everything except 'type'). Send a nested field as
null in the JSON body to clear it (e.g. {"app_store":{"shared_secret":null}}).

Examples:
    revcat apps update app_abc --name "Acme iOS (renamed)"
    revcat apps update app_abc --file ./patch.json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var body map[string]any
		if updateFile != "" {
			b, err := cliutil.LoadJSON(updateFile)
			if err != nil {
				return err
			}
			body = b
		} else if updateName != "" {
			body = map[string]any{"name": updateName}
		} else {
			return errors.New("pass --name or --file")
		}

		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		a, err := client.UpdateApp(ctx, args[0], body)
		if err != nil {
			return err
		}
		if output.IsJSON() {
			return output.JSON(a)
		}
		return printApp(a)
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete <app_id>",
	Short: "Delete an app (hard delete)",
	Long: `Delete an app. This is a hard delete and can return 409 if the app
has dependent resources (offerings, products, etc.). Prefer to drain
those first.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !deleteYes {
			var ok bool
			if err := survey.AskOne(&survey.Confirm{
				Message: fmt.Sprintf("Permanently delete app %q? This cannot be undone.", args[0]),
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
		if err := client.DeleteApp(ctx, args[0]); err != nil {
			return err
		}
		output.Success("deleted app %q", args[0])
		return nil
	},
}

func init() {
	createCmd.Flags().StringVar(&createName, "name", "", "App name")
	createCmd.Flags().StringVar(&createType, "type", "", "Storefront type: app_store | play_store | amazon | mac_app_store (use --file for stripe / rc_billing / paddle / roku)")
	createCmd.Flags().StringVar(&createBundle, "bundle", "", "Bundle id (app_store / mac_app_store)")
	createCmd.Flags().StringVar(&createPackage, "package", "", "Package name (play_store / amazon)")
	createCmd.Flags().StringVar(&createFile, "file", "", "Path to a JSON file with the full v2 body (use - for stdin)")

	updateCmd.Flags().StringVar(&updateName, "name", "", "Rename the app")
	updateCmd.Flags().StringVar(&updateFile, "file", "", "Path to a JSON file with the v2 update body (use - for stdin)")

	deleteCmd.Flags().BoolVarP(&deleteYes, "confirm", "y", false, "Skip the confirmation prompt")
}

func printApp(a *api.App) error {
	bundleOrPkg := a.BundleID
	bundleLabel := "bundle_id"
	if bundleOrPkg == "" {
		bundleOrPkg = a.PackageName
		bundleLabel = "package_name"
	}
	rows := [][]any{
		{"id", a.ID},
		{"name", a.Name},
		{"platform", a.Type},
		{bundleLabel, dash(bundleOrPkg)},
		{"created", formatTime(a.CreatedAt)},
	}
	return output.Table([]string{"field", "value"}, rows)
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
