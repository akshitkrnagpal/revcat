// Package offerings holds `revcat offerings ...`.
package offerings

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
	Use:     "offerings",
	Aliases: []string{"offer"},
	Short:   "Manage RevenueCat offerings",
	Long: `An offering is a presentation grouping of packages displayed on a
paywall. Each project has 0..N offerings; exactly one is "current" and is
returned by SDKs that ask for the current offering.

To set an offering current along with a paywall config in one shot, use
` + "`revcat publish offering`" + `.`,
}

func init() {
	Cmd.AddCommand(listCmd, viewCmd, createCmd, updateCmd, deleteCmd, archiveCmd, unarchiveCmd, setCurrentCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all offerings in the active project",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		offers, err := client.ListOfferings(ctx)
		if err != nil {
			return err
		}
		rows := make([][]any, 0, len(offers))
		for _, o := range offers {
			marker := ""
			if o.IsCurrent {
				marker = "*"
			}
			rows = append(rows, []any{marker, o.LookupKey, o.DisplayName, cliutil.FormatTime(o.CreatedAt)})
		}
		return output.Table([]string{"", "id", "display_name", "created"}, rows)
	},
}

var viewWithPackages bool

var viewCmd = &cobra.Command{
	Use:   "view <id>",
	Short: "Show one offering by lookup_key",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		o, err := client.GetOffering(ctx, args[0], viewWithPackages || !output.IsJSON())
		if err != nil {
			return err
		}
		if output.IsJSON() {
			return output.JSON(o)
		}
		current := "no"
		if o.IsCurrent {
			current = "yes (current)"
		}
		head := [][]any{
			{"id", o.LookupKey},
			{"display_name", o.DisplayName},
			{"is_current", current},
			{"created", cliutil.FormatTime(o.CreatedAt)},
			{"package count", len(o.Packages)},
		}
		if err := output.Table([]string{"field", "value"}, head); err != nil {
			return err
		}
		if len(o.Packages) > 0 {
			pkgRows := make([][]any, 0, len(o.Packages))
			for _, p := range o.Packages {
				pkgRows = append(pkgRows, []any{p.Position, p.Identifier, cliutil.Dash(p.ProductID)})
			}
			return output.Table([]string{"#", "identifier", "product"}, pkgRows)
		}
		return nil
	},
}

func init() {
	viewCmd.Flags().BoolVar(&viewWithPackages, "packages", true, "Include packages in the rendered card")
}

var (
	createFile string
	createKey  string
	createName string
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an offering",
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
		o, err := client.CreateOffering(ctx, body)
		if err != nil {
			return err
		}
		output.Success("created %s", o.LookupKey)
		if output.IsJSON() {
			return output.JSON(o)
		}
		return nil
	},
}

func init() {
	createCmd.Flags().StringVarP(&createFile, "file", "f", "", "Body as JSON file")
	createCmd.Flags().StringVar(&createKey, "id", "", "Offering lookup_key")
	createCmd.Flags().StringVar(&createName, "display-name", "", "Display name")
}

var (
	updateFile string
	updateName string
)

var updateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update an offering",
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
		o, err := client.UpdateOffering(ctx, args[0], body)
		if err != nil {
			return err
		}
		output.Success("updated %s", o.LookupKey)
		if output.IsJSON() {
			return output.JSON(o)
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
	Short: "Delete an offering",
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
		if err := client.DeleteOffering(ctx, args[0]); err != nil {
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
	Short: "Archive an offering",
	Args:  cobra.ExactArgs(1),
	RunE:  archiveAction(true),
}

var unarchiveCmd = &cobra.Command{
	Use:   "unarchive <id>",
	Short: "Unarchive an offering",
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
		if err := client.ArchiveOffering(ctx, args[0], archive); err != nil {
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

var setCurrentCmd = &cobra.Command{
	Use:   "set-current <id>",
	Short: "Promote an offering to current",
	Long: `Promote an offering so SDKs returning "current offering" hand back this
one. Equivalent to ` + "`revcat publish offering <id> --current`" + ` minus the
plan/confirm flow.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		o, err := client.SetCurrentOffering(ctx, args[0])
		if err != nil {
			return err
		}
		output.Success("offering %s is now current", o.LookupKey)
		return nil
	},
}
