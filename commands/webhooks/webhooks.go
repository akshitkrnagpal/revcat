// Package webhooks holds `revcat webhooks ...`.
package webhooks

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
	Use:   "webhooks",
	Short: "Manage webhook integrations",
	Long: `Webhooks are project integrations that receive event POSTs (purchases,
renewals, cancellations, refunds, ...). Each webhook has a target URL,
a list of events it subscribes to, and a disabled flag.`,
}

func init() {
	Cmd.AddCommand(listCmd, viewCmd, createCmd, updateCmd, deleteCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List webhook integrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		hooks, err := client.ListWebhooks(ctx)
		if err != nil {
			return err
		}
		rows := make([][]any, 0, len(hooks))
		for _, h := range hooks {
			marker := ""
			if h.Disabled {
				marker = "off"
			}
			rows = append(rows, []any{marker, h.ID, h.URL, len(h.Events), cliutil.FormatTime(h.CreatedAt)})
		}
		return output.Table([]string{"", "id", "url", "events", "created"}, rows)
	},
}

var viewCmd = &cobra.Command{
	Use:   "view <id>",
	Short: "Show one webhook",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		h, err := client.GetWebhook(ctx, args[0])
		if err != nil {
			return err
		}
		if output.IsJSON() {
			return output.JSON(h)
		}
		rows := [][]any{
			{"id", h.ID},
			{"url", h.URL},
			{"description", cliutil.Dash(h.Description)},
			{"disabled", h.Disabled},
			{"events", strings.Join(h.Events, ", ")},
			{"created", cliutil.FormatTime(h.CreatedAt)},
		}
		return output.Table([]string{"field", "value"}, rows)
	},
}

var (
	createFile   string
	createURL    string
	createEvents []string
	createDesc   string
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a webhook integration",
	RunE: func(cmd *cobra.Command, args []string) error {
		var body map[string]any
		if createFile != "" {
			b, err := cliutil.LoadJSON(createFile)
			if err != nil {
				return err
			}
			body = b
		} else if createURL != "" {
			body = map[string]any{"url": createURL}
			if len(createEvents) > 0 {
				body["events"] = createEvents
			}
			if createDesc != "" {
				body["description"] = createDesc
			}
		} else {
			return errors.New("pass --file or --url")
		}
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		h, err := client.CreateWebhook(ctx, body)
		if err != nil {
			return err
		}
		output.Success("created %s", h.ID)
		if output.IsJSON() {
			return output.JSON(h)
		}
		return nil
	},
}

func init() {
	createCmd.Flags().StringVarP(&createFile, "file", "f", "", "JSON body")
	createCmd.Flags().StringVar(&createURL, "url", "", "Target URL")
	createCmd.Flags().StringSliceVar(&createEvents, "events", nil, "Events to subscribe to (comma-separated)")
	createCmd.Flags().StringVar(&createDesc, "description", "", "Optional description")
}

var (
	updateFile     string
	updateURL      string
	updateEvents   []string
	updateDisabled string
)

var updateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a webhook",
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
		if updateURL != "" {
			body["url"] = updateURL
		}
		if len(updateEvents) > 0 {
			body["events"] = updateEvents
		}
		switch updateDisabled {
		case "true":
			body["disabled"] = true
		case "false":
			body["disabled"] = false
		}
		if len(body) == 0 {
			return errors.New("pass --file or one of --url / --events / --disabled")
		}
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		h, err := client.UpdateWebhook(ctx, args[0], body)
		if err != nil {
			return err
		}
		output.Success("updated %s", h.ID)
		if output.IsJSON() {
			return output.JSON(h)
		}
		return nil
	},
}

func init() {
	updateCmd.Flags().StringVarP(&updateFile, "file", "f", "", "Patch body as JSON")
	updateCmd.Flags().StringVar(&updateURL, "url", "", "New target URL")
	updateCmd.Flags().StringSliceVar(&updateEvents, "events", nil, "New events list")
	updateCmd.Flags().StringVar(&updateDisabled, "disabled", "", "Set 'true' or 'false'")
}

var deleteConfirm bool

var deleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a webhook",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !deleteConfirm {
			var ok bool
			if err := survey.AskOne(&survey.Confirm{Message: "delete webhook " + args[0] + "?", Default: false}, &ok); err != nil {
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
		if err := client.DeleteWebhook(ctx, args[0]); err != nil {
			return err
		}
		output.Success("deleted %s", args[0])
		return nil
	},
}

func init() {
	deleteCmd.Flags().BoolVarP(&deleteConfirm, "confirm", "y", false, "Skip the prompt")
}
