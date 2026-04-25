package output

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// Table renders rows in the active output format. Headers are used as
// JSON keys when the format is JSON. Cells are converted to strings
// blindly for the table renderer.
//
// Example:
//
//	Table([]string{"name", "id"}, [][]any{{"premium", "ent_a"}, {"pro", "ent_b"}})
func Table(headers []string, rows [][]any) error {
	switch Resolved() {
	case FormatJSON:
		return JSON(rowsToObjects(headers, rows))
	case FormatCSV:
		return printCSV(headers, rows)
	case FormatMarkdown:
		return printMarkdown(headers, rows)
	default:
		return printTable(headers, rows)
	}
}

func rowsToObjects(headers []string, rows [][]any) []map[string]any {
	out := make([]map[string]any, 0, len(rows))
	for _, r := range rows {
		o := make(map[string]any, len(headers))
		for i, h := range headers {
			if i < len(r) {
				o[h] = r[i]
			}
		}
		out = append(out, o)
	}
	return out
}

func printTable(headers []string, rows [][]any) error {
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	cellStyle := lipgloss.NewStyle().PaddingRight(2)

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("8"))).
		Headers(headers...).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle.PaddingRight(2)
			}
			return cellStyle
		})

	for _, r := range rows {
		strs := make([]string, len(r))
		for i, c := range r {
			strs[i] = fmt.Sprint(c)
		}
		t.Row(strs...)
	}
	fmt.Fprintln(current.stdout, t.Render())
	return nil
}

func printCSV(headers []string, rows [][]any) error {
	fmt.Fprintln(current.stdout, strings.Join(headers, ","))
	for _, r := range rows {
		strs := make([]string, len(r))
		for i, c := range r {
			s := fmt.Sprint(c)
			if strings.ContainsAny(s, ",\"\n") {
				s = `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
			}
			strs[i] = s
		}
		fmt.Fprintln(current.stdout, strings.Join(strs, ","))
	}
	return nil
}

func printMarkdown(headers []string, rows [][]any) error {
	fmt.Fprintln(current.stdout, "| "+strings.Join(headers, " | ")+" |")
	sep := make([]string, len(headers))
	for i := range sep {
		sep[i] = "---"
	}
	fmt.Fprintln(current.stdout, "| "+strings.Join(sep, " | ")+" |")
	for _, r := range rows {
		strs := make([]string, len(r))
		for i, c := range r {
			strs[i] = strings.ReplaceAll(fmt.Sprint(c), "|", "\\|")
		}
		fmt.Fprintln(current.stdout, "| "+strings.Join(strs, " | ")+" |")
	}
	return nil
}
