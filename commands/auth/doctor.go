package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	authstore "github.com/akshitkrnagpal/revcat/internal/auth"
	"github.com/akshitkrnagpal/revcat/internal/cliutil"
	"github.com/akshitkrnagpal/revcat/internal/output"
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
	bypass := bypassKeychain(cmd)

	resolved, err := cliutil.ResolveCreds(cmd)
	if err != nil {
		store := "keychain"
		if bypass {
			store = "file (~/.revcat/config.json)"
		}
		checks = append(checks, check{name: "credential resolve", ok: false, msg: err.Error(), hint: "checked: REVCAT_REFRESH_TOKEN env, walked-up .revcat/config.json, " + store + " (active profile). run `revcat auth login`."})
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
