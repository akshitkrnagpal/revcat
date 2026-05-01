package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/akshitkrnagpal/revcat/internal/api"
	authstore "github.com/akshitkrnagpal/revcat/internal/auth"
	"github.com/akshitkrnagpal/revcat/internal/cliutil"
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
	name := statusName
	if name == "" {
		name = cliutil.Profile(cmd)
	}
	profile, err := authstore.Resolve(store, name)
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
		{"auth_type", string(profile.EffectiveAuthType())},
	}
	if profile.EffectiveAuthType() == authstore.AuthTypeOAuth {
		rows = append(rows,
			[]any{"access_token", redactKey(profile.AccessToken)},
			[]any{"expires", expiresLine(profile.ExpiresAt)},
			[]any{"scope", emptyDash(profile.Scope)},
		)
	} else {
		rows = append(rows, []any{"secret_key", redactKey(profile.SecretKey)})
	}
	rows = append(rows, []any{"project_id", emptyDash(profile.ProjectID)})

	if statusValidate {
		opts := api.Options{ProjectID: profile.ProjectID, Version: cmd.Root().Version}
		if profile.EffectiveAuthType() == authstore.AuthTypeOAuth {
			opts.TokenSource = authstore.NewOAuthTokenSource(store, profile)
		} else {
			opts.SecretKey = profile.SecretKey
		}
		client := api.New(opts)
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

func expiresLine(ms int64) string {
	if ms == 0 {
		return "-"
	}
	t := time.UnixMilli(ms).Local()
	delta := time.Until(t)
	if delta < 0 {
		return t.Format("2006-01-02 15:04 MST") + " (EXPIRED)"
	}
	return t.Format("2006-01-02 15:04 MST")
}

func emptyDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
