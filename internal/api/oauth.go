package api

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
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
	Port int
	URL  string // full redirect URI registered with RC

	// ProfileName is echoed in the success page. Optional - the caller
	// (login.go) sets it after NewLoopbackServer so the user sees which
	// profile slot was just populated.
	ProfileName string

	addr net.Addr
	srv  *http.Server
	done chan struct{}
	got  *AuthorizationResponse
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
		w.Header().Set("Cache-Control", "no-store")
		_ = callbackTmpl.Execute(w, callbackData{
			Success:     ls.got.Err == "",
			ProfileName: ls.ProfileName,
			ErrCode:     ls.got.Err,
			ErrDesc:     ls.got.ErrDesc,
		})
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

// callbackData drives the post-redirect HTML rendered in the user's
// browser. Success vs error branches in the template; ProfileName is
// echoed when set. ErrCode/ErrDesc come straight from the OAuth server
// (RFC 6749 §4.1.2.1) and are auto-escaped by html/template.
type callbackData struct {
	Success     bool
	ProfileName string
	ErrCode     string
	ErrDesc     string
}

// errorHint maps known OAuth error codes to a one-line hint shown
// alongside the raw err/desc. Keeps the page actionable instead of
// leaving the user to grep error_description for clues.
func (d callbackData) ErrorHint() string {
	switch d.ErrCode {
	case "access_denied":
		return "You clicked Deny on the authorization screen. Run `revcat auth login` again to retry."
	case "invalid_scope", "unauthorized_client":
		return "The OAuth client isn't authorized for the requested scope. Check --client-id or REVCAT_OAUTH_CLIENT_ID."
	case "server_error", "temporarily_unavailable":
		return "RevenueCat's OAuth server returned a transient error. Try `revcat auth login` again in a moment."
	}
	return ""
}

// callbackTmpl is the HTML rendered to the browser tab after the
// redirect. Auto-closes after 3 seconds where the browser permits it
// (Chrome/Edge for tabs opened by the OS launcher; Firefox blocks
// window.close on tabs not opened by script). Falls back to an explicit
// Close button. Honors prefers-color-scheme so it doesn't blast white
// on a dark setup.
var callbackTmpl = template.Must(template.New("callback").Parse(`<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>revcat auth</title>
<style>
:root {
  --bg: #f7f7f8;
  --card: #ffffff;
  --fg: #0f1115;
  --muted: #5c6370;
  --border: #e5e7eb;
  --accent: #4f46e5;
  --code-bg: #f3f4f6;
  --success: #10b981;
  --error: #ef4444;
}
@media (prefers-color-scheme: dark) {
  :root {
    --bg: #0b0d12;
    --card: #14171f;
    --fg: #e6e8ec;
    --muted: #9ba1ad;
    --border: #2a2f3a;
    --accent: #818cf8;
    --code-bg: #1c2030;
  }
}
* { box-sizing: border-box; }
html, body { height: 100%; margin: 0; }
body {
  font: 15px/1.5 -apple-system, BlinkMacSystemFont, "Segoe UI", system-ui, sans-serif;
  background: var(--bg);
  color: var(--fg);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
}
.card {
  width: 100%;
  max-width: 460px;
  background: var(--card);
  border: 1px solid var(--border);
  border-radius: 14px;
  padding: 32px;
  box-shadow: 0 4px 24px rgba(0,0,0,0.04);
}
.brand {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--muted);
  margin-bottom: 22px;
}
.brand-dot {
  width: 7px; height: 7px; border-radius: 999px;
  background: var(--accent);
}
.icon {
  width: 44px; height: 44px;
  border-radius: 999px;
  display: flex; align-items: center; justify-content: center;
  margin-bottom: 16px;
}
.icon.success { background: rgba(16,185,129,0.14); color: var(--success); }
.icon.error { background: rgba(239,68,68,0.12); color: var(--error); }
.icon svg { width: 22px; height: 22px; }
h1 {
  font-size: 19px;
  font-weight: 600;
  margin: 0 0 6px;
  letter-spacing: -0.01em;
}
p { color: var(--muted); margin: 0 0 6px; line-height: 1.5; }
.next {
  margin-top: 18px;
  padding: 12px 14px;
  background: var(--code-bg);
  border-radius: 8px;
  font: 13px/1.5 ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  color: var(--fg);
  word-break: break-all;
}
.next .label {
  display: block;
  font: 11px/1 -apple-system, system-ui, sans-serif;
  font-weight: 600;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--muted);
  margin-bottom: 8px;
}
.errbox {
  margin-top: 14px;
  padding: 12px 14px;
  background: var(--code-bg);
  border-radius: 8px;
  font: 12.5px/1.5 ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  color: var(--fg);
  word-break: break-word;
}
.foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-top: 22px;
  font-size: 12px;
  color: var(--muted);
}
.close {
  appearance: none;
  border: 1px solid var(--border);
  background: transparent;
  color: var(--fg);
  padding: 6px 12px;
  border-radius: 6px;
  font: inherit;
  cursor: pointer;
}
.close:hover { background: var(--code-bg); }
</style>
</head>
<body>
<main class="card">
  <div class="brand"><span class="brand-dot"></span>revcat</div>
  {{if .Success}}
    <div class="icon success" aria-hidden="true">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12.5l4.5 4.5L19 7"/></svg>
    </div>
    <h1>You're signed in{{if .ProfileName}} as &ldquo;{{.ProfileName}}&rdquo;{{end}}</h1>
    <p>You can close this tab and return to the terminal.</p>
    <div class="next">
      <span class="label">Next step</span>cd ~/your-repo &amp;&amp; revcat init
    </div>
  {{else}}
    <div class="icon error" aria-hidden="true">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"><path d="M6 6l12 12M18 6L6 18"/></svg>
    </div>
    <h1>Authorization failed</h1>
    {{with .ErrorHint}}<p>{{.}}</p>{{end}}
    <div class="errbox"><strong>{{.ErrCode}}</strong>{{if .ErrDesc}}: {{.ErrDesc}}{{end}}</div>
  {{end}}
  <div class="foot">
    <span id="countdown">You can close this tab.</span>
    <button class="close" type="button" onclick="window.close()">Close</button>
  </div>
</main>
<script>
(function () {
  var el = document.getElementById('countdown');
  var n = 3;
  function tick() {
    if (n <= 0) {
      window.close();
      el.textContent = 'You can close this tab.';
      return;
    }
    el.textContent = 'Closing in ' + n + 's…';
    n--;
    setTimeout(tick, 1000);
  }
  tick();
})();
</script>
</body>
</html>`))

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
