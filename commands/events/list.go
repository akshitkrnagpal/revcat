package events

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/akshitkrnagpal/revcat/internal/api"
	"github.com/akshitkrnagpal/revcat/internal/cliutil"
	"github.com/akshitkrnagpal/revcat/internal/output"
)

var (
	listTypes []string
	listSince string
	listLimit int
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Print one page of recent events",
	RunE:  runList,
}

func init() {
	listCmd.Flags().StringSliceVar(&listTypes, "type", nil, "Filter by event type (repeatable). e.g. INITIAL_PURCHASE,CANCELLATION")
	listCmd.Flags().StringVar(&listSince, "since", "", "Only events at or after this time. RFC3339 (2026-04-25T00:00:00Z) or relative (1h, 30m, 7d)")
	listCmd.Flags().IntVar(&listLimit, "limit", 100, "Max events per page (1-1000)")
}

func runList(cmd *cobra.Command, args []string) error {
	client, _, err := cliutil.Client(cmd)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
	defer cancel()

	sinceMS, err := parseSince(listSince)
	if err != nil {
		return err
	}

	page, err := client.ListEvents(ctx, api.ListEventsOptions{
		Types:   listTypes,
		SinceMS: sinceMS,
		Limit:   listLimit,
	})
	if err != nil {
		return err
	}
	return printEvents(page.Items)
}

func printEvents(events []api.Event) error {
	if output.IsJSON() {
		return output.JSON(events)
	}
	rows := make([][]any, 0, len(events))
	for _, e := range events {
		rows = append(rows, []any{formatEventTime(e.OccurredAt), e.Type, e.AppUserID, e.ProductID, dashOrPrice(e.Price, e.Currency), e.Store})
	}
	return output.Table([]string{"when", "type", "user", "product", "price", "store"}, rows)
}

// parseSince accepts RFC3339 ("2026-04-25T00:00:00Z") or a relative
// duration (1h, 30m, 7d, 90s). Empty string is fine and returns 0.
func parseSince(s string) (int64, error) {
	if s == "" {
		return 0, nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.UnixMilli(), nil
	}
	// Allow "7d" by mapping to "168h" before time.ParseDuration.
	if strings.HasSuffix(s, "d") {
		n := strings.TrimSuffix(s, "d")
		if d, err := time.ParseDuration(n + "h"); err == nil {
			d *= 24
			return time.Now().Add(-d).UnixMilli(), nil
		}
	}
	if d, err := time.ParseDuration(s); err == nil {
		return time.Now().Add(-d).UnixMilli(), nil
	}
	return 0, fmt.Errorf("unrecognized --since value %q (expected RFC3339 or duration like 1h/30m/7d)", s)
}

func formatEventTime(ms int64) string {
	if ms == 0 {
		return "-"
	}
	t := time.UnixMilli(ms).Local()
	delta := time.Since(t)
	if delta < time.Minute {
		return fmt.Sprintf("%ds ago", int(delta.Seconds()))
	}
	if delta < time.Hour {
		return fmt.Sprintf("%dm ago", int(delta.Minutes()))
	}
	if delta < 24*time.Hour {
		return t.Format("15:04")
	}
	return t.Format("Jan 02 15:04")
}

func dashOrPrice(price float64, currency string) string {
	if price == 0 {
		return "-"
	}
	return fmt.Sprintf("%.2f %s", price, currency)
}
