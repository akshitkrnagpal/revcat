package subscribers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"

	"github.com/akshitkrnagpal/revcat/internal/api"
	"github.com/akshitkrnagpal/revcat/internal/cliutil"
	"github.com/akshitkrnagpal/revcat/internal/output"
)

var createFile string

var createCmd = &cobra.Command{
	Use:   "create <user_id>",
	Short: "Pre-create a customer (migration / import)",
	Long: `Pre-create a customer record. Most apps let the SDK create customers
on first launch; this is for migrations, test seeding, or cases where
you want to seed attributes before the user opens the app.

Pass --file <path.json> for arbitrary v2 fields (attributes, etc.).`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		body := map[string]any{"id": args[0]}
		if createFile != "" {
			b, err := cliutil.LoadJSON(createFile)
			if err != nil {
				return err
			}
			for k, v := range b {
				body[k] = v
			}
		}
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		c, err := client.CreateCustomer(ctx, body)
		if err != nil {
			return err
		}
		output.Success("created %s", c.ID)
		if output.IsJSON() {
			return output.JSON(c)
		}
		return nil
	},
}

func init() {
	createCmd.Flags().StringVarP(&createFile, "file", "f", "", "Optional JSON body merged into the request")
	Cmd.AddCommand(createCmd)
}

var deleteConfirm bool

var deleteCmd = &cobra.Command{
	Use:   "delete <user_id>",
	Short: "Permanently delete a customer (GDPR / test cleanup)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !deleteConfirm {
			var ok bool
			if err := survey.AskOne(&survey.Confirm{
				Message: "permanently delete customer " + args[0] + "? this cannot be undone.",
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
		if err := client.DeleteCustomer(ctx, args[0]); err != nil {
			return err
		}
		output.Success("deleted %s", args[0])
		return nil
	},
}

func init() {
	deleteCmd.Flags().BoolVarP(&deleteConfirm, "confirm", "y", false, "Skip the prompt")
	Cmd.AddCommand(deleteCmd)
}

var transferConfirm bool

var transferCmd = &cobra.Command{
	Use:   "transfer <src_user_id> <dst_user_id>",
	Short: "Transfer entitlements/subscriptions from one customer to another",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !transferConfirm {
			var ok bool
			if err := survey.AskOne(&survey.Confirm{
				Message: "transfer everything from " + args[0] + " to " + args[1] + "?",
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
		if err := client.TransferCustomer(ctx, args[0], args[1]); err != nil {
			return err
		}
		output.Success("transferred %s -> %s", args[0], args[1])
		return nil
	},
}

func init() {
	transferCmd.Flags().BoolVarP(&transferConfirm, "confirm", "y", false, "Skip the prompt")
	Cmd.AddCommand(transferCmd)
}

// Note: revcat used to expose `subscribers override-offering` and
// `subscribers restore-google-play`. Both endpoints 404 against v2 in
// our smoke test - they were v1 / dashboard-only actions and never
// got a v2 customer-scoped equivalent. Removed from the CLI.

var (
	attrsSetFile string
	attrsSetKV   []string
)

var attrsCmd = &cobra.Command{
	Use:   "attributes <user_id>",
	Short: "Get or set subscriber attributes",
	Long: `With no flags, lists current attributes. With --set key=value (repeatable)
or --file <path.json>, upserts the listed attributes.

The v2 attributes endpoint takes an array of {name, value} objects.
Both the file and --set forms are normalized to that shape before
sending.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		if attrsSetFile != "" || len(attrsSetKV) > 0 {
			merged := map[string]string{}
			if attrsSetFile != "" {
				b, err := cliutil.LoadJSON(attrsSetFile)
				if err != nil {
					return err
				}
				for k, v := range b {
					merged[k] = fmt.Sprint(v)
				}
			}
			for _, kv := range attrsSetKV {
				k, v := splitKV(kv)
				if k == "" {
					return errors.New("bad --set value: " + kv + " (expected key=value)")
				}
				merged[k] = v
			}
			attrs := make([]api.CustomerAttribute, 0, len(merged))
			for k, v := range merged {
				attrs = append(attrs, api.CustomerAttribute{Name: k, Value: v})
			}
			if err := client.SetAttributes(ctx, args[0], attrs); err != nil {
				return err
			}
			output.Success("set %d attribute(s) on %s", len(attrs), args[0])
			return nil
		}
		out, err := client.GetAttributes(ctx, args[0])
		if err != nil {
			return err
		}
		if output.IsJSON() {
			return output.JSON(out)
		}
		rows := make([][]any, 0, len(out))
		for _, a := range out {
			rows = append(rows, []any{a.Name, a.Value})
		}
		return output.Table([]string{"key", "value"}, rows)
	},
}

func init() {
	attrsCmd.Flags().StringVar(&attrsSetFile, "file", "", "JSON map of attributes to upsert")
	attrsCmd.Flags().StringArrayVar(&attrsSetKV, "set", nil, "key=value to upsert (repeatable)")
	Cmd.AddCommand(attrsCmd)
}

func splitKV(s string) (string, string) {
	for i := 0; i < len(s); i++ {
		if s[i] == '=' {
			return s[:i], s[i+1:]
		}
	}
	return "", ""
}

var invoicesCmd = &cobra.Command{
	Use:   "invoices <user_id>",
	Short: "List invoices for a customer",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		invs, err := client.ListInvoices(ctx, args[0])
		if err != nil {
			return err
		}
		return output.JSON(invs)
	},
}

func init() {
	Cmd.AddCommand(invoicesCmd)
}
