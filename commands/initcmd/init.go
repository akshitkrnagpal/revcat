// Package initcmd implements `revcat init` - the per-repo project
// context bootstrap.
//
// Writes two files at cwd:
//
//   - revcat.toml (committed): project_id + optional apps. Useful for
//     human readers and so a fresh clone shows which RC project this
//     repo belongs to.
//
//   - .revcat/config.json (gitignored, mode 0600): credential half plus
//     project_id and apps. Walked up from cwd by every revcat command
//     so an agent inside the directory can run without touching the
//     user's keychain.
//
// Also appends ".revcat/" to .gitignore (idempotent), creating it if
// missing.
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
	authstore "github.com/akshitkrnagpal/revcat/internal/auth"
	"github.com/akshitkrnagpal/revcat/internal/cliutil"
	"github.com/akshitkrnagpal/revcat/internal/output"
	"github.com/akshitkrnagpal/revcat/internal/project"
)

var (
	flagAppIDs       []string
	flagForce        bool
	flagPath         string
	flagNoApps       bool
	flagNoLocalCreds bool
)

// Cmd is the cobra command exported to the root.
var Cmd = &cobra.Command{
	Use:   "init",
	Short: "Bootstrap project context (revcat.toml + .revcat/config.json)",
	Long: `Bind the current directory to a RevenueCat project. Writes:

  - revcat.toml    (committed): project_id + optional apps
  - .revcat/config.json (gitignored, mode 0600): credentials + project_id

After init, every command run inside this directory inherits the project
context. Agents and sandboxes that have access to the directory can run
revcat without touching the user's keychain.

Interactive (default): lists projects you can access, prompts for one,
then optionally lists apps in that project and lets you tag them.

Scripted: pass --project-id (and optional --app-id, repeated). Skip the
apps block entirely with --no-apps. Skip the local creds copy with
--no-local-creds (writes only revcat.toml).`,
	RunE: runInit,
}

func init() {
	Cmd.Flags().StringSliceVar(&flagAppIDs, "app-id", nil, "App ids to record (repeatable)")
	Cmd.Flags().BoolVar(&flagForce, "force", false, "Overwrite an existing revcat.toml or .revcat/config.json")
	Cmd.Flags().StringVar(&flagPath, "path", "", "Where to write files (default: cwd)")
	Cmd.Flags().BoolVar(&flagNoApps, "no-apps", false, "Skip the apps section entirely")
	Cmd.Flags().BoolVar(&flagNoLocalCreds, "no-local-creds", false, "Don't write .revcat/config.json (only revcat.toml)")
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
	tomlTarget := filepath.Join(dir, project.FileName)
	credsTarget := filepath.Join(dir, authstore.LocalConfigPath)

	if _, err := os.Stat(tomlTarget); err == nil && !flagForce {
		return fmt.Errorf("%s already exists; pass --force to overwrite", tomlTarget)
	}
	if !flagNoLocalCreds {
		if _, err := os.Stat(credsTarget); err == nil && !flagForce {
			return fmt.Errorf("%s already exists; pass --force to overwrite", credsTarget)
		}
	}

	resolved, err := cliutil.ResolveCreds(cmd)
	if err != nil {
		return fmt.Errorf("init needs an authenticated profile: %w", err)
	}
	if resolved.Source == authstore.SourceLocal {
		return errors.New("a .revcat/config.json already exists in this tree; pass --force to reinit")
	}

	// Build a client with whatever creds we have, no project bound yet.
	client, _, err := cliutil.Client(cmd)
	if err != nil {
		return err
	}

	projectID, err := pickProjectID(cmd.Context(), cmd, client)
	if err != nil {
		return err
	}

	tomlCfg := &project.Config{ProjectID: projectID}
	var apps []project.App
	if !flagNoApps {
		apps, err = pickApps(cmd.Context(), cmd, projectID)
		if err != nil {
			return err
		}
		tomlCfg.Apps = apps
	}

	if err := project.Save(tomlTarget, tomlCfg); err != nil {
		return fmt.Errorf("write %s: %w", tomlTarget, err)
	}
	output.Success("wrote %s", tomlTarget)

	if !flagNoLocalCreds {
		localCfg := &authstore.LocalConfig{
			ProjectID: projectID,
			Profile:   *resolved.Profile,
			Apps:      toLocalApps(apps),
		}
		if err := authstore.SaveLocal(credsTarget, localCfg); err != nil {
			return fmt.Errorf("write %s: %w", credsTarget, err)
		}
		output.Success("wrote %s (mode 0600)", credsTarget)

		added, err := authstore.EnsureGitignored(dir)
		if err != nil {
			output.Warn("could not update .gitignore: %v", err)
		} else if added {
			output.Info("appended `.revcat/` to .gitignore")
		}
	}

	output.Info("  project_id: %s", projectID)
	if len(apps) > 0 {
		ids := make([]string, len(apps))
		for i, a := range apps {
			ids[i] = a.ID
		}
		output.Info("  apps:       %s", strings.Join(ids, ", "))
	}
	output.Info("commit revcat.toml; do NOT commit .revcat/")
	return nil
}

func toLocalApps(in []project.App) []authstore.LocalApp {
	if len(in) == 0 {
		return nil
	}
	out := make([]authstore.LocalApp, len(in))
	for i, a := range in {
		out[i] = authstore.LocalApp{ID: a.ID, Name: a.Name}
	}
	return out
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
	scoped, _, err := cliutil.ClientForProject(cmd, projectID)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(parent, 10*time.Second)
	defer cancel()
	apps, err := scoped.ListApps(ctx)
	if err != nil {
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
