// Package entitlements holds `revcat entitlements ...`.
package entitlements

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
	Use:     "entitlements",
	Aliases: []string{"ent"},
	Short:   "Manage RevenueCat entitlements",
	Long: `Entitlements are project-level access flags (e.g., "premium", "pro").
Customers gain entitlements via products attached on offerings, or via
promotional grants. Use ` + "`revcat subscribers info <user_id>`" + ` to see what
a specific customer has.`,
}

func init() {
	Cmd.AddCommand(listCmd, viewCmd, createCmd, updateCmd, deleteCmd, archiveCmd, unarchiveCmd, productsCmd, attachCmd, detachCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all entitlements in the active project",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		ents, err := client.ListEntitlements(ctx)
		if err != nil {
			return err
		}
		rows := make([][]any, 0, len(ents))
		for _, e := range ents {
			rows = append(rows, []any{e.LookupKey, e.DisplayName, cliutil.FormatTime(e.CreatedAt)})
		}
		return output.Table([]string{"id", "display_name", "created"}, rows)
	},
}

var viewCmd = &cobra.Command{
	Use:   "view <id>",
	Short: "Show one entitlement by lookup_key",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		e, raw, err := client.GetEntitlementRaw(ctx, args[0])
		if err != nil {
			return err
		}
		if output.IsJSON() {
			return output.JSON(raw)
		}
		rows := [][]any{
			{"id", e.LookupKey},
			{"display_name", e.DisplayName},
			{"created", cliutil.FormatTime(e.CreatedAt)},
			{"internal_id", e.ID},
		}
		return output.Table([]string{"field", "value"}, rows)
	},
}

var (
	createFile string
	createKey  string
	createName string
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an entitlement",
	Long: `Create an entitlement. For the common case use shortcut flags:

    revcat entitlements create --id premium --display-name "Premium"

For arbitrary v2 fields, pass --file <path.json>.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var body map[string]any
		if createFile != "" {
			b, err := cliutil.LoadJSON(createFile)
			if err != nil {
				return err
			}
			body = b
		} else if createKey != "" {
			body = map[string]any{"lookup_key": createKey}
			if createName != "" {
				body["display_name"] = createName
			}
		} else {
			return errors.New("pass --file or --id")
		}
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		e, err := client.CreateEntitlement(ctx, body)
		if err != nil {
			return err
		}
		output.Success("created %s", e.LookupKey)
		if output.IsJSON() {
			return output.JSON(e)
		}
		return nil
	},
}

func init() {
	createCmd.Flags().StringVarP(&createFile, "file", "f", "", "Body as JSON file (or '-' for stdin)")
	createCmd.Flags().StringVar(&createKey, "id", "", "Entitlement lookup_key (e.g., premium)")
	createCmd.Flags().StringVar(&createName, "display-name", "", "Display name shown in dashboards")
}

var (
	updateFile string
	updateName string
)

var updateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update an entitlement",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var body map[string]any
		if updateFile != "" {
			b, err := cliutil.LoadJSON(updateFile)
			if err != nil {
				return err
			}
			body = b
		} else if updateName != "" {
			body = map[string]any{"display_name": updateName}
		} else {
			return errors.New("pass --file or --display-name")
		}
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		e, err := client.UpdateEntitlement(ctx, args[0], body)
		if err != nil {
			return err
		}
		output.Success("updated %s", e.LookupKey)
		if output.IsJSON() {
			return output.JSON(e)
		}
		return nil
	},
}

func init() {
	updateCmd.Flags().StringVarP(&updateFile, "file", "f", "", "Patch body as JSON")
	updateCmd.Flags().StringVar(&updateName, "display-name", "", "New display name")
}

var deleteConfirm bool

var deleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete an entitlement",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !deleteConfirm {
			var ok bool
			if err := survey.AskOne(&survey.Confirm{Message: "delete " + args[0] + "?", Default: false}, &ok); err != nil {
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
		if err := client.DeleteEntitlement(ctx, args[0]); err != nil {
			return err
		}
		output.Success("deleted %s", args[0])
		return nil
	},
}

func init() {
	deleteCmd.Flags().BoolVarP(&deleteConfirm, "confirm", "y", false, "Skip the prompt")
}

var archiveCmd = &cobra.Command{
	Use:   "archive <id>",
	Short: "Archive an entitlement",
	Args:  cobra.ExactArgs(1),
	RunE:  archiveAction(true),
}

var unarchiveCmd = &cobra.Command{
	Use:   "unarchive <id>",
	Short: "Unarchive an entitlement",
	Args:  cobra.ExactArgs(1),
	RunE:  archiveAction(false),
}

func archiveAction(archive bool) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		if err := client.ArchiveEntitlement(ctx, args[0], archive); err != nil {
			return err
		}
		verb := "archived"
		if !archive {
			verb = "unarchived"
		}
		output.Success("%s %s", verb, args[0])
		return nil
	}
}

var productsCmd = &cobra.Command{
	Use:   "products <id>",
	Short: "List products attached to an entitlement",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		prods, err := client.ListEntitlementProducts(ctx, args[0])
		if err != nil {
			return err
		}
		rows := make([][]any, 0, len(prods))
		for _, p := range prods {
			rows = append(rows, []any{p.ID, p.StoreIdentifier, p.Type, p.DisplayName})
		}
		return output.Table([]string{"id", "store_id", "type", "display_name"}, rows)
	},
}

var attachCmd = &cobra.Command{
	Use:   "attach <id> <product_id> [<product_id> ...]",
	Short: "Attach product(s) to an entitlement",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		if err := client.AttachProductsToEntitlement(ctx, args[0], args[1:]); err != nil {
			return err
		}
		output.Success("attached %d product(s) to %s", len(args)-1, args[0])
		return nil
	},
}

var detachCmd = &cobra.Command{
	Use:   "detach <id> <product_id> [<product_id> ...]",
	Short: "Detach product(s) from an entitlement",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		if err := client.DetachProductsFromEntitlement(ctx, args[0], args[1:]); err != nil {
			return err
		}
		output.Success("detached %d product(s) from %s", len(args)-1, args[0])
		return nil
	},
}

