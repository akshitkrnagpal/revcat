// Package cliutil holds helpers shared by command implementations that
// would otherwise be copy-pasted across every leaf command file.
package cliutil

import (
	"errors"
	"fmt"
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

// ResolveCreds resolves the active credential via the precedence chain
// (REVCAT_REFRESH_TOKEN env > walked-up .revcat/config.json > global
// store / active profile). Wraps authstore.Resolve with the cobra
// flags pulled out of the command.
func ResolveCreds(cmd *cobra.Command) (*authstore.Resolved, error) {
	return authstore.Resolve(authstore.ResolveOptions{
		Bypass:      BypassKeychain(cmd),
		FlagProfile: Profile(cmd),
	})
}

// ResolveProjectID returns the active RevenueCat project id with this
// precedence:
//
//  1. --project-id flag
//  2. REVCAT_PROJECT_ID env
//  3. resolved credential's bound project_id (set by the env hatch or
//     by walked-up .revcat/config.json)
//  4. revcat.toml walking up from cwd (committed half)
//
// Returns empty string if none produce a value.
func ResolveProjectID(cmd *cobra.Command, resolved *authstore.Resolved) string {
	if v := ProjectIDFlag(cmd); v != "" {
		return v
	}
	if v := os.Getenv(envProjectID); v != "" {
		return v
	}
	if resolved != nil && resolved.ProjectID != "" {
		return resolved.ProjectID
	}
	if cfg, err := project.LoadFromCwd(); err == nil && cfg.ProjectID != "" {
		return cfg.ProjectID
	}
	return ""
}

// ErrNoProjectID is returned by RequireProjectID when no project id
// could be resolved.
var ErrNoProjectID = errors.New("no project id - pass --project-id, set REVCAT_PROJECT_ID, or run `revcat init` in your repo")

// RequireProjectID is ResolveProjectID + a hint-friendly error when
// nothing is configured.
func RequireProjectID(cmd *cobra.Command, resolved *authstore.Resolved) (string, error) {
	if id := ResolveProjectID(cmd, resolved); id != "" {
		return id, nil
	}
	return "", ErrNoProjectID
}

// Client opens credentials, resolves the project id, and returns a
// ready-to-use API client backed by a refreshing TokenSource.
func Client(cmd *cobra.Command) (*api.Client, *authstore.Resolved, error) {
	return ClientForProject(cmd, "")
}

// ClientForProject is like Client but lets the caller force a specific
// project_id (overriding flag/env/file). Used by `revcat init` after
// the user picks a project but hasn't written .revcat/config.json yet.
func ClientForProject(cmd *cobra.Command, projectIDOverride string) (*api.Client, *authstore.Resolved, error) {
	resolved, err := ResolveCreds(cmd)
	if err != nil {
		return nil, nil, err
	}

	pid := projectIDOverride
	if pid == "" {
		pid = ResolveProjectID(cmd, resolved)
	}

	// For SourceKeychain / SourceGlobalFile the OAuthTokenSource
	// needs a store handle to persist refreshed tokens back. Open
	// the matching store; cheap.
	var store authstore.GlobalStore
	switch resolved.Source {
	case authstore.SourceKeychain, authstore.SourceGlobalFile:
		store, err = authstore.OpenGlobal(BypassKeychain(cmd))
		if err != nil {
			return nil, nil, fmt.Errorf("reopen global store for refresh: %w", err)
		}
	}

	opts := api.Options{
		ProjectID:   pid,
		Version:     cmd.Root().Version,
		TokenSource: authstore.NewOAuthTokenSource(resolved, store),
	}
	return api.New(opts), resolved, nil
}
