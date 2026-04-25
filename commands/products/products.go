// Package products holds `revcat products ...`.
package products

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"

	"github.com/akshitkrnagpal/revcat/internal/cliutil"
	"github.com/akshitkrnagpal/revcat/internal/output"
)

var Cmd = &cobra.Command{
	Use:     "products",
	Aliases: []string{"prod"},
	Short:   "Manage RevenueCat products (store SKUs)",
	Long: `A product is a project-level catalog entry that mirrors a store SKU
(App Store / Play Store / Stripe / Web Billing). Products are attached
to packages, packages live inside offerings.

Most edits accept a JSON file via --file, since the product schema
differs per store and revcat does not pin a specific shape. Use the
v2 docs to author the body, then ` + "`revcat products create -f product.json`" + `.`,
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(viewCmd)
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(updateCmd)
	Cmd.AddCommand(deleteCmd)
	Cmd.AddCommand(archiveCmd)
	Cmd.AddCommand(unarchiveCmd)
	Cmd.AddCommand(pushCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List products",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		products, err := client.ListProducts(ctx)
		if err != nil {
			return err
		}
		rows := make([][]any, 0, len(products))
		for _, p := range products {
			rows = append(rows, []any{p.ID, p.StoreIdentifier, p.Type, p.DisplayName, formatTime(p.CreatedAt)})
		}
		return output.Table([]string{"id", "store_id", "type", "display_name", "created"}, rows)
	},
}

var viewCmd = &cobra.Command{
	Use:   "view <id>",
	Short: "Show one product",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		p, err := client.GetProduct(ctx, args[0])
		if err != nil {
			return err
		}
		if output.IsJSON() {
			return output.JSON(p)
		}
		rows := [][]any{
			{"id", p.ID},
			{"store_identifier", p.StoreIdentifier},
			{"type", p.Type},
			{"display_name", p.DisplayName},
			{"app_id", p.AppID},
			{"created", formatTime(p.CreatedAt)},
		}
		return output.Table([]string{"field", "value"}, rows)
	},
}

var createFile string

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a product from a JSON body",
	Long: `Create a product from a JSON body on disk. The body shape follows the
v2 docs - store_identifier, type (subscription | non_subscription | ...),
display_name, app_id are usually required.

Example body:
    {
      "store_identifier": "app.monthly",
      "type": "subscription",
      "display_name": "Monthly",
      "app_id": "app_xxx"
    }`,
	RunE: func(cmd *cobra.Command, args []string) error {
		body, err := loadJSON(createFile)
		if err != nil {
			return err
		}
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		p, err := client.CreateProduct(ctx, body)
		if err != nil {
			return err
		}
		output.Success("created %s", p.ID)
		if output.IsJSON() {
			return output.JSON(p)
		}
		return nil
	},
}

func init() {
	createCmd.Flags().StringVarP(&createFile, "file", "f", "", "Path to JSON body (required)")
	_ = createCmd.MarkFlagRequired("file")
}

var updateFile string
var updateName string

var updateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a product",
	Long: `Update a product. Pass --file <path.json> for an arbitrary patch, or
--display-name <new> for the common case.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var body map[string]any
		if updateFile != "" {
			b, err := loadJSON(updateFile)
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
		p, err := client.UpdateProduct(ctx, args[0], body)
		if err != nil {
			return err
		}
		output.Success("updated %s", p.ID)
		if output.IsJSON() {
			return output.JSON(p)
		}
		return nil
	},
}

func init() {
	updateCmd.Flags().StringVarP(&updateFile, "file", "f", "", "Patch body as JSON")
	updateCmd.Flags().StringVar(&updateName, "display-name", "", "New display name (shortcut for the common case)")
}

var deleteConfirm bool

var deleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a product (most teams should archive instead)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !deleteConfirm {
			var ok bool
			if err := survey.AskOne(&survey.Confirm{
				Message: fmt.Sprintf("delete product %q? (irreversible; archive is usually better)", args[0]),
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
		if err := client.DeleteProduct(ctx, args[0]); err != nil {
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
	Short: "Archive a product",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		if err := client.ArchiveProduct(ctx, args[0], true); err != nil {
			return err
		}
		output.Success("archived %s", args[0])
		return nil
	},
}

var unarchiveCmd = &cobra.Command{
	Use:   "unarchive <id>",
	Short: "Unarchive a product",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		if err := client.ArchiveProduct(ctx, args[0], false); err != nil {
			return err
		}
		output.Success("unarchived %s", args[0])
		return nil
	},
}

var pushCmd = &cobra.Command{
	Use:   "push-to-store <id>",
	Short: "Push a product config to the linked store",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		resp, err := client.PushProductToStore(ctx, args[0])
		if err != nil {
			return err
		}
		output.Success("push initiated")
		if output.IsJSON() {
			return output.JSON(resp)
		}
		return nil
	},
}

var loadJSON = cliutil.LoadJSON
var formatTime = cliutil.FormatTime
