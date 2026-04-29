package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// V1BaseURL is the legacy v1 REST root used by the SDK-facing endpoints
// (subscriber state, offerings preview). Public SDK keys talk to v1, not v2.
const V1BaseURL = "https://api.revenuecat.com/v1"

// PreviewOffering mirrors the v1 `/offerings` shape returned to the SDK.
// Fields are the SDK-facing identifiers, not v2 internal ids.
type PreviewOffering struct {
	Identifier  string           `json:"identifier"`
	Description string           `json:"description,omitempty"`
	Packages    []PreviewPackage `json:"packages"`
	Metadata    map[string]any   `json:"metadata,omitempty"`
}

// PreviewPackage is a single SDK-facing package as it appears in the v1
// offerings payload.
type PreviewPackage struct {
	Identifier                string `json:"identifier"`
	PlatformProductIdentifier string `json:"platform_product_identifier"`
}

// PreviewOfferingsResponse is the full body returned by GET
// /v1/subscribers/{user_id}/offerings.
type PreviewOfferingsResponse struct {
	CurrentOfferingID string            `json:"current_offering_id"`
	Offerings         []PreviewOffering `json:"offerings"`
}

// PreviewOfferings hits the v1 SDK-facing offerings endpoint with the given
// public SDK key (Bearer) and X-Platform header. This is the same payload
// the on-device SDK receives from `Purchases.getOfferings()`, so it is the
// source of truth when debugging "SDK sees zero packages" symptoms.
//
// publicSDKKey must be a public key (one of the keys returned by
// `revcat apps public-keys <app_id>`), NOT the v2 secret key. platform is
// sent verbatim in the X-Platform header (e.g. "iOS", "android", "web").
func (c *Client) PreviewOfferings(ctx context.Context, userID, publicSDKKey, platform string) (*PreviewOfferingsResponse, error) {
	if userID == "" {
		return nil, fmt.Errorf("preview offerings: user id required")
	}
	if publicSDKKey == "" {
		return nil, fmt.Errorf("preview offerings: public SDK key required")
	}
	if platform == "" {
		return nil, fmt.Errorf("preview offerings: platform required")
	}
	path := "/subscribers/" + url.PathEscape(userID) + "/offerings"
	headers := map[string]string{"X-Platform": platform}
	var out PreviewOfferingsResponse
	if err := c.doV1Bearer(ctx, "GET", path, publicSDKKey, headers, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// doV1Bearer performs a single request against the v1 base URL with a
// caller-supplied bearer token (typically a public SDK key) instead of the
// client's stored v2 secret. No retries: v1 SDK endpoints are read-only and
// callers can re-run trivially.
func (c *Client) doV1Bearer(ctx context.Context, method, path, bearer string, extraHeaders map[string]string, out any) error {
	base := V1BaseURL
	if env := strings.TrimSpace(os.Getenv("REVCAT_V1_BASE_URL")); env != "" {
		base = env
	}
	reqURL := base + path

	req, err := http.NewRequestWithContext(ctx, method, reqURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+bearer)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent+"/"+c.version)
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}

	c.logRequest(req, nil)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	c.logResponse(resp, respBody)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return decodeError(resp.StatusCode, resp.Status, respBody)
	}
	if out == nil || len(respBody) == 0 {
		return nil
	}
	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}
