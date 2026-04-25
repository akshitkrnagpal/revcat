package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/akshitkrnagpal/revcat/internal/api"
	authstore "github.com/akshitkrnagpal/revcat/internal/auth"
	"github.com/akshitkrnagpal/revcat/internal/output"
)

var (
	statusValidate bool
	statusName     string
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the active auth profile",
	Long: `Print the active auth profile and where it's stored. Pass --validate
to also hit the RevenueCat API and confirm the key still works.`,
	RunE: runStatus,
}

func init() {
	statusCmd.Flags().BoolVar(&statusValidate, "validate", false, "Hit the API to confirm the key is accepted")
	statusCmd.Flags().StringVarP(&statusName, "name", "n", "", "Profile name (default: REVCAT_PROFILE or 'default')")
}

func runStatus(cmd *cobra.Command, args []string) error {
	store, err := authstore.Open(bypassKeychain(cmd))
	if err != nil {
		return err
	}
	profile, err := authstore.Resolve(store, statusName)
	if err != nil {
		return err
	}

	source := "OS keychain"
	if bypassKeychain(cmd) {
		source = "./.revcat/config.json"
	}
	if profile.Name == "$REVCAT_API_KEY" {
		source = "REVCAT_API_KEY env"
	}

	rows := [][]any{
		{"profile", profile.Name},
		{"source", source},
		{"secret_key", redactKey(profile.SecretKey)},
		{"project_id", emptyDash(profile.ProjectID)},
	}

	if statusValidate {
		client := api.New(api.Options{SecretKey: profile.SecretKey, ProjectID: profile.ProjectID, Version: cmd.Root().Version})
		ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Second)
		defer cancel()
		projects, err := client.ListProjects(ctx)
		if err != nil {
			rows = append(rows, []any{"validation", fmt.Sprintf("FAILED: %v", err)})
			if err := output.Table([]string{"field", "value"}, rows); err != nil {
				return err
			}
			return err
		}
		rows = append(rows, []any{"validation", fmt.Sprintf("OK (%d project access)", len(projects))})
	}

	return output.Table([]string{"field", "value"}, rows)
}

func emptyDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
