// Package virtualcurrencies holds `revcat virtual-currencies ...`.
package virtualcurrencies

import (
	"context"
	"errors"
	"strings"
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
	Long: `Project-level virtual currencies (in-game coins, credits, tokens).
v2 keys VCs by their uppercase code (e.g., COIN, GEM) - that's the
identifier you pass to view/update/delete/archive.

Per-customer balances and transactions are NOT exposed by v2 REST.`,
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
			rows = append(rows, []any{v.Code, v.Name, cliutil.Dash(v.Description), cliutil.FormatTime(v.CreatedAt)})
		}
		return output.Table([]string{"code", "name", "description", "created"}, rows)
	},
}

var viewCmd = &cobra.Command{
	Use:   "view <code>",
	Short: "Show one virtual currency by uppercase code (e.g. COIN)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		v, raw, err := client.GetVirtualCurrencyRaw(ctx, strings.ToUpper(args[0]))
		if err != nil {
			return err
		}
		if output.IsJSON() {
			return output.JSON(raw)
		}
		rows := [][]any{
			{"code", v.Code},
			{"name", v.Name},
			{"description", cliutil.Dash(v.Description)},
			{"state", cliutil.Dash(v.State)},
			{"created", cliutil.FormatTime(v.CreatedAt)},
		}
		return output.Table([]string{"field", "value"}, rows)
	},
}

var (
	createFile string
	createName string
	createCode string
	createDesc string
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a virtual currency",
	Long: `Create a virtual currency. Required: name + code. Code is uppercase
and acts as the identifier for view/update/delete.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var body map[string]any
		if createFile != "" {
			b, err := cliutil.LoadJSON(createFile)
			if err != nil {
				return err
			}
			body = b
		} else if createName != "" && createCode != "" {
			body = map[string]any{
				"name": createName,
				"code": strings.ToUpper(createCode),
			}
			if createDesc != "" {
				body["description"] = createDesc
			}
		} else {
			return errors.New("pass --file or --name + --code")
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
		output.Success("created %s", v.Code)
		if output.IsJSON() {
			return output.JSON(v)
		}
		return nil
	},
}

func init() {
	createCmd.Flags().StringVarP(&createFile, "file", "f", "", "JSON body")
	createCmd.Flags().StringVar(&createName, "name", "", "Display name")
	createCmd.Flags().StringVar(&createCode, "code", "", "Uppercase code (e.g. COIN)")
	createCmd.Flags().StringVar(&createDesc, "description", "", "Optional description")
}

var (
	updateFile string
	updateName string
	updateDesc string
)

var updateCmd = &cobra.Command{
	Use:   "update <code>",
	Short: "Update a virtual currency",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		body := map[string]any{}
		if updateFile != "" {
			b, err := cliutil.LoadJSON(updateFile)
			if err != nil {
				return err
			}
			for k, v := range b {
				body[k] = v
			}
		}
		if updateName != "" {
			body["name"] = updateName
		}
		if updateDesc != "" {
			body["description"] = updateDesc
		}
		if len(body) == 0 {
			return errors.New("pass --file or --name / --description")
		}
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		v, err := client.UpdateVirtualCurrency(ctx, strings.ToUpper(args[0]), body)
		if err != nil {
			return err
		}
		output.Success("updated %s", v.Code)
		if output.IsJSON() {
			return output.JSON(v)
		}
		return nil
	},
}

func init() {
	updateCmd.Flags().StringVarP(&updateFile, "file", "f", "", "Patch body as JSON")
	updateCmd.Flags().StringVar(&updateName, "name", "", "New name")
	updateCmd.Flags().StringVar(&updateDesc, "description", "", "New description")
}

var deleteConfirm bool

var deleteCmd = &cobra.Command{
	Use:   "delete <code>",
	Short: "Delete a virtual currency",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		code := strings.ToUpper(args[0])
		if !deleteConfirm {
			var ok bool
			if err := survey.AskOne(&survey.Confirm{Message: "delete " + code + "?", Default: false}, &ok); err != nil {
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
		if err := client.DeleteVirtualCurrency(ctx, code); err != nil {
			return err
		}
		output.Success("deleted %s", code)
		return nil
	},
}

func init() {
	deleteCmd.Flags().BoolVarP(&deleteConfirm, "confirm", "y", false, "Skip the prompt")
}

var archiveCmd = &cobra.Command{
	Use:   "archive <code>",
	Short: "Archive a virtual currency",
	Args:  cobra.ExactArgs(1),
	RunE:  archiveAction(true),
}

var unarchiveCmd = &cobra.Command{
	Use:   "unarchive <code>",
	Short: "Unarchive a virtual currency",
	Args:  cobra.ExactArgs(1),
	RunE:  archiveAction(false),
}

func archiveAction(archive bool) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		code := strings.ToUpper(args[0])
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		if err := client.ArchiveVirtualCurrency(ctx, code, archive); err != nil {
			return err
		}
		verb := "archived"
		if !archive {
			verb = "unarchived"
		}
		output.Success("%s %s", verb, code)
		return nil
	}
}
