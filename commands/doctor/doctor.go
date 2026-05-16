// Package doctor implements `revcat doctor` - the top-level health check.
//
// `revcat auth doctor` is auth-specific. This command is a higher-level
// summary: profile resolves, API responds, project binding valid.
package doctor

import (
	"context"
	"errors"
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

var errDoctorFailed = errors.New("one or more checks failed")

func runDoctor(cmd *cobra.Command, args []string) error {
	rows := [][]any{
		{"OK", "platform", fmt.Sprintf("%s/%s %s", runtime.GOOS, runtime.GOARCH, runtime.Version())},
		{"OK", "revcat", cmd.Root().Version},
	}

	client, resolved, err := cliutil.Client(cmd)
	if err != nil {
		rows = append(rows, []any{"FAIL", "credential resolve", err.Error()})
		return renderRows(rows, true)
	}
	rows = append(rows, []any{"OK", "active credential", fmt.Sprintf("%s (source=%s)", resolved.Profile.Name, resolved.Source)})

	ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Second)
	defer cancel()
	projects, err := client.ListProjects(ctx)
	if err != nil {
		rows = append(rows, []any{"FAIL", "api reach", err.Error()})
		return renderRows(rows, true)
	}
	rows = append(rows, []any{"OK", "api reach", fmt.Sprintf("ok, %d project access", len(projects))})

	if pid := cliutil.ResolveProjectID(cmd, resolved); pid != "" {
		rows = append(rows, []any{"OK", "project context", pid})
	} else {
		rows = append(rows, []any{"WARN", "project context", "none - run `revcat init` or pass --project-id"})
	}

	return renderRows(rows, false)
}

func renderRows(rows [][]any, failed bool) error {
	if err := output.Table([]string{"status", "check", "detail"}, rows); err != nil {
		return err
	}
	if failed {
		return errDoctorFailed
	}
	return nil
}
