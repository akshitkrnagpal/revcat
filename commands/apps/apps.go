// Package apps holds `revcat apps ...`.
package apps

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/akshitkrnagpal/revcat/internal/cliutil"
	"github.com/akshitkrnagpal/revcat/internal/output"
)

var Cmd = &cobra.Command{
	Use:   "apps",
	Short: "Inspect RevenueCat apps (per-platform inside a project)",
	Long: `Each project has one app per platform/storefront (one for iOS, one for
Android, etc.). App create/update/delete needs a partner-tier key and is
not exposed here; this group is read-only for normal secret keys.`,
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(viewCmd)
	Cmd.AddCommand(keysCmd)
	Cmd.AddCommand(storeKitCmd)
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
