// Package packages holds `revcat packages ...`.
//
// Packages live within offerings; this list flattens them by fetching every
// offering's packages, optionally narrowed with --offering.
package packages

import (
	"context"
	"errors"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"

	"github.com/akshitkrnagpal/revcat/internal/api"
	"github.com/akshitkrnagpal/revcat/internal/cliutil"
	"github.com/akshitkrnagpal/revcat/internal/output"
)

var Cmd = &cobra.Command{
	Use:     "packages",
	Aliases: []string{"pkg"},
	Short:   "Manage RevenueCat packages (purchasables inside an offering)",
	Long: `Packages are the purchasable units inside an offering. Identifiers
follow RC's $rc_monthly / $rc_annual / custom convention.`,
}

func init() {
	Cmd.AddCommand(listCmd, viewCmd, createCmd, updateCmd, deleteCmd, productsCmd, attachCmd, detachCmd)
}

var listOffering string

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List packages across one offering or all offerings",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()

		offerings := []string{}
		if listOffering != "" {
			offerings = append(offerings, listOffering)
		} else {
			all, err := client.ListOfferings(ctx)
			if err != nil {
				return err
			}
			for _, o := range all {
				offerings = append(offerings, o.LookupKey)
			}
		}

		var rows [][]any
		var pkgs []api.Package
		for _, off := range offerings {
			ps, err := client.ListPackages(ctx, off)
			if err != nil {
				output.Warn("offering %q: %v", off, err)
				continue
			}
			for _, p := range ps {
				pkgs = append(pkgs, p)
				rows = append(rows, []any{off, p.Position, p.Identifier, cliutil.Dash(p.ProductID), p.ID})
			}
		}

		if output.IsJSON() {
			return output.JSON(pkgs)
		}
		return output.Table([]string{"offering", "#", "identifier", "product", "internal_id"}, rows)
	},
}

func init() {
	listCmd.Flags().StringVarP(&listOffering, "offering", "o", "", "Restrict to a single offering by lookup_key")
}

var viewCmd = &cobra.Command{
	Use:   "view <id>",
	Short: "Show one package by internal id",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		p, err := client.GetPackage(ctx, args[0])
		if err != nil {
			return err
		}
		if output.IsJSON() {
			return output.JSON(p)
		}
		rows := [][]any{
			{"id", p.ID},
			{"identifier", p.Identifier},
			{"position", p.Position},
			{"offering_id", cliutil.Dash(p.OfferingID)},
			{"product_id", cliutil.Dash(p.ProductID)},
			{"created", cliutil.FormatTime(p.CreatedAt)},
		}
		return output.Table([]string{"field", "value"}, rows)
	},
}

var createFile string

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a package from a JSON body",
	Long: `Create a package. Required fields per v2: identifier, offering_id.
Pass them inside --file <path>.

Example body:
    {
      "identifier": "$rc_monthly",
      "offering_id": "ofr_xxx",
      "position": 1
    }`,
	RunE: func(cmd *cobra.Command, args []string) error {
		body, err := cliutil.LoadJSON(createFile)
		if err != nil {
			return err
		}
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		p, err := client.CreatePackage(ctx, body)
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
	createCmd.Flags().StringVarP(&createFile, "file", "f", "", "Body as JSON file")
	_ = createCmd.MarkFlagRequired("file")
}

var updateFile string

var updateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a package",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		body, err := cliutil.LoadJSON(updateFile)
		if err != nil {
			return err
		}
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		p, err := client.UpdatePackage(ctx, args[0], body)
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
	_ = updateCmd.MarkFlagRequired("file")
}

var deleteConfirm bool

var deleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a package",
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
		if err := client.DeletePackage(ctx, args[0]); err != nil {
			return err
		}
		output.Success("deleted %s", args[0])
		return nil
	},
}

func init() {
	deleteCmd.Flags().BoolVarP(&deleteConfirm, "confirm", "y", false, "Skip the prompt")
}

var productsCmd = &cobra.Command{
	Use:   "products <id>",
	Short: "List products attached to a package",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		prods, err := client.ListPackageProducts(ctx, args[0])
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
	Short: "Attach product(s) to a package",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		if err := client.AttachProductsToPackage(ctx, args[0], args[1:]); err != nil {
			return err
		}
		output.Success("attached %d product(s) to %s", len(args)-1, args[0])
		return nil
	},
}

var detachCmd = &cobra.Command{
	Use:   "detach <id> <product_id> [<product_id> ...]",
	Short: "Detach product(s) from a package",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		if err := client.DetachProductsFromPackage(ctx, args[0], args[1:]); err != nil {
			return err
		}
		output.Success("detached %d product(s) from %s", len(args)-1, args[0])
		return nil
	},
}
