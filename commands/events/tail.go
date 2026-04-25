package events

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/akshitkrnagpal/revcat/internal/api"
	"github.com/akshitkrnagpal/revcat/internal/cliutil"
	"github.com/akshitkrnagpal/revcat/internal/output"
)

var (
	tailTypes    []string
	tailSince    string
	tailInterval time.Duration
)

var tailCmd = &cobra.Command{
	Use:   "tail",
	Short: "Follow events as they arrive",
	Long: `Poll the events endpoint at --interval and print new events as they
appear, like ` + "`kubectl logs -f`" + ` for subscription activity.

Filters apply to every poll (no --filter set means all event types).
The --since flag controls where the tail STARTS; subsequent polls track
the highest occurred_at seen so far. Press Ctrl-C to stop.

Examples:
    revcat events tail
    revcat events tail --type INITIAL_PURCHASE,CANCELLATION
    revcat events tail --since 1h --interval 5s
    revcat events tail --output json | jq '.type'    # JSON-line per event`,
	RunE: runTail,
}

func init() {
	tailCmd.Flags().StringSliceVar(&tailTypes, "type", nil, "Filter by event type (repeatable). e.g. INITIAL_PURCHASE,CANCELLATION")
	tailCmd.Flags().StringVar(&tailSince, "since", "5m", "Where to start. RFC3339 or duration (1h, 30m, 7d). Default 5m.")
	tailCmd.Flags().DurationVar(&tailInterval, "interval", 5*time.Second, "Poll interval (min 2s)")
}

func runTail(cmd *cobra.Command, args []string) error {
	if tailInterval < 2*time.Second {
		return errors.New("--interval must be >= 2s to avoid rate limits")
	}

	client, _, err := cliutil.Client(cmd)
	if err != nil {
		return err
	}

	sinceMS, err := parseSince(tailSince)
	if err != nil {
		return err
	}

	// Stop on SIGINT / SIGTERM. cobra's cmd.Context() is plain background.
	ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	if !output.IsJSON() {
		fmt.Fprintln(os.Stderr, mutedStyle.Render(fmt.Sprintf("tailing events (interval=%s, since=%s, types=%v); ctrl-c to stop", tailInterval, tailSince, tailTypes)))
	}

	highest := sinceMS
	first := true
	for {
		page, err := client.ListEvents(ctx, api.ListEventsOptions{
			Types:   tailTypes,
			SinceMS: highest,
			Limit:   100,
		})
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			output.Warn("poll: %v", err)
		} else {
			// Sort ascending by occurred_at so output is chronological even
			// if RC returns the page in reverse order.
			sort.Slice(page.Items, func(i, j int) bool {
				return page.Items[i].OccurredAt < page.Items[j].OccurredAt
			})
			for _, ev := range page.Items {
				if ev.OccurredAt <= highest && !first {
					continue
				}
				if err := emit(ev); err != nil {
					return err
				}
				if ev.OccurredAt > highest {
					highest = ev.OccurredAt
				}
			}
			first = false
		}

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(tailInterval):
		}
	}
}

// emit prints a single event in the active format. Each event is a
// standalone JSON object on its own line in JSON mode (ndjson), so the
// stream stays parseable mid-tail.
func emit(e api.Event) error {
	if output.IsJSON() {
		return output.JSON(e)
	}
	when := time.UnixMilli(e.OccurredAt).Local().Format("15:04:05")
	typeStyle := lipgloss.NewStyle().Bold(true).Foreground(typeColor(e.Type))
	priceStr := ""
	if e.Price > 0 {
		priceStr = fmt.Sprintf("  %.2f %s", e.Price, e.Currency)
	}
	store := ""
	if e.Store != "" {
		store = "  " + e.Store
	}
	fmt.Printf("%s  %s  %s  %s%s%s\n", when, typeStyle.Render(padType(e.Type)), e.AppUserID, e.ProductID, priceStr, store)
	return nil
}

func padType(t string) string {
	if len(t) >= 22 {
		return t
	}
	return t + spaces(22-len(t))
}

func spaces(n int) string {
	out := make([]byte, n)
	for i := range out {
		out[i] = ' '
	}
	return string(out)
}

func typeColor(t string) lipgloss.Color {
	switch t {
	case "INITIAL_PURCHASE", "RENEWAL", "PRODUCT_CHANGE":
		return lipgloss.Color("2") // green
	case "TRIAL_STARTED", "TRIAL_CONVERTED":
		return lipgloss.Color("4") // blue
	case "CANCELLATION", "EXPIRATION":
		return lipgloss.Color("3") // yellow
	case "REFUND", "BILLING_ISSUE":
		return lipgloss.Color("1") // red
	default:
		return lipgloss.Color("7")
	}
}
