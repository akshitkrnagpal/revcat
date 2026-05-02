package auth

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/akshitkrnagpal/revcat/internal/api"
	authstore "github.com/akshitkrnagpal/revcat/internal/auth"
	"github.com/akshitkrnagpal/revcat/internal/cliutil"
	"github.com/akshitkrnagpal/revcat/internal/output"
	"github.com/akshitkrnagpal/revcat/internal/project"
)

var (
	statusValidate bool
	statusName     string
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the active auth profile and resolved project",
	Long: `Print the active credential and where it came from (local
.revcat/config.json, global ~/.revcat/config.json, or env hatch). Pass
--validate to also hit the RevenueCat API and confirm the token is
accepted.`,
	RunE: runStatus,
}

func init() {
	statusCmd.Flags().BoolVar(&statusValidate, "validate", false, "Hit the API to confirm the credential is accepted")
	statusCmd.Flags().StringVarP(&statusName, "name", "n", "", "Override the active global profile (ignored when a local config is present)")
}

func runStatus(cmd *cobra.Command, args []string) error {
	flagProfile := statusName
	if flagProfile == "" {
		flagProfile = cliutil.Profile(cmd)
	}
	resolved, err := authstore.Resolve(authstore.ResolveOptions{
		FlagProfile: flagProfile,
	})
	if err != nil {
		return err
	}

	rows := [][]any{
		{"profile", resolved.Profile.Name},
		{"source", string(resolved.Source)},
	}
	if resolved.Path != "" {
		rows = append(rows, []any{"source_path", resolved.Path})
	}
	rows = append(rows,
		[]any{"access_token", redactKey(resolved.Profile.AccessToken)},
		[]any{"expires", expiresLine(resolved.Profile.ExpiresAt)},
		[]any{"scope", emptyDash(resolved.Profile.Scope)},
	)

	resolvedProject := cliutil.ResolveProjectID(cmd, resolved)
	rows = append(rows,
		[]any{"project_id", emptyDash(resolvedProject)},
		[]any{"project_source", projectSource(cmd, resolved)},
	)

	if statusValidate {
		var store authstore.GlobalStore
		if resolved.Source == authstore.SourceGlobalFile {
			store, _ = authstore.OpenGlobal()
		}
		client := api.New(api.Options{
			ProjectID:   resolvedProject,
			Version:     cmd.Root().Version,
			TokenSource: authstore.NewOAuthTokenSource(resolved, store),
		})
		ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Second)
		defer cancel()
		projects, err := client.ListProjects(ctx)
		if err != nil {
			rows = append(rows, []any{"validation", fmt.Sprintf("FAILED: %v", err)})
			_ = output.Table([]string{"field", "value"}, rows)
			return err
		}
		rows = append(rows, []any{"validation", fmt.Sprintf("OK (%d project access)", len(projects))})
	}

	return output.Table([]string{"field", "value"}, rows)
}

// projectSource explains where the resolved project_id came from so the
// user can debug "why am I hitting the wrong project?".
func projectSource(cmd *cobra.Command, resolved *authstore.Resolved) string {
	if v := cliutil.ProjectIDFlag(cmd); v != "" {
		return "--project-id flag"
	}
	if v := os.Getenv("REVCAT_PROJECT_ID"); v != "" {
		return "REVCAT_PROJECT_ID env"
	}
	if resolved != nil && resolved.ProjectID != "" {
		switch resolved.Source {
		case authstore.SourceLocal:
			return "local: " + resolved.Path
		case authstore.SourceEnv:
			return "REVCAT_PROJECT_ID env (via cred resolution)"
		}
		return string(resolved.Source)
	}
	if cfg, err := project.LoadFromCwd(); err == nil && cfg.ProjectID != "" {
		return cfg.Path
	}
	return "-"
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
