package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/akshitkrnagpal/revcat/internal/api"
	authstore "github.com/akshitkrnagpal/revcat/internal/auth"
	"github.com/akshitkrnagpal/revcat/internal/output"
)

// runOAuthLogin drives the PKCE flow:
//  1. resolve the client_id (flag > env > embedded)
//  2. start a local callback server, generate PKCE
//  3. open the browser to RC's authorize URL
//  4. wait for the redirect, exchange the code, store tokens
func runOAuthLogin(cmd *cobra.Command) error {
	clientID := loginClientID
	if clientID == "" {
		clientID = authstore.OAuthClientID()
	}
	if clientID == "" {
		return errors.New("no OAuth client_id configured. set REVCAT_OAUTH_CLIENT_ID or pass --client-id. RevenueCat must register revcat as a public OAuth client first; see https://www.revenuecat.com/docs/projects/oauth-setup")
	}

	scopes := loginScopes
	if len(scopes) == 0 {
		scopes = api.DefaultScopes
	}

	pkce, err := api.NewPKCE()
	if err != nil {
		return err
	}

	server, err := api.NewLoopbackServer()
	if err != nil {
		return err
	}
	defer server.Close()

	state, err := randomStateOrFatal()
	if err != nil {
		return err
	}

	authURL := api.AuthorizeURL(clientID, server.URL, scopes, state, pkce.Challenge)

	output.Info("opening browser to authorize revcat")
	output.Info("if it doesn't open automatically, paste this URL:\n  %s", authURL)
	if err := api.OpenBrowser(authURL); err != nil {
		// non-fatal: user can copy/paste manually.
		output.Warn("could not auto-open browser: %v", err)
	}

	ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Minute)
	defer cancel()
	resp, err := server.Wait(ctx)
	if err != nil {
		return fmt.Errorf("waiting for callback: %w", err)
	}
	if resp.Err != "" {
		return fmt.Errorf("authorization rejected (%s): %s", resp.Err, resp.ErrDesc)
	}
	if resp.State != state {
		return errors.New("state mismatch on callback - possible CSRF; refusing to continue")
	}

	tok, err := api.ExchangeCode(ctx, clientID, "", resp.Code, server.URL, pkce.Verifier)
	if err != nil {
		return fmt.Errorf("token exchange: %w", err)
	}

	store, err := authstore.Open(bypassKeychain(cmd))
	if err != nil {
		return err
	}

	expiresAt := int64(0)
	if tok.ExpiresIn > 0 {
		expiresAt = time.Now().Add(time.Duration(tok.ExpiresIn) * time.Second).UnixMilli()
	}

	profile := &authstore.Profile{
		Name:         loginName,
		AuthType:     authstore.AuthTypeOAuth,
		ProjectID:    loginProjectID,
		AccessToken:  tok.AccessToken,
		RefreshToken: tok.RefreshToken,
		ExpiresAt:    expiresAt,
		Scope:        tok.Scope,
		ClientID:     clientID,
	}

	if !loginNoVerify {
		client := api.New(api.Options{
			TokenSource: authstore.NewOAuthTokenSource(store, profile),
			ProjectID:   profile.ProjectID,
			Version:     cmd.Root().Version,
		})
		vctx, vcancel := context.WithTimeout(cmd.Context(), 10*time.Second)
		defer vcancel()
		projects, err := client.ListProjects(vctx)
		if err != nil {
			return fmt.Errorf("verify access: %w", err)
		}
		if profile.ProjectID == "" && len(projects) > 0 {
			id, err := pickProject(projects)
			if err != nil {
				return err
			}
			profile.ProjectID = id
		}
	}

	if err := store.Set(profile); err != nil {
		return err
	}

	loc := "OS keychain"
	if bypassKeychain(cmd) {
		loc = "./.revcat/config.json"
	}
	output.Success("oauth login saved as %q in %s", loginName, loc)
	if profile.ProjectID != "" {
		output.Info("  project_id: %s", profile.ProjectID)
	}
	output.Info("  expires:    %s", time.UnixMilli(profile.ExpiresAt).Local().Format("2006-01-02 15:04 MST"))
	output.Info("  scope:      %s", profile.Scope)
	output.Info("  use it: revcat --profile %s ...", loginName)
	return nil
}

// randomStateOrFatal wraps the api.randomState helper - the real impl
// lives in the api package so api/oauth_test can exercise it.
func randomStateOrFatal() (string, error) {
	// borrow the same generator the api package uses for PKCE.
	p, err := api.NewPKCE()
	if err != nil {
		return "", err
	}
	return p.Verifier, nil
}
