// Package initcmd implements `revcat init` - the per-repo project
// context bootstrap. Writes a revcat.toml at the cwd that pins the
// project_id so subsequent commands run in the right RC project without
// the user having to remember.
package initcmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"

	"github.com/akshitkrnagpal/revcat/internal/api"
	"github.com/akshitkrnagpal/revcat/internal/cliutil"
	"github.com/akshitkrnagpal/revcat/internal/output"
	"github.com/akshitkrnagpal/revcat/internal/project"
)

var (
	flagAppIDs []string
	flagForce  bool
	flagPath   string
	flagNoApps bool
)

// Cmd is the cobra command exported to the root.
var Cmd = &cobra.Command{
	Use:   "init",
	Short: "Bootstrap revcat.toml in the current directory",
	Long: `Write a revcat.toml at the current directory pinning the project_id
(and optional apps). Once present, every command run from this repo
inherits the binding without --project-id.

Interactive (default): lists projects you can access, prompts for one,
then optionally lists apps in that project and lets you tag them.

Scripted: pass --project-id (and optional --app-id, repeated). Skip the
apps block entirely with --no-apps.`,
	RunE: runInit,
}

func init() {
	Cmd.Flags().StringSliceVar(&flagAppIDs, "app-id", nil, "App ids to record (repeatable)")
	Cmd.Flags().BoolVar(&flagForce, "force", false, "Overwrite an existing revcat.toml")
	Cmd.Flags().StringVar(&flagPath, "path", "", "Where to write revcat.toml (default: cwd)")
	Cmd.Flags().BoolVar(&flagNoApps, "no-apps", false, "Skip the apps section entirely")
}

func runInit(cmd *cobra.Command, args []string) error {
	dir := flagPath
	if dir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		dir = cwd
	}
	target := filepath.Join(dir, project.FileName)

	if _, err := os.Stat(target); err == nil && !flagForce {
		return fmt.Errorf("%s already exists; pass --force to overwrite", target)
	}

	client, _, err := cliutil.Client(cmd)
	if err != nil {
		return err
	}

	projectID, err := pickProjectID(cmd.Context(), cmd, client)
	if err != nil {
		return err
	}

	cfg := &project.Config{ProjectID: projectID}

	if !flagNoApps {
		apps, err := pickApps(cmd.Context(), cmd, projectID)
		if err != nil {
			return err
		}
		cfg.Apps = apps
	}

	if err := project.Save(target, cfg); err != nil {
		return fmt.Errorf("write %s: %w", target, err)
	}

	output.Success("wrote %s", target)
	output.Info("  project_id: %s", cfg.ProjectID)
	if len(cfg.Apps) > 0 {
		ids := make([]string, len(cfg.Apps))
		for i, a := range cfg.Apps {
			ids[i] = a.ID
		}
		output.Info("  apps:       %s", strings.Join(ids, ", "))
	}
	output.Info("commit revcat.toml so collaborators get the same context")
	return nil
}

func pickProjectID(parent context.Context, cmd *cobra.Command, client *api.Client) (string, error) {
	if v := cliutil.ProjectIDFlag(cmd); v != "" {
		return v, nil
	}
	if v := os.Getenv("REVCAT_PROJECT_ID"); v != "" {
		return v, nil
	}
	ctx, cancel := context.WithTimeout(parent, 10*time.Second)
	defer cancel()
	projects, err := client.ListProjects(ctx)
	if err != nil {
		return "", fmt.Errorf("list projects: %w", err)
	}
	if len(projects) == 0 {
		return "", errors.New("no project access on this profile")
	}
	if len(projects) == 1 {
		output.Info("only one accessible project: %s (%s)", projects[0].Name, projects[0].ID)
		return projects[0].ID, nil
	}
	options := make([]string, len(projects))
	for i, p := range projects {
		options[i] = fmt.Sprintf("%s (%s)", p.Name, p.ID)
	}
	var idx int
	if err := survey.AskOne(&survey.Select{
		Message: "Which project?",
		Options: options,
	}, &idx); err != nil {
		return "", err
	}
	return projects[idx].ID, nil
}

func pickApps(parent context.Context, cmd *cobra.Command, projectID string) ([]project.App, error) {
	if len(flagAppIDs) > 0 {
		out := make([]project.App, len(flagAppIDs))
		for i, id := range flagAppIDs {
			out[i] = project.App{ID: id}
		}
		return out, nil
	}
	// Re-bind the API client to the chosen project_id so /projects/{id}/apps
	// hits the right path (the caller's client may have no project bound
	// yet — there's no revcat.toml yet to resolve from).
	scoped, _, err := cliutil.ClientForProject(cmd, projectID)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(parent, 10*time.Second)
	defer cancel()
	apps, err := scoped.ListApps(ctx)
	if err != nil {
		// Don't block init on app listing failure - the file is still
		// useful with project_id alone.
		output.Warn("could not list apps (%v); writing project_id only", err)
		return nil, nil
	}
	if len(apps) == 0 {
		return nil, nil
	}
	options := make([]string, len(apps))
	for i, a := range apps {
		bundle := a.BundleID
		if bundle == "" {
			bundle = a.PackageName
		}
		if bundle != "" {
			options[i] = fmt.Sprintf("%s · %s · %s (%s)", a.Name, a.Type, bundle, a.ID)
		} else {
			options[i] = fmt.Sprintf("%s · %s (%s)", a.Name, a.Type, a.ID)
		}
	}
	defaults := make([]string, 0, len(options))
	for _, o := range options {
		defaults = append(defaults, o)
	}
	var picked []int
	if err := survey.AskOne(&survey.MultiSelect{
		Message: "Which apps to record? (space to toggle, enter to confirm; leave empty to skip)",
		Options: options,
		Default: defaults,
	}, &picked); err != nil {
		return nil, err
	}
	if len(picked) == 0 {
		return nil, nil
	}
	out := make([]project.App, 0, len(picked))
	for _, i := range picked {
		out = append(out, project.App{ID: apps[i].ID, Name: apps[i].Name})
	}
	return out, nil
}
