package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/akshitkrnagpal/revcat/internal/api"
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

	store, err := authstore.Open(bypass)
	if err != nil {
		checks = append(checks, check{name: "credential store", ok: false, msg: err.Error(), hint: "if no keychain is available, retry with --bypass-keychain"})
		return renderChecks(checks)
	}
	storeName := "keychain"
	if bypass {
		storeName = "local file"
	}
	checks = append(checks, check{name: "credential store", ok: true, msg: storeName + " accessible"})

	profile, err := authstore.Resolve(store, cliutil.Profile(cmd))
	if err != nil {
		checks = append(checks, check{name: "active profile", ok: false, msg: err.Error(), hint: "run `revcat auth login --name default --secret-key sk_...`"})
		return renderChecks(checks)
	}
	checks = append(checks, check{name: "active profile", ok: true, msg: profile.Name})

	authType := profile.EffectiveAuthType()
	checks = append(checks, check{name: "auth type", ok: true, msg: string(authType)})

	opts := api.Options{ProjectID: profile.ProjectID, Version: cmd.Root().Version}
	switch authType {
	case authstore.AuthTypeOAuth:
		if profile.AccessToken == "" {
			checks = append(checks, check{name: "oauth token", ok: false, msg: "no access_token on profile", hint: "rerun `revcat auth login --oauth`"})
		} else {
			checks = append(checks, check{name: "oauth token", ok: true, msg: "present"})
		}
		opts.TokenSource = authstore.NewOAuthTokenSource(store, profile)
	default:
		if !strings.HasPrefix(profile.SecretKey, "sk_") {
			checks = append(checks, check{name: "key format", ok: false, msg: "stored key does not start with sk_", hint: "v2 secret keys begin with sk_; double-check it isn't a public SDK key"})
		} else {
			checks = append(checks, check{name: "key format", ok: true, msg: "looks like a v2 secret key"})
		}
		opts.SecretKey = profile.SecretKey
	}

	client := api.New(opts)
	ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Second)
	defer cancel()
	projects, apiErr := client.ListProjects(ctx)
	if apiErr != nil {
		checks = append(checks, check{name: "API reach", ok: false, msg: apiErr.Error(), hint: "is the network up? check VPN/proxy. set REVCAT_DEBUG=api for full request log"})
	} else {
		checks = append(checks, check{name: "API reach", ok: true, msg: fmt.Sprintf("ok, %d project access", len(projects))})

		if profile.ProjectID == "" {
			checks = append(checks, check{name: "project binding", ok: false, msg: "no project_id stored on profile", hint: "rerun `revcat auth login` and pick a project"})
		} else {
			found := false
			for _, p := range projects {
				if p.ID == profile.ProjectID {
					found = true
					break
				}
			}
			if !found {
				checks = append(checks, check{name: "project binding", ok: false, msg: profile.ProjectID + " not in this key's project list", hint: "the key may have been rotated to a different project; rerun `revcat auth login`"})
			} else {
				checks = append(checks, check{name: "project binding", ok: true, msg: profile.ProjectID})
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
		return fmt.Errorf("one or more checks failed")
	}
	return nil
}
