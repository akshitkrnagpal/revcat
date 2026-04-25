// Package metrics holds `revcat metrics ...` and `revcat charts ...`.
package metrics

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

var Cmd = &cobra.Command{
	Use:   "metrics",
	Short: "Project-level revenue + subscription metrics",
}

func init() {
	Cmd.AddCommand(overviewCmd)
}

var overviewCmd = &cobra.Command{
	Use:   "overview",
	Short: "Headline metrics for the active project",
	Long: `Returns the dashboard's headline numbers (active subscribers, MRR,
lifetime revenue, conversion). Shape is preserved verbatim; the table
view flattens top-level keys.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		m, err := client.GetMetricsOverview(ctx)
		if err != nil {
			return err
		}
		if output.IsJSON() {
			return output.JSON(m)
		}
		rows := make([][]any, 0, len(m))
		for k, v := range m {
			rows = append(rows, []any{k, fmt.Sprint(v)})
		}
		return output.Table([]string{"metric", "value"}, rows)
	},
}

// ChartsCmd is a sibling top-level group registered separately.
var ChartsCmd = &cobra.Command{
	Use:   "charts",
	Short: "Project charts (revenue, active subs, conversion, etc.)",
	Long: `Charts mirror the dashboard graphs: revenue, active subscribers,
conversion, MRR, churn, etc. Run ` + "`revcat charts options <name>`" + ` for
the supported filters before requesting data.`,
}

func init() {
	ChartsCmd.AddCommand(getCmd, optionsCmd)
}

var (
	getStart  string
	getEnd    string
	getPeriod string
	getFilter []string
)

var getCmd = &cobra.Command{
	Use:   "get <chart_name>",
	Short: "Fetch chart data (raw JSON)",
	Long: `Fetch chart data. Filters via --filter key=value (repeatable). Date
range via --start / --end (YYYY-MM-DD). Period via --period (day | week | month).`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filters := map[string]string{}
		for _, kv := range getFilter {
			i := strings.IndexByte(kv, '=')
			if i < 0 {
				return fmt.Errorf("bad --filter %q (expected key=value)", kv)
			}
			filters[kv[:i]] = kv[i+1:]
		}
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		c, err := client.GetChart(ctx, args[0], api.ChartOptions{
			StartDate: getStart,
			EndDate:   getEnd,
			Period:    getPeriod,
			Filters:   filters,
		})
		if err != nil {
			return err
		}
		return output.JSON(c)
	},
}

func init() {
	getCmd.Flags().StringVar(&getStart, "start", "", "Start date YYYY-MM-DD")
	getCmd.Flags().StringVar(&getEnd, "end", "", "End date YYYY-MM-DD")
	getCmd.Flags().StringVar(&getPeriod, "period", "", "day | week | month")
	getCmd.Flags().StringArrayVar(&getFilter, "filter", nil, "key=value (repeatable)")
}

var optionsCmd = &cobra.Command{
	Use:   "options <chart_name>",
	Short: "Show the available filters/dimensions for a chart",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		opts, err := client.GetChartOptions(ctx, args[0])
		if err != nil {
			return err
		}
		return output.JSON(opts)
	},
}
