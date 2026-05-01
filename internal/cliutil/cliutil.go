// Package cliutil holds helpers shared by command implementations that
// would otherwise be copy-pasted across every leaf command file.
package cliutil

import (
	"errors"
	"os"

	"github.com/spf13/cobra"

	"github.com/akshitkrnagpal/revcat/internal/api"
	authstore "github.com/akshitkrnagpal/revcat/internal/auth"
	"github.com/akshitkrnagpal/revcat/internal/project"
)

// envProjectID is the env var that overrides revcat.toml.
const envProjectID = "REVCAT_PROJECT_ID"

// BypassKeychain reads the global --bypass-keychain flag from cobra root.
func BypassKeychain(cmd *cobra.Command) bool {
	flag := cmd.Root().PersistentFlags().Lookup("bypass-keychain")
	if flag == nil {
		return false
	}
	return flag.Value.String() == "true"
}

// Profile reads the global --profile flag from cobra root.
func Profile(cmd *cobra.Command) string {
	flag := cmd.Root().PersistentFlags().Lookup("profile")
	if flag == nil {
		return ""
	}
	return flag.Value.String()
}

// ProjectIDFlag reads the global --project-id flag from cobra root.
func ProjectIDFlag(cmd *cobra.Command) string {
	flag := cmd.Root().PersistentFlags().Lookup("project-id")
	if flag == nil {
		return ""
	}
	return flag.Value.String()
}

// ResolveProjectID returns the active RevenueCat project id with this
// precedence:
//
//  1. --project-id flag
//  2. REVCAT_PROJECT_ID env
//  3. revcat.toml walking up from cwd (Terraform-style per-repo context)
//  4. profile.ProjectID (legacy: secret-key profiles bind one at login)
//
// Returns empty string if none of those produce a value. Callers that
// need a project id should error with a hint pointing at `revcat init`.
func ResolveProjectID(cmd *cobra.Command, profile *authstore.Profile) string {
	if v := ProjectIDFlag(cmd); v != "" {
		return v
	}
	if v := os.Getenv(envProjectID); v != "" {
		return v
	}
	if cfg, err := project.LoadFromCwd(); err == nil && cfg.ProjectID != "" {
		return cfg.ProjectID
	}
	if profile != nil {
		return profile.ProjectID
	}
	return ""
}

// ErrNoProjectID is returned by RequireProjectID when no project id
// could be resolved.
var ErrNoProjectID = errors.New("no project id - pass --project-id, set REVCAT_PROJECT_ID, run `revcat init` in your repo, or attach one to the profile")

// RequireProjectID is ResolveProjectID + a hint-friendly error when
// nothing is configured.
func RequireProjectID(cmd *cobra.Command, profile *authstore.Profile) (string, error) {
	if id := ResolveProjectID(cmd, profile); id != "" {
		return id, nil
	}
	return "", ErrNoProjectID
}

// Client opens the credential store, resolves the active profile, and
// returns a ready-to-use API client. Returns an error if the profile is
// missing or the store can't be opened. Most commands use this.
//
// Project id is resolved via ResolveProjectID, so per-repo revcat.toml
// and the --project-id flag both apply.
//
// For OAuth profiles, the returned client carries a refreshing
// TokenSource that updates the stored profile on each refresh.
func Client(cmd *cobra.Command) (*api.Client, *authstore.Profile, error) {
	return ClientForProject(cmd, "")
}

// ClientForProject is like Client but lets the caller force a specific
// project_id (overriding flag/env/toml). Used by `revcat init` after the
// user picks a project but hasn't written revcat.toml yet.
func ClientForProject(cmd *cobra.Command, projectIDOverride string) (*api.Client, *authstore.Profile, error) {
	store, err := authstore.Open(BypassKeychain(cmd))
	if err != nil {
		return nil, nil, err
	}
	prof, err := authstore.Resolve(store, Profile(cmd))
	if err != nil {
		return nil, nil, err
	}
	pid := projectIDOverride
	if pid == "" {
		pid = ResolveProjectID(cmd, prof)
	}
	opts := api.Options{
		ProjectID: pid,
		Version:   cmd.Root().Version,
	}
	if prof.EffectiveAuthType() == authstore.AuthTypeOAuth {
		opts.TokenSource = authstore.NewOAuthTokenSource(store, prof)
	} else {
		opts.SecretKey = prof.SecretKey
	}
	return api.New(opts), prof, nil
}
