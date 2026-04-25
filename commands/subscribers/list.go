package subscribers

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/akshitkrnagpal/revcat/internal/cliutil"
	"github.com/akshitkrnagpal/revcat/internal/output"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List customers in the active project (paged)",
	Long: `Page through every customer. For support workflows that look up a
specific user by email or order id, prefer the searches under
` + "`revcat purchases search`" + ` and ` + "`revcat subscriptions search`" + ` (faster
and indexed).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 60*time.Second)
		defer cancel()
		customers, err := client.ListCustomers(ctx)
		if err != nil {
			return err
		}
		rows := make([][]any, 0, len(customers))
		for _, c := range customers {
			rows = append(rows, []any{c.ID, cliutil.Dash(c.Country), cliutil.FormatTime(c.FirstSeen), cliutil.FormatTime(c.LastSeen), c.ActiveEntCount})
		}
		return output.Table([]string{"id", "country", "first_seen", "last_seen", "active_ents"}, rows)
	},
}

func init() {
	Cmd.AddCommand(listCmd)
}
