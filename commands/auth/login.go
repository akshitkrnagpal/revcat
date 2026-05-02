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

var (
	loginName     string
	loginClientID string
	loginScopes   []string
	loginNoVerify bool
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate revcat against RevenueCat via OAuth",
	Long: `Run the browser OAuth flow and store the resulting tokens. The OAuth
client is RevenueCat's public PKCE client by default; override with
REVCAT_OAUTH_CLIENT_ID or --client-id.

Tokens are written to your OS keychain. Pass --bypass-keychain (or set
REVCAT_BYPASS_KEYCHAIN=1) to use ~/.revcat/config.json (HOME) instead,
useful on Linux without secret-service or in containers.

After login, run ` + "`revcat init`" + ` inside your project directory to bind
a project_id (or set REVCAT_PROJECT_ID per command).

Examples:
  revcat auth login                          # browser, default profile
  revcat auth login --name work              # browser, named profile
  revcat auth login --client-id <id>         # custom OAuth client_id
`,
	RunE: runLogin,
}

func init() {
	loginCmd.Flags().StringVarP(&loginName, "name", "n", "", "Profile name (default: 'default')")
	loginCmd.Flags().StringVar(&loginClientID, "client-id", "", "OAuth client_id (defaults to REVCAT_OAUTH_CLIENT_ID env or the embedded public client)")
	loginCmd.Flags().StringSliceVar(&loginScopes, "scope", nil, "OAuth scopes (default: revcat's full read/write set)")
	loginCmd.Flags().BoolVar(&loginNoVerify, "no-verify", false, "Skip the API check that the credentials work after login")
}

func runLogin(cmd *cobra.Command, args []string) error {
	if loginName == "" {
		loginName = "default"
	}
	clientID := loginClientID
	if clientID == "" {
		clientID = authstore.OAuthClientID()
	}
	if clientID == "" {
		return errors.New("no OAuth client_id configured. set REVCAT_OAUTH_CLIENT_ID, pass --client-id, or build with -ldflags '-X .../auth.EmbeddedClientID=<id>'")
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
	server.ProfileName = loginName
	defer server.Close()

	state, err := api.RandomState()
	if err != nil {
		return err
	}

	authURL := api.AuthorizeURL(clientID, server.URL, scopes, state, pkce.Challenge)

	output.Info("opening browser to authorize revcat")
	output.Info("if it doesn't open automatically, paste this URL:\n  %s", authURL)
	if err := api.OpenBrowser(authURL); err != nil {
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

	store, err := authstore.OpenGlobal(bypassKeychain(cmd))
	if err != nil {
		return err
	}

	expiresAt := int64(0)
	if tok.ExpiresIn > 0 {
		expiresAt = time.Now().Add(time.Duration(tok.ExpiresIn) * time.Second).UnixMilli()
	}

	profile := &authstore.Profile{
		Name:         loginName,
		AccessToken:  tok.AccessToken,
		RefreshToken: tok.RefreshToken,
		ExpiresAt:    expiresAt,
		Scope:        tok.Scope,
		ClientID:     clientID,
	}

	if !loginNoVerify {
		resolved := &authstore.Resolved{Profile: profile, Source: authstore.SourceKeychain}
		client := api.New(api.Options{
			TokenSource: authstore.NewOAuthTokenSource(resolved, store),
			Version:     cmd.Root().Version,
		})
		vctx, vcancel := context.WithTimeout(cmd.Context(), 10*time.Second)
		defer vcancel()
		if _, err := client.ListProjects(vctx); err != nil {
			return fmt.Errorf("verify access: %w", err)
		}
	}

	if err := store.Set(profile); err != nil {
		return err
	}

	loc := "OS keychain"
	if bypassKeychain(cmd) {
		loc = "~/.revcat/config.json"
	}
	output.Success("oauth login saved as %q in %s", loginName, loc)
	output.Info("  expires:    %s", time.UnixMilli(profile.ExpiresAt).Local().Format("2006-01-02 15:04 MST"))
	output.Info("  scope:      %s", profile.Scope)
	output.Info("next: cd into your repo and run `revcat init` to bind a project_id")
	return nil
}

func redactKey(k string) string {
	if len(k) < 8 {
		return "***"
	}
	return k[:4] + "..." + k[len(k)-4:]
}
