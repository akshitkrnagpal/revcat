// Package projects holds `revcat projects ...`.
package projects

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/akshitkrnagpal/revcat/internal/cliutil"
	"github.com/akshitkrnagpal/revcat/internal/output"
)

var Cmd = &cobra.Command{
	Use:     "projects",
	Aliases: []string{"proj"},
	Short:   "Inspect RevenueCat projects",
	Long: `A project is RevenueCat's top-level container - one per app or app
family. revcat is bound to a single project per profile (the one the
secret key has access to). Use these commands to inspect the project
and switch between profiles for different ones.

Project create, app CRUD, audit logs, and collaborators all require a
higher key tier than the per-project secret key revcat uses, so they
are not exposed here. Manage those in the dashboard.`,
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(viewCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List projects accessible to the active secret key",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		projects, err := client.ListProjects(ctx)
		if err != nil {
			return err
		}
		rows := make([][]any, 0, len(projects))
		for _, p := range projects {
			rows = append(rows, []any{p.ID, p.Name, formatTime(p.CreatedAt)})
		}
		return output.Table([]string{"id", "name", "created"}, rows)
	},
}

var viewCmd = &cobra.Command{
	Use:   "view [id]",
	Short: "Show one project by id (defaults to the active profile's project)",
	Args:  cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, prof, err := cliutil.Client(cmd)
		if err != nil {
			return err
		}
		id := prof.ProjectID
		if len(args) == 1 {
			id = args[0]
		}
		if id == "" {
			return errProjectIDRequired
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		p, err := client.GetProject(ctx, id)
		if err != nil {
			return err
		}
		if output.IsJSON() {
			return output.JSON(p)
		}
		rows := [][]any{
			{"id", p.ID},
			{"name", p.Name},
			{"created", formatTime(p.CreatedAt)},
		}
		return output.Table([]string{"field", "value"}, rows)
	},
}

func formatTime(unix int64) string {
	if unix == 0 {
		return "-"
	}
	t := time.Unix(unix, 0).UTC()
	if unix > 9999999999 {
		t = time.UnixMilli(unix).UTC()
	}
	return t.Format("2006-01-02")
}

type sentinelErr string

func (e sentinelErr) Error() string { return string(e) }

const errProjectIDRequired = sentinelErr("no project id given and active profile has none; pass <id>")
