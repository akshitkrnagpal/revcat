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
		credential := redactKey(p.SecretKey)
		if p.EffectiveAuthType() == authstore.AuthTypeOAuth {
			credential = "oauth:" + redactKey(p.AccessToken)
		}
		rows = append(rows, []any{marker, n, string(p.EffectiveAuthType()), credential, emptyDash(p.ProjectID)})
	}
	return output.Table([]string{"", "name", "auth_type", "credential", "project_id"}, rows)
}
