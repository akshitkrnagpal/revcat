package api

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// OAuth endpoints. RC v2 OAuth doc:
//
//	GET https://api.revenuecat.com/oauth2/authorize
//	POST https://api.revenuecat.com/oauth2/token
//
// Vars (not consts) so tests can point at httptest servers via
// overrideTokenURL.
var (
	OAuthAuthorizeURL = "https://api.revenuecat.com/oauth2/authorize"
	OAuthTokenURL     = "https://api.revenuecat.com/oauth2/token"
)

// overrideTokenURL swaps the OAuth token endpoint and returns a restore
// function. Test helper - not used by production code.
func overrideTokenURL(u string) func() {
	prev := OAuthTokenURL
	OAuthTokenURL = u
	return func() { OAuthTokenURL = prev }
}

// DefaultScopes is the set we ask for on `revcat auth login --oauth`.
// Tracks what the CLI actually exercises today, plus project bootstrap.
var DefaultScopes = []string{
	"project_configuration:projects:read_write",
	"project_configuration:apps:read_write",
	"project_configuration:entitlements:read_write",
	"project_configuration:offerings:read_write",
	"project_configuration:packages:read_write",
	"project_configuration:products:read_write",
	"project_configuration:integrations:read_write",
	"project_configuration:virtual_currencies:read_write",
	"customer_information:customers:read_write",
	"customer_information:subscriptions:read_write",
	"customer_information:purchases:read_write",
	"customer_information:invoices:read",
	"charts_metrics:overview:read",
	"charts_metrics:charts:read",
}

// TokenResponse mirrors RC's OAuth token endpoint payload.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

// PKCEPair is one round of code_verifier + code_challenge used by the
// PKCE flow. Generated fresh per login.
type PKCEPair struct {
	Verifier  string
	Challenge string
}

// NewPKCE generates a fresh verifier (43-128 url-safe chars per RFC 7636)
// and the corresponding S256 challenge.
func NewPKCE() (*PKCEPair, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return nil, err
	}
	verifier := base64.RawURLEncoding.EncodeToString(buf)
	sum := sha256.Sum256([]byte(verifier))
	return &PKCEPair{
		Verifier:  verifier,
		Challenge: base64.RawURLEncoding.EncodeToString(sum[:]),
	}, nil
}

// RandomState returns a base64url-encoded random string suitable for the
// OAuth `state` parameter (CSRF guard for the redirect).
func RandomState() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// AuthorizeURL builds the URL the user opens in the browser.
func AuthorizeURL(clientID, redirectURI string, scopes []string, state, codeChallenge string) string {
	q := url.Values{}
	q.Set("client_id", clientID)
	q.Set("response_type", "code")
	q.Set("redirect_uri", redirectURI)
	q.Set("scope", strings.Join(scopes, " "))
	q.Set("code_challenge", codeChallenge)
	q.Set("code_challenge_method", "S256")
	if state != "" {
		q.Set("state", state)
	}
	return OAuthAuthorizeURL + "?" + q.Encode()
}

// LoopbackServer waits for the OAuth callback on localhost. RC redirects
// the browser to http://127.0.0.1:<port>/callback?code=...&state=... and
// we capture the code. Cancellation via ctx.
type LoopbackServer struct {
	Port  int
	URL   string // full redirect URI registered with RC
	addr  net.Addr
	srv   *http.Server
	done  chan struct{}
	got   *AuthorizationResponse
}

// AuthorizationResponse is the result of the redirect, parsed.
type AuthorizationResponse struct {
	Code  string
	State string
	Err   string // populated when RC redirects with ?error=...
	ErrDesc string
}

// NewLoopbackServer binds 127.0.0.1 on a free port and serves /callback.
// Caller must Close() the server when done.
func NewLoopbackServer() (*LoopbackServer, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("bind callback port: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	ls := &LoopbackServer{
		Port: port,
		URL:  fmt.Sprintf("http://127.0.0.1:%d/callback", port),
		addr: listener.Addr(),
		done: make(chan struct{}),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		ls.got = &AuthorizationResponse{
			Code:    q.Get("code"),
			State:   q.Get("state"),
			Err:     q.Get("error"),
			ErrDesc: q.Get("error_description"),
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if ls.got.Err != "" {
			fmt.Fprintf(w, callbackHTML, "Authorization failed", ls.got.Err+": "+ls.got.ErrDesc)
		} else {
			fmt.Fprintf(w, callbackHTML, "You're signed in", "You can close this tab and return to the terminal.")
		}
		go func() { close(ls.done) }()
	})
	ls.srv = &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	go func() { _ = ls.srv.Serve(listener) }()
	return ls, nil
}

// Wait blocks until the redirect arrives or ctx is cancelled.
func (ls *LoopbackServer) Wait(ctx context.Context) (*AuthorizationResponse, error) {
	select {
	case <-ls.done:
		return ls.got, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Close stops the HTTP server.
func (ls *LoopbackServer) Close() {
	_ = ls.srv.Shutdown(context.Background())
}

// callbackHTML is rendered to the browser tab after the redirect.
const callbackHTML = `<!doctype html>
<html><head><title>revcat auth</title>
<style>
body { font: 16px system-ui, sans-serif; max-width: 480px; margin: 80px auto; padding: 0 24px; color: #1a1a1a; }
h1 { font-size: 22px; margin-bottom: 12px; }
p { color: #555; line-height: 1.5; }
</style></head><body><h1>%s</h1><p>%s</p></body></html>`

// OpenBrowser tries to open the URL in the user's default browser. Best-
// effort - if it fails, the caller should still print the URL so the
// user can copy it manually.
func OpenBrowser(target string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", target)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", target)
	default:
		cmd = exec.Command("xdg-open", target)
	}
	return cmd.Start()
}

// ExchangeCode swaps an authorization code (+ PKCE verifier) for tokens.
// clientSecret is empty for public clients.
func ExchangeCode(ctx context.Context, clientID, clientSecret, code, redirectURI, codeVerifier string) (*TokenResponse, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", redirectURI)
	form.Set("client_id", clientID)
	form.Set("code_verifier", codeVerifier)
	if clientSecret != "" {
		form.Set("client_secret", clientSecret)
	}
	return postToken(ctx, form)
}

// RefreshToken trades a refresh_token for a new access_token (+ usually
// a new refresh_token).
func RefreshToken(ctx context.Context, clientID, clientSecret, refreshToken string) (*TokenResponse, error) {
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)
	form.Set("client_id", clientID)
	if clientSecret != "" {
		form.Set("client_secret", clientSecret)
	}
	return postToken(ctx, form)
}

func postToken(ctx context.Context, form url.Values) (*TokenResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", OAuthTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	resp, err := (&http.Client{Timeout: 30 * time.Second}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("oauth token endpoint %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var out TokenResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decode token response: %w", err)
	}
	if out.AccessToken == "" {
		return nil, errors.New("oauth token response missing access_token")
	}
	return &out, nil
}
