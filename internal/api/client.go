// Package api wraps the RevenueCat v2 REST API.
//
// Hand-rolled (RC ships no Go SDK). Uses net/http stdlib with a thin retry
// layer for 429 + 5xx, and exposes typed methods per resource.
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// BaseURL is the v2 REST root. Override for test/staging via env.
const BaseURL = "https://api.revenuecat.com/v2"

const userAgent = "revcat-cli"

// envDebug toggles full request/response logging (key redacted).
const envDebug = "REVCAT_DEBUG"

// TokenSource hands out a current bearer token. Used to plug in either a
// static secret key or an OAuth flow that auto-refreshes. token() is
// called before each request, so an OAuth implementation can do a
// refresh round-trip on the fly without the caller knowing.
type TokenSource interface {
	Token(ctx context.Context) (string, error)
}

// staticToken is the trivial source for project secret keys.
type staticToken string

func (s staticToken) Token(_ context.Context) (string, error) {
	if s == "" {
		return "", errors.New("no auth token configured")
	}
	return string(s), nil
}

// Client talks to the RC v2 REST API on behalf of a single profile.
type Client struct {
	http      *http.Client
	baseURL   string
	tokenSrc  TokenSource
	projectID string
	debug     bool
	version   string
}

// New constructs a Client. SecretKey is required; ProjectID may be empty
// for endpoints that don't need it (auth checks, project listing).
//
// Pass TokenSource for OAuth profiles where the access token rotates;
// otherwise pass SecretKey and revcat builds a static source.
type Options struct {
	SecretKey   string
	TokenSource TokenSource
	ProjectID   string
	BaseURL     string
	Version     string
}

func New(opts Options) *Client {
	base := opts.BaseURL
	if base == "" {
		base = BaseURL
	}
	src := opts.TokenSource
	if src == nil {
		src = staticToken(opts.SecretKey)
	}
	return &Client{
		http:      &http.Client{Timeout: 30 * time.Second},
		baseURL:   base,
		tokenSrc:  src,
		projectID: opts.ProjectID,
		debug:     strings.Contains(os.Getenv(envDebug), "api"),
		version:   opts.Version,
	}
}

// ProjectID returns the configured project id (may be "").
func (c *Client) ProjectID() string { return c.projectID }

// APIError is returned for non-2xx responses. The body, if it parsed as
// JSON, is exposed verbatim for downstream rendering.
type APIError struct {
	Status     int
	StatusText string
	Type       string `json:"type"`
	Message    string `json:"message"`
	DocURL     string `json:"doc_url"`
	Raw        string
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("api %d %s: %s", e.Status, e.StatusText, e.Message)
	}
	return fmt.Sprintf("api %d %s", e.Status, e.StatusText)
}

// IsNotFound is true when the API returned 404.
func (e *APIError) IsNotFound() bool { return e.Status == http.StatusNotFound }

// Do issues a JSON request and decodes the response into out.
//
// Retries: 429 (respecting Retry-After when present) and 5xx, up to 3
// attempts with exponential backoff (200ms, 600ms, 1.8s).
func (c *Client) Do(ctx context.Context, method, path string, body any, out any) error {
	_, err := c.doRaw(ctx, method, path, body, out)
	return err
}

// DoRaw is like Do but also returns the verbatim response body. Used by
// view commands that want to surface the full v2 response in --output json
// without dropping fields the typed struct doesn't model.
func (c *Client) DoRaw(ctx context.Context, method, path string, body any, out any) (json.RawMessage, error) {
	return c.doRaw(ctx, method, path, body, out)
}

func (c *Client) doRaw(ctx context.Context, method, path string, body any, out any) (json.RawMessage, error) {
	url := c.baseURL + path

	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("encode request: %w", err)
		}
	}

	const maxAttempts = 3
	backoff := 200 * time.Millisecond

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(bodyBytes))
		if err != nil {
			return nil, err
		}
		token, err := c.tokenSrc.Token(ctx)
		if err != nil {
			return fmt.Errorf("auth: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", userAgent+"/"+c.version)
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		c.logRequest(req, bodyBytes)

		resp, err := c.http.Do(req)
		if err != nil {
			if attempt == maxAttempts {
				return nil, fmt.Errorf("request: %w", err)
			}
			time.Sleep(backoff)
			backoff *= 3
			continue
		}

		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		c.logResponse(resp, respBody)

		if resp.StatusCode == 429 || resp.StatusCode >= 500 {
			if attempt == maxAttempts {
				return nil, decodeError(resp.StatusCode, resp.Status, respBody)
			}
			wait := backoff
			if ra := resp.Header.Get("Retry-After"); ra != "" {
				if d, perr := time.ParseDuration(ra + "s"); perr == nil {
					wait = d
				}
			}
			time.Sleep(wait)
			backoff *= 3
			continue
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, decodeError(resp.StatusCode, resp.Status, respBody)
		}

		if len(respBody) == 0 {
			return nil, nil
		}
		if out != nil {
			if err := json.Unmarshal(respBody, out); err != nil {
				return nil, fmt.Errorf("decode response: %w", err)
			}
		}
		return json.RawMessage(respBody), nil
	}
	return nil, errors.New("exhausted retries")
}

func decodeError(status int, statusText string, body []byte) error {
	e := &APIError{Status: status, StatusText: statusText, Raw: string(body)}
	if len(body) > 0 {
		_ = json.Unmarshal(body, e)
	}
	return e
}

func (c *Client) logRequest(req *http.Request, body []byte) {
	if !c.debug {
		return
	}
	fmt.Fprintf(os.Stderr, "→ %s %s\n", req.Method, req.URL)
	for k, v := range req.Header {
		val := strings.Join(v, ",")
		if k == "Authorization" {
			val = "Bearer ***"
		}
		fmt.Fprintf(os.Stderr, "  %s: %s\n", k, val)
	}
	if len(body) > 0 {
		fmt.Fprintf(os.Stderr, "  body: %s\n", string(body))
	}
}

func (c *Client) logResponse(resp *http.Response, body []byte) {
	if !c.debug {
		return
	}
	fmt.Fprintf(os.Stderr, "← %d %s\n", resp.StatusCode, resp.Status)
	if len(body) > 0 && len(body) < 4096 {
		fmt.Fprintf(os.Stderr, "  body: %s\n", string(body))
	}
}
