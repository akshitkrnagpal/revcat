package api

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

// stubToken is a TokenSource for tests - returns the embedded literal,
// no refresh logic.
type stubToken string

func (s stubToken) Token(_ context.Context) (string, error) {
	return string(s), nil
}

func TestNewPKCE_S256ChallengeMatchesVerifier(t *testing.T) {
	p, err := NewPKCE()
	if err != nil {
		t.Fatalf("NewPKCE: %v", err)
	}
	if len(p.Verifier) < 43 || len(p.Verifier) > 128 {
		t.Fatalf("verifier length %d out of RFC 7636 range [43,128]", len(p.Verifier))
	}
	sum := sha256.Sum256([]byte(p.Verifier))
	want := base64.RawURLEncoding.EncodeToString(sum[:])
	if p.Challenge != want {
		t.Fatalf("challenge does not match SHA-256(verifier): got %q want %q", p.Challenge, want)
	}
}

func TestNewPKCE_VerifierIsRandom(t *testing.T) {
	a, _ := NewPKCE()
	b, _ := NewPKCE()
	if a.Verifier == b.Verifier {
		t.Fatalf("two PKCE pairs have the same verifier; rand source broken")
	}
}

func TestAuthorizeURL_IncludesPKCEAndState(t *testing.T) {
	got := AuthorizeURL("client_x", "http://127.0.0.1:1234/callback",
		[]string{"a:read", "b:read_write"}, "state_xyz", "challenge_xyz")
	u, err := url.Parse(got)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	q := u.Query()
	for k, want := range map[string]string{
		"response_type":         "code",
		"client_id":             "client_x",
		"redirect_uri":          "http://127.0.0.1:1234/callback",
		"state":                 "state_xyz",
		"code_challenge":        "challenge_xyz",
		"code_challenge_method": "S256",
		"scope":                 "a:read b:read_write",
	} {
		if got := q.Get(k); got != want {
			t.Errorf("%s: got %q want %q", k, got, want)
		}
	}
}

func TestExchangeCode_PostsFormAndReturnsTokens(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method: got %s want POST", r.Method)
		}
		if got := r.Header.Get("Content-Type"); got != "application/x-www-form-urlencoded" {
			t.Errorf("Content-Type: %q", got)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		for k, want := range map[string]string{
			"grant_type":    "authorization_code",
			"code":          "code_abc",
			"redirect_uri":  "http://127.0.0.1:1/callback",
			"client_id":     "client_x",
			"code_verifier": "verifier_xyz",
		} {
			if got := r.PostForm.Get(k); got != want {
				t.Errorf("form[%s]: got %q want %q", k, got, want)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"access_token": "atk_new",
			"token_type": "Bearer",
			"expires_in": 3600,
			"refresh_token": "rtk_new",
			"scope": "a b"
		}`))
	}))
	t.Cleanup(srv.Close)

	t.Setenv("REVCAT_OAUTH_TOKEN_URL", "") // unused, but documents intent
	prev := overrideTokenURL(srv.URL)
	defer prev()

	tok, err := ExchangeCode(context.Background(), "client_x", "", "code_abc",
		"http://127.0.0.1:1/callback", "verifier_xyz")
	if err != nil {
		t.Fatalf("ExchangeCode: %v", err)
	}
	if tok.AccessToken != "atk_new" || tok.RefreshToken != "rtk_new" {
		t.Fatalf("tokens: %+v", tok)
	}
	if tok.ExpiresIn != 3600 {
		t.Fatalf("expires_in: %d", tok.ExpiresIn)
	}
}

func TestRefreshToken_PostsRefreshGrant(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if got := r.PostForm.Get("grant_type"); got != "refresh_token" {
			t.Errorf("grant_type: %q", got)
		}
		if got := r.PostForm.Get("refresh_token"); got != "rtk_old" {
			t.Errorf("refresh_token: %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"atk_x","token_type":"Bearer","expires_in":1200,"refresh_token":"rtk_rotated"}`))
	}))
	t.Cleanup(srv.Close)

	prev := overrideTokenURL(srv.URL)
	defer prev()

	tok, err := RefreshToken(context.Background(), "client_x", "", "rtk_old")
	if err != nil {
		t.Fatalf("RefreshToken: %v", err)
	}
	if tok.AccessToken != "atk_x" || tok.RefreshToken != "rtk_rotated" {
		t.Fatalf("tokens: %+v", tok)
	}
}

func TestPostToken_ErrorBodyIsSurfaced(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid_grant","error_description":"code expired"}`))
	}))
	t.Cleanup(srv.Close)

	prev := overrideTokenURL(srv.URL)
	defer prev()

	_, err := ExchangeCode(context.Background(), "c", "", "code", "http://x/cb", "v")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "invalid_grant") || !strings.Contains(err.Error(), "code expired") {
		t.Fatalf("error did not surface body: %v", err)
	}
}

func TestLoopbackServer_CapturesCodeAndState(t *testing.T) {
	ls, err := NewLoopbackServer()
	if err != nil {
		t.Fatalf("NewLoopbackServer: %v", err)
	}
	t.Cleanup(ls.Close)

	go func() {
		_, _ = http.Get(ls.URL + "?code=auth_code_1&state=s_1")
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	resp, err := ls.Wait(ctx)
	if err != nil {
		t.Fatalf("Wait: %v", err)
	}
	if resp.Code != "auth_code_1" || resp.State != "s_1" {
		t.Fatalf("got %+v", resp)
	}
}
