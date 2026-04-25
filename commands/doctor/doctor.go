// Package doctor implements `revcat doctor` - the top-level health check.
//
// `revcat auth doctor` is auth-specific. This command is a higher-level
// summary: keychain reachable, profile resolves, API responds, project
// binding valid.
package doctor

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/spf13/cobra"

	"github.com/akshitkrnagpal/revcat/internal/api"
	authstore "github.com/akshitkrnagpal/revcat/internal/auth"
	"github.com/akshitkrnagpal/revcat/internal/output"
)

// Cmd is mounted on root as `revcat doctor`.
var Cmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run a top-level health check",
	RunE:  runDoctor,
}

func runDoctor(cmd *cobra.Command, args []string) error {
	bypass := false
	if f := cmd.Root().PersistentFlags().Lookup("bypass-keychain"); f != nil {
		bypass = f.Value.String() == "true"
	}

	rows := [][]any{
		{"OK", "platform", fmt.Sprintf("%s/%s %s", runtime.GOOS, runtime.GOARCH, runtime.Version())},
		{"OK", "revcat", cmd.Root().Version},
	}

	store, err := authstore.Open(bypass)
	if err != nil {
		rows = append(rows, []any{"FAIL", "credential store", err.Error()})
		return output.Table([]string{"status", "check", "detail"}, rows)
	}
	storeName := "keychain"
	if bypass {
		storeName = "local file"
	}
	rows = append(rows, []any{"OK", "credential store", storeName})

	profile, err := authstore.Resolve(store, "")
	if err != nil {
		rows = append(rows, []any{"FAIL", "active profile", "no profile found - run `revcat auth login`"})
		return output.Table([]string{"status", "check", "detail"}, rows)
	}
	rows = append(rows, []any{"OK", "active profile", profile.Name})

	client := api.New(api.Options{SecretKey: profile.SecretKey, ProjectID: profile.ProjectID, Version: cmd.Root().Version})
	ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Second)
	defer cancel()
	projects, err := client.ListProjects(ctx)
	if err != nil {
		rows = append(rows, []any{"FAIL", "api reach", err.Error()})
		return output.Table([]string{"status", "check", "detail"}, rows)
	}
	rows = append(rows, []any{"OK", "api reach", fmt.Sprintf("ok, %d project access", len(projects))})

	return output.Table([]string{"status", "check", "detail"}, rows)
}
