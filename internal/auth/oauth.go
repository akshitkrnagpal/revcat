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

// OAuthTokenSource is an api.TokenSource backed by an OAuth profile. It
// auto-refreshes when the access_token is within refreshSkew of expiry
// and persists the new tokens back via the Store.
type OAuthTokenSource struct {
	store   Store
	profile *Profile
	mu      sync.Mutex
}

// refreshSkew - refresh slightly before actual expiry to avoid mid-flight
// 401s from clock drift / network latency.
const refreshSkew = 60 * time.Second

// NewOAuthTokenSource constructs a refreshing token source bound to a
// profile. The profile is updated in place + persisted on every refresh.
func NewOAuthTokenSource(store Store, profile *Profile) *OAuthTokenSource {
	return &OAuthTokenSource{store: store, profile: profile}
}

// Token returns a valid access token, refreshing transparently if needed.
func (s *OAuthTokenSource) Token(ctx context.Context) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.profile.AccessToken == "" {
		return "", errors.New("oauth profile has no access token; run `revcat auth login --oauth` again")
	}
	if !s.expiringSoon() {
		return s.profile.AccessToken, nil
	}
	if s.profile.RefreshToken == "" {
		return "", errors.New("oauth access token expired and no refresh_token; run `revcat auth login --oauth` again")
	}
	clientID := s.profile.ClientID
	if clientID == "" {
		clientID = OAuthClientID()
	}
	if clientID == "" {
		return "", errors.New("oauth client_id is empty; set REVCAT_OAUTH_CLIENT_ID or use a build that has one baked in")
	}
	tok, err := api.RefreshToken(ctx, clientID, "", s.profile.RefreshToken)
	if err != nil {
		return "", fmt.Errorf("oauth refresh: %w", err)
	}
	s.applyToken(tok)
	if err := s.store.Set(s.profile); err != nil {
		// Surfacing the persistence error blocks the request but the
		// in-memory token is still valid for this call. Bubble up so
		// the user knows they need to re-login if persistence is
		// genuinely broken.
		return "", fmt.Errorf("persist refreshed tokens: %w", err)
	}
	return s.profile.AccessToken, nil
}

func (s *OAuthTokenSource) expiringSoon() bool {
	if s.profile.ExpiresAt == 0 {
		return false // token has no expiry recorded; assume usable
	}
	deadline := time.UnixMilli(s.profile.ExpiresAt).Add(-refreshSkew)
	return time.Now().After(deadline)
}

// applyToken updates the profile struct from a fresh token response.
// Refresh tokens may be rotated or omitted; we keep the existing one
// when the response doesn't return a new one.
func (s *OAuthTokenSource) applyToken(t *api.TokenResponse) {
	s.profile.AccessToken = t.AccessToken
	if t.RefreshToken != "" {
		s.profile.RefreshToken = t.RefreshToken
	}
	if t.ExpiresIn > 0 {
		s.profile.ExpiresAt = time.Now().Add(time.Duration(t.ExpiresIn) * time.Second).UnixMilli()
	}
	if t.Scope != "" {
		s.profile.Scope = t.Scope
	}
}

