package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	authstore "github.com/akshitkrnagpal/revcat/internal/auth"
	"github.com/akshitkrnagpal/revcat/internal/cliutil"
	"github.com/akshitkrnagpal/revcat/internal/output"
	"github.com/akshitkrnagpal/revcat/internal/project"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Self-diagnose auth setup",
	Long: `Walk through the most common auth misconfigurations and report what's
working, what isn't, and how to fix it.`,
	RunE: runAuthDoctor,
}

type check struct {
	name string
	ok   bool
	msg  string
	hint string
}

func runAuthDoctor(cmd *cobra.Command, args []string) error {
	checks := []check{}

	resolved, err := cliutil.ResolveCreds(cmd)
	if err != nil {
		checks = append(checks, check{name: "credential resolve", ok: false, msg: err.Error(), hint: "checked: REVCAT_REFRESH_TOKEN env, walked-up .revcat/config.json, ~/.revcat/config.json (active profile). run `revcat auth login`."})
		return renderChecks(checks)
	}
	checks = append(checks, check{name: "credential resolve", ok: true, msg: fmt.Sprintf("source=%s profile=%s", resolved.Source, resolved.Profile.Name)})
	if resolved.Path != "" {
		checks = append(checks, check{name: "credential path", ok: true, msg: resolved.Path})
	}

	client, _, err := cliutil.Client(cmd)
	if err != nil {
		checks = append(checks, check{name: "client build", ok: false, msg: err.Error()})
		return renderChecks(checks)
	}

	ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Second)
	defer cancel()
	projects, apiErr := client.ListProjects(ctx)
	if apiErr != nil {
		checks = append(checks, check{name: "API reach", ok: false, msg: apiErr.Error(), hint: "is the network up? check VPN/proxy. set REVCAT_DEBUG=api for full request log"})
		return renderChecks(checks)
	}
	checks = append(checks, check{name: "API reach", ok: true, msg: fmt.Sprintf("ok, %d project access", len(projects))})

	resolvedProject := cliutil.ResolveProjectID(cmd, resolved)
	if resolvedProject == "" {
		checks = append(checks, check{name: "project context", ok: false, msg: "no project_id resolved", hint: "run `revcat init` in your repo, pass --project-id, or set REVCAT_PROJECT_ID"})
	} else {
		found := false
		for _, p := range projects {
			if p.ID == resolvedProject {
				found = true
				break
			}
		}
		if !found {
			checks = append(checks, check{name: "project context", ok: false, msg: resolvedProject + " not accessible to this credential", hint: "the project may have been moved or this credential lacks access"})
		} else {
			checks = append(checks, check{name: "project context", ok: true, msg: resolvedProject})
		}
	}

	// Mismatch detector: revcat.toml is committed and meant to declare
	// which project the repo belongs to; .revcat/config.json (loaded
	// via SourceLocal) is the gitignored credential half. They should
	// agree. Disagreement means someone init'd a different project, or
	// edited the toml without rerunning init.
	if resolved.Source == authstore.SourceLocal {
		if cfg, err := project.LoadFromCwd(); err == nil && cfg.ProjectID != "" {
			if cfg.ProjectID != resolved.ProjectID {
				checks = append(checks, check{
					name: "toml/local mismatch",
					ok:   false,
					msg:  fmt.Sprintf("revcat.toml says %s, .revcat/config.json says %s", cfg.ProjectID, resolved.ProjectID),
					hint: "rerun `revcat init --force` to realign, or edit revcat.toml to match",
				})
			}
		}
	}

	return renderChecks(checks)
}

func renderChecks(checks []check) error {
	rows := make([][]any, 0, len(checks))
	allOK := true
	for _, c := range checks {
		status := "OK"
		if !c.ok {
			status = "FAIL"
			allOK = false
		}
		msg := c.msg
		if c.hint != "" {
			msg += "\n  hint: " + c.hint
		}
		rows = append(rows, []any{status, c.name, msg})
	}
	if err := output.Table([]string{"status", "check", "detail"}, rows); err != nil {
		return err
	}
	if !allOK {
		_ = authstore.GetActive
		return fmt.Errorf("one or more checks failed")
	}
	return nil
}
