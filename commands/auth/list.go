package auth

import (
	"github.com/spf13/cobra"

	authstore "github.com/akshitkrnagpal/revcat/internal/auth"
	"github.com/akshitkrnagpal/revcat/internal/output"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List stored auth profiles",
	RunE:  runList,
}

func runList(cmd *cobra.Command, args []string) error {
	store, err := authstore.Open(bypassKeychain(cmd))
	if err != nil {
		return err
	}
	names, err := store.List()
	if err != nil {
		return err
	}
	active, _ := authstore.GetActive()
	rows := make([][]any, 0, len(names))
	for _, n := range names {
		marker := ""
		if n == active {
			marker = "*"
		}
		p, err := store.Get(n)
		if err != nil {
			rows = append(rows, []any{marker, n, "(unreadable)", ""})
			continue
		}
		rows = append(rows, []any{marker, n, redactKey(p.SecretKey), emptyDash(p.ProjectID)})
	}
	return output.Table([]string{"", "name", "secret_key", "project_id"}, rows)
}
