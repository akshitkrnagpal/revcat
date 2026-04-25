// Package invoices holds `revcat invoices view`.
package invoices

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/akshitkrnagpal/revcat/internal/cliutil"
	"github.com/akshitkrnagpal/revcat/internal/output"
)

var Cmd = &cobra.Command{
	Use:   "invoices",
	Short: "Inspect invoices",
}

func init() {
	Cmd.AddCommand(viewCmd)
}

var viewCmd = &cobra.Command{
	Use:   "view <id>",
	Short: "Show one invoice (raw JSON)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		inv, err := client.GetInvoice(ctx, args[0])
		if err != nil {
			return err
		}
		return output.JSON(inv)
	},
}
