// Package virtualcurrencies holds `revcat virtual-currencies ...`.
package virtualcurrencies

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
	Use:     "virtual-currencies",
	Aliases: []string{"vc"},
	Short:   "Manage virtual currencies (coins / credits)",
	Long: `Project-level virtual currencies (in-game coins, credits, tokens). For
per-customer balances and transactions see the related ` + "`revcat subscribers vc-balance`" + ` /
` + "`revcat subscribers vc-tx`" + ` commands.`,
}

func init() {
	Cmd.AddCommand(listCmd, viewCmd, createCmd, updateCmd, deleteCmd, archiveCmd, unarchiveCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List virtual currencies",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		vcs, err := client.ListVirtualCurrencies(ctx)
		if err != nil {
			return err
		}
		rows := make([][]any, 0, len(vcs))
		for _, v := range vcs {
			rows = append(rows, []any{v.LookupKey, v.DisplayName, cliutil.Dash(v.Code), cliutil.FormatTime(v.CreatedAt)})
		}
		return output.Table([]string{"id", "display_name", "code", "created"}, rows)
	},
}

var viewCmd = &cobra.Command{
	Use:   "view <id>",
	Short: "Show one virtual currency",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		v, err := client.GetVirtualCurrency(ctx, args[0])
		if err != nil {
			return err
		}
		if output.IsJSON() {
			return output.JSON(v)
		}
		rows := [][]any{
			{"id", v.LookupKey},
			{"display_name", v.DisplayName},
			{"code", cliutil.Dash(v.Code)},
			{"created", cliutil.FormatTime(v.CreatedAt)},
			{"internal_id", v.ID},
		}
		return output.Table([]string{"field", "value"}, rows)
	},
}

var createFile string

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a virtual currency",
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
		v, err := client.CreateVirtualCurrency(ctx, body)
		if err != nil {
			return err
		}
		output.Success("created %s", v.LookupKey)
		if output.IsJSON() {
			return output.JSON(v)
		}
		return nil
	},
}

func init() {
	createCmd.Flags().StringVarP(&createFile, "file", "f", "", "JSON body (required)")
	_ = createCmd.MarkFlagRequired("file")
}

var updateFile string

var updateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a virtual currency",
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
		v, err := client.UpdateVirtualCurrency(ctx, args[0], body)
		if err != nil {
			return err
		}
		output.Success("updated %s", v.LookupKey)
		if output.IsJSON() {
			return output.JSON(v)
		}
		return nil
	},
}

func init() {
	updateCmd.Flags().StringVarP(&updateFile, "file", "f", "", "Patch body as JSON (required)")
	_ = updateCmd.MarkFlagRequired("file")
}

var deleteConfirm bool

var deleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a virtual currency",
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
		if err := client.DeleteVirtualCurrency(ctx, args[0]); err != nil {
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
	Short: "Archive a virtual currency",
	Args:  cobra.ExactArgs(1),
	RunE:  archiveAction(true),
}

var unarchiveCmd = &cobra.Command{
	Use:   "unarchive <id>",
	Short: "Unarchive a virtual currency",
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
		if err := client.ArchiveVirtualCurrency(ctx, args[0], archive); err != nil {
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
