// Package paywalls holds `revcat paywalls ...` (top-level paywall resource).
//
// Distinct from the offering-scoped paywall config that
// `revcat publish offering --paywall <file>` PUTs.
package paywalls

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
	Use:   "paywalls",
	Short: "Manage top-level paywall resources",
	Long: `Manage paywall records in the project's paywall library. To deploy a
paywall config to an offering use ` + "`revcat publish offering --paywall <file>`" + `.`,
}

func init() {
	Cmd.AddCommand(listCmd, viewCmd, createCmd, deleteCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List paywalls in the project",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		pws, err := client.ListPaywalls(ctx)
		if err != nil {
			return err
		}
		rows := make([][]any, 0, len(pws))
		for _, p := range pws {
			rows = append(rows, []any{p.ID, p.Name, cliutil.Dash(p.Template), cliutil.FormatTime(p.CreatedAt)})
		}
		return output.Table([]string{"id", "name", "template", "created"}, rows)
	},
}

var viewCmd = &cobra.Command{
	Use:   "view <id>",
	Short: "Show one paywall (raw JSON)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		p, err := client.GetPaywallByID(ctx, args[0])
		if err != nil {
			return err
		}
		return output.JSON(p)
	},
}

var (
	createFile     string
	createOffering string
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a paywall record scoped to an offering",
	Long: `Create a paywall record. v2 only takes {offering_id} - the paywall
content (template, copy, components) is set later via
` + "`revcat publish offering <id> --paywall <file>`" + `.

You usually want publish offering instead of this command.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var body map[string]any
		if createFile != "" {
			b, err := cliutil.LoadJSON(createFile)
			if err != nil {
				return err
			}
			body = b
		} else if createOffering != "" {
			body = map[string]any{"offering_id": createOffering}
		} else {
			return errors.New("pass --file or --offering")
		}
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		out, err := client.CreatePaywall(ctx, body)
		if err != nil {
			return err
		}
		output.Success("created paywall")
		if output.IsJSON() {
			return output.JSON(out)
		}
		return nil
	},
}

func init() {
	createCmd.Flags().StringVarP(&createFile, "file", "f", "", "JSON body")
	createCmd.Flags().StringVar(&createOffering, "offering", "", "Offering id (required if --file not used)")
}

var deleteConfirm bool

var deleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a paywall",
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
		if err := client.DeletePaywall(ctx, args[0]); err != nil {
			return err
		}
		output.Success("deleted %s", args[0])
		return nil
	},
}

func init() {
	deleteCmd.Flags().BoolVarP(&deleteConfirm, "confirm", "y", false, "Skip the prompt")
}
