package auth

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/akshitkrnagpal/revcat/internal/api"
)

// envClientID lets the user override the baked-in OAuth client_id.
const envClientID = "REVCAT_OAUTH_CLIENT_ID"

// EmbeddedClientID is the OAuth client_id baked into release builds.
// Defaults to revcat's public PKCE client registered with RevenueCat.
//
// Override at build time via:
//
//	go build -ldflags "-X github.com/akshitkrnagpal/revcat/internal/auth.EmbeddedClientID=<id>"
//
// Or at runtime via the REVCAT_OAUTH_CLIENT_ID env var.
var EmbeddedClientID = "UmV2Q2F0"

// OAuthClientID returns the active client_id, env override taking
// precedence over the embedded constant.
func OAuthClientID() string {
	if v := os.Getenv(envClientID); v != "" {
		return v
	}
	return EmbeddedClientID
}

// OAuthTokenSource is an api.TokenSource backed by a Resolved profile.
// Auto-refreshes when the access_token is within refreshSkew of expiry
// and persists the new tokens back to wherever they came from (local
// config file, global store, or just in-memory for the env hatch).
type OAuthTokenSource struct {
	resolved *Resolved
	store    GlobalStore // optional, persists global profiles after refresh
	mu       sync.Mutex
}

// refreshSkew - refresh slightly before actual expiry to avoid mid-flight
// 401s from clock drift / network latency.
const refreshSkew = 60 * time.Second

// NewOAuthTokenSource constructs a refreshing token source. store may be
// nil for SourceLocal / SourceEnv (local config writes back to its own
// path; env hatch is single-process and doesn't persist).
func NewOAuthTokenSource(resolved *Resolved, store GlobalStore) *OAuthTokenSource {
	return &OAuthTokenSource{resolved: resolved, store: store}
}

// Token returns a valid access token, refreshing transparently if needed.
func (s *OAuthTokenSource) Token(ctx context.Context) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	p := s.resolved.Profile
	if p == nil {
		return "", errors.New("no credential available")
	}

	hasAccess := p.AccessToken != ""
	if hasAccess && !p.NeedsRefresh(refreshSkew) {
		return p.AccessToken, nil
	}
	if p.RefreshToken == "" {
		return "", errors.New("access token expired and no refresh_token; run `revcat auth login` again")
	}

	clientID := p.ClientID
	if clientID == "" {
		clientID = OAuthClientID()
	}
	if clientID == "" {
		return "", errors.New("oauth client_id is empty; set REVCAT_OAUTH_CLIENT_ID or use a build that has one baked in")
	}
	tok, err := api.RefreshToken(ctx, clientID, "", p.RefreshToken)
	if err != nil {
		return "", fmt.Errorf("oauth refresh: %w", err)
	}
	s.applyToken(tok)
	if err := s.persist(); err != nil {
		return "", fmt.Errorf("persist refreshed tokens: %w", err)
	}
	return p.AccessToken, nil
}

func (s *OAuthTokenSource) applyToken(t *api.TokenResponse) {
	p := s.resolved.Profile
	p.AccessToken = t.AccessToken
	if t.RefreshToken != "" {
		p.RefreshToken = t.RefreshToken
	}
	if t.ExpiresIn > 0 {
		p.ExpiresAt = time.Now().Add(time.Duration(t.ExpiresIn) * time.Second).UnixMilli()
	}
	if t.Scope != "" {
		p.Scope = t.Scope
	}
}

func (s *OAuthTokenSource) persist() error {
	switch s.resolved.Source {
	case SourceLocal:
		if s.resolved.Local == nil || s.resolved.Local.Path == "" {
			return nil
		}
		s.resolved.Local.Profile = *s.resolved.Profile
		return SaveLocal(s.resolved.Local.Path, s.resolved.Local)
	case SourceGlobalFile:
		if s.store == nil {
			return nil
		}
		return s.store.Set(s.resolved.Profile)
	case SourceEnv:
		// In-memory only - the env-hatch path is single-process by
		// design. Refreshed tokens live for the duration of the
		// invocation and disappear with it.
		return nil
	}
	return nil
}
