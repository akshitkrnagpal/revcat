package subscribers

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/akshitkrnagpal/revcat/internal/api"
	"github.com/akshitkrnagpal/revcat/internal/cliutil"
	"github.com/akshitkrnagpal/revcat/internal/output"
)

var infoCmd = &cobra.Command{
	Use:   "info <user_id>",
	Short: "Show a full debug card for a subscriber",
	Long: `Fan out across the v2 customer endpoints (customer + active_entitlements
+ subscriptions + purchases + aliases) and render a single card.

Pipe the output to a script and you get JSON instead of the card. Each
section appears as a top-level key in the JSON object.

Example:
    revcat subscribers info app_user_123
    revcat subscribers info app_user_123 --output json | jq .entitlements`,
	Args: cobra.ExactArgs(1),
	RunE: runInfo,
}

func runInfo(cmd *cobra.Command, args []string) error {
	customerID := args[0]

	client, _, err := cliutil.Client(cmd)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
	defer cancel()

	snap, err := client.SnapshotCustomer(ctx, customerID)
	if err != nil {
		return err
	}

	if output.IsJSON() {
		return output.JSON(snap)
	}
	return renderCard(snap)
}

// renderCard prints a multi-section TTY view. Layout:
//
//	╭─ subscriber ──────────────────────────────╮
//	│ id          app_user_123                  │
//	│ project     proj_abc                      │
//	│ first seen  2024-01-15                    │
//	│ last seen   today (3h ago)                │
//	╰───────────────────────────────────────────╯
//
//	active entitlements (1)
//	┌──────────┬──────────────┬──────────────┐
//	│ id       │ expires      │ store        │
//	├──────────┼──────────────┼──────────────┤
//	│ premium  │ 2026-05-01   │ app_store    │
//	└──────────┴──────────────┴──────────────┘
//
//	subscriptions (2)  ...
func renderCard(s *api.CustomerSnapshot) error {
	c := s.Customer
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("8")).
		Padding(0, 2)

	header := strings.Builder{}
	fmt.Fprintf(&header, "%-12s %s\n", "id", c.ID)
	fmt.Fprintf(&header, "%-12s %s\n", "project", c.ProjectID)
	if c.Country != "" {
		fmt.Fprintf(&header, "%-12s %s\n", "country", c.Country)
	}
	fmt.Fprintf(&header, "%-12s %s\n", "first seen", formatTime(c.FirstSeen))
	fmt.Fprintf(&header, "%-12s %s", "last seen", formatTime(c.LastSeen))

	fmt.Println(headerStyle.Render("subscriber"))
	fmt.Println(box.Render(strings.TrimRight(header.String(), "\n")))
	fmt.Println()

	renderSection("active entitlements", len(s.Entitlements), func() {
		rows := make([][]any, 0, len(s.Entitlements))
		for _, e := range s.Entitlements {
			expires := "-"
			if e.ExpiresAt != 0 {
				expires = formatTime(e.ExpiresAt)
			}
			renew := "yes"
			if !e.WillRenew {
				renew = "no"
			}
			tag := emptyDash(e.Store)
			if e.IsPromotional {
				tag = "promo"
			}
			rows = append(rows, []any{e.LookupKey, expires, renew, tag})
		}
		_ = output.Table([]string{"id", "expires", "renews", "store"}, rows)
	})

	renderSection("subscriptions", len(s.Subscriptions), func() {
		sort.Slice(s.Subscriptions, func(i, j int) bool {
			return s.Subscriptions[i].StartsAt > s.Subscriptions[j].StartsAt
		})
		rows := make([][]any, 0, len(s.Subscriptions))
		for _, sub := range s.Subscriptions {
			ends := "-"
			if sub.CurrentEndsAt != 0 {
				ends = formatTime(sub.CurrentEndsAt)
			}
			tag := sub.Status
			if sub.IsTrial {
				tag = "trial"
			}
			if sub.IsSandbox {
				tag += " (sandbox)"
			}
			rows = append(rows, []any{sub.ProductID, tag, ends, sub.Store})
		}
		_ = output.Table([]string{"product", "status", "current period ends", "store"}, rows)
	})

	renderSection("purchases", len(s.Purchases), func() {
		sort.Slice(s.Purchases, func(i, j int) bool {
			return s.Purchases[i].PurchaseAt > s.Purchases[j].PurchaseAt
		})
		rows := make([][]any, 0, len(s.Purchases))
		for _, p := range s.Purchases {
			rows = append(rows, []any{p.ProductID, formatTime(p.PurchaseAt), p.Store})
		}
		_ = output.Table([]string{"product", "purchased", "store"}, rows)
	})

	if len(s.Aliases) > 0 {
		renderSection("aliases", len(s.Aliases), func() {
			rows := make([][]any, 0, len(s.Aliases))
			for _, a := range s.Aliases {
				rows = append(rows, []any{a.Alias, a.Type})
			}
			_ = output.Table([]string{"alias", "type"}, rows)
		})
	}

	if len(s.Errors) > 0 {
		fmt.Println(mutedStyle.Render("partial fetch errors:"))
		for k, v := range s.Errors {
			fmt.Println(mutedStyle.Render("  " + k + ": " + v))
		}
	}
	return nil
}

func renderSection(title string, count int, render func()) {
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	header := fmt.Sprintf("%s %s", headerStyle.Render(title), mutedStyle.Render(fmt.Sprintf("(%d)", count)))
	fmt.Println(header)
	if count == 0 {
		fmt.Println(mutedStyle.Render("  (none)"))
	} else {
		render()
	}
	fmt.Println()
}

func formatTime(unix int64) string {
	if unix == 0 {
		return "-"
	}
	// RC mixes seconds and ms in different endpoints; assume ms when value
	// looks too big for seconds (after year 5000-ish).
	t := time.Unix(unix, 0).UTC()
	if unix > 9999999999 {
		t = time.UnixMilli(unix).UTC()
	}
	delta := time.Since(t)
	switch {
	case delta < 0:
		return t.Format("2006-01-02") + " (in " + humanize(-delta) + ")"
	case delta < 24*time.Hour:
		return "today (" + humanize(delta) + " ago)"
	case delta < 7*24*time.Hour:
		return t.Format("Mon") + " (" + humanize(delta) + " ago)"
	default:
		return t.Format("2006-01-02")
	}
}

func humanize(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dd", int(d.Hours()/24))
}

func emptyDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
