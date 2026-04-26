// Package cliutil holds helpers shared by command implementations that
// would otherwise be copy-pasted across every leaf command file.
package cliutil

import (
	"github.com/spf13/cobra"

	"github.com/akshitkrnagpal/revcat/internal/api"
	authstore "github.com/akshitkrnagpal/revcat/internal/auth"
)

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

// Client opens the credential store, resolves the active profile, and
// returns a ready-to-use API client. Returns an error if the profile is
// missing or the store can't be opened. Most commands use this.
//
// For OAuth profiles, the returned client carries a refreshing
// TokenSource that updates the stored profile on each refresh.
func Client(cmd *cobra.Command) (*api.Client, *authstore.Profile, error) {
	store, err := authstore.Open(BypassKeychain(cmd))
	if err != nil {
		return nil, nil, err
	}
	prof, err := authstore.Resolve(store, Profile(cmd))
	if err != nil {
		return nil, nil, err
	}
	opts := api.Options{
		ProjectID: prof.ProjectID,
		Version:   cmd.Root().Version,
	}
	if prof.EffectiveAuthType() == authstore.AuthTypeOAuth {
		opts.TokenSource = authstore.NewOAuthTokenSource(store, prof)
	} else {
		opts.SecretKey = prof.SecretKey
	}
	return api.New(opts), prof, nil
}
