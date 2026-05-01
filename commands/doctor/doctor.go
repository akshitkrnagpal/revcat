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

	"github.com/akshitkrnagpal/revcat/internal/cliutil"
	"github.com/akshitkrnagpal/revcat/internal/output"
)

// Cmd is mounted on root as `revcat doctor`.
var Cmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run a top-level health check",
	RunE:  runDoctor,
}

func runDoctor(cmd *cobra.Command, args []string) error {
	rows := [][]any{
		{"OK", "platform", fmt.Sprintf("%s/%s %s", runtime.GOOS, runtime.GOARCH, runtime.Version())},
		{"OK", "revcat", cmd.Root().Version},
	}

	storeName := "keychain"
	if cliutil.BypassKeychain(cmd) {
		storeName = "local file"
	}

	client, profile, err := cliutil.Client(cmd)
	if err != nil {
		rows = append(rows, []any{"FAIL", "credential store / active profile", err.Error()})
		return output.Table([]string{"status", "check", "detail"}, rows)
	}
	rows = append(rows, []any{"OK", "credential store", storeName})
	rows = append(rows, []any{"OK", "active profile", fmt.Sprintf("%s (%s)", profile.Name, profile.EffectiveAuthType())})

	ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Second)
	defer cancel()
	projects, err := client.ListProjects(ctx)
	if err != nil {
		rows = append(rows, []any{"FAIL", "api reach", err.Error()})
		return output.Table([]string{"status", "check", "detail"}, rows)
	}
	rows = append(rows, []any{"OK", "api reach", fmt.Sprintf("ok, %d project access", len(projects))})

	if pid := cliutil.ResolveProjectID(cmd, profile); pid != "" {
		rows = append(rows, []any{"OK", "project context", pid})
	} else {
		rows = append(rows, []any{"WARN", "project context", "none - run `revcat init` or pass --project-id"})
	}

	return output.Table([]string{"status", "check", "detail"}, rows)
}
