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
renewals, cancellations, refunds, ...). Each webhook has a name, target
URL, and a list of event_types it subscribes to.

Event values are LOWERCASE in the API config (initial_purchase,
renewal, cancellation, ...) - even though the webhook payload itself
uses screaming case (INITIAL_PURCHASE). revcat lowercases values
passed via --events for you.`,
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
			rows = append(rows, []any{h.ID, h.Name, h.URL, len(h.EventTypes), cliutil.FormatTime(h.CreatedAt)})
		}
		return output.Table([]string{"id", "name", "url", "events", "created"}, rows)
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
			{"name", h.Name},
			{"url", h.URL},
			{"event_types", strings.Join(h.EventTypes, ", ")},
			{"app_id", cliutil.Dash(h.AppID)},
			{"environment", cliutil.Dash(h.Environment)},
			{"created", cliutil.FormatTime(h.CreatedAt)},
		}
		return output.Table([]string{"field", "value"}, rows)
	},
}

var (
	createFile   string
	createName   string
	createURL    string
	createEvents []string
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a webhook integration",
	Long: `Create a webhook. Required: name, url, event_types. URL must be a
valid HTTPS endpoint that RC can validate as reachable - localhost and
example.com don't pass.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var body map[string]any
		if createFile != "" {
			b, err := cliutil.LoadJSON(createFile)
			if err != nil {
				return err
			}
			body = b
			normalizeEventTypes(body)
		} else if createURL != "" && createName != "" {
			body = map[string]any{
				"name":        createName,
				"url":         createURL,
				"event_types": lowerSlice(createEvents),
			}
		} else {
			return errors.New("pass --file or --name + --url + --events")
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
	createCmd.Flags().StringVar(&createName, "name", "", "Webhook name (required)")
	createCmd.Flags().StringVar(&createURL, "url", "", "Target URL (required)")
	createCmd.Flags().StringSliceVar(&createEvents, "events", nil, "Event types (comma-separated). Values are lowercased before send.")
}

var (
	updateFile   string
	updateName   string
	updateURL    string
	updateEvents []string
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
			normalizeEventTypes(body)
		}
		if updateName != "" {
			body["name"] = updateName
		}
		if updateURL != "" {
			body["url"] = updateURL
		}
		if len(updateEvents) > 0 {
			body["event_types"] = lowerSlice(updateEvents)
		}
		if len(body) == 0 {
			return errors.New("pass --file or one of --name / --url / --events")
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
	updateCmd.Flags().StringVar(&updateName, "name", "", "New name")
	updateCmd.Flags().StringVar(&updateURL, "url", "", "New target URL")
	updateCmd.Flags().StringSliceVar(&updateEvents, "events", nil, "New event_types list")
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

// lowerSlice returns a new slice with each element lowercased.
func lowerSlice(in []string) []string {
	out := make([]string, len(in))
	for i, s := range in {
		out[i] = strings.ToLower(s)
	}
	return out
}

// normalizeEventTypes lowercases any event_types entries inside a body.
// v2 rejects screaming-case event names like INITIAL_PURCHASE.
func normalizeEventTypes(body map[string]any) {
	raw, ok := body["event_types"].([]any)
	if !ok {
		return
	}
	out := make([]any, len(raw))
	for i, v := range raw {
		if s, ok := v.(string); ok {
			out[i] = strings.ToLower(s)
		} else {
			out[i] = v
		}
	}
	body["event_types"] = out
}
