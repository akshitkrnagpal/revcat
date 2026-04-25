package api

import (
	"context"
	"net/url"
)

// Offering is a presentation grouping of packages shown on a paywall.
// Each project has 0..N offerings; exactly one is "current" (the default
// shown when SDKs ask for the current offering).
type Offering struct {
	ID          string    `json:"id"`
	LookupKey   string    `json:"lookup_key"`
	DisplayName string    `json:"display_name"`
	IsCurrent   bool      `json:"is_current"`
	CreatedAt   int64     `json:"created_at"`
	Packages    []Package `json:"packages,omitempty"`
}

// Package is a single purchasable inside an offering. Identifier follows
// RC's $rc_monthly / $rc_annual / custom convention.
type Package struct {
	ID         string `json:"id"`
	Identifier string `json:"identifier"`
	Position   int    `json:"position"`
	ProductID  string `json:"product_id,omitempty"`
	OfferingID string `json:"offering_id,omitempty"`
	CreatedAt  int64  `json:"created_at"`
}

// ListOfferings pages through every offering in the active project.
func (c *Client) ListOfferings(ctx context.Context) ([]Offering, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	return paginate[Offering](ctx, c, c.projectPath("/offerings"))
}

// GetOffering fetches a single offering. If withPackages is true the
// packages list is fetched in a follow-up call.
func (c *Client) GetOffering(ctx context.Context, lookupKey string, withPackages bool) (*Offering, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	var out Offering
	path := c.projectPath("/offerings/" + url.PathEscape(lookupKey))
	if err := c.Do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	if withPackages {
		pkgs, err := c.ListPackages(ctx, lookupKey)
		if err != nil {
			return &out, err
		}
		out.Packages = pkgs
	}
	return &out, nil
}

// ListPackages pages through the packages in an offering.
func (c *Client) ListPackages(ctx context.Context, offeringLookupKey string) ([]Package, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	return paginate[Package](ctx, c, c.projectPath("/offerings/"+url.PathEscape(offeringLookupKey)+"/packages"))
}

// SetCurrentOffering promotes the named offering to current. The v2
// docs model this as an offering update with `is_current: true`, not a
// dedicated action. Wraps POST /offerings/{id}.
func (c *Client) SetCurrentOffering(ctx context.Context, lookupKey string) (*Offering, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	var out Offering
	path := c.projectPath("/offerings/" + url.PathEscape(lookupKey))
	body := map[string]any{"is_current": true}
	if err := c.Do(ctx, "POST", path, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetPaywall returns the raw paywall config json for an offering. Returned
// as a generic map so we don't have to model RC's evolving paywalls v2
// schema in revcat just to ship publish.
func (c *Client) GetPaywall(ctx context.Context, offeringLookupKey string) (map[string]any, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	var out map[string]any
	path := c.projectPath("/offerings/" + url.PathEscape(offeringLookupKey) + "/paywall")
	if err := c.Do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// PutPaywall replaces the paywall config for an offering with the supplied
// json blob. Validation is server-side; we pass the body through verbatim
// after a json.Marshal round-trip so a typo in the file is caught before
// the network call.
func (c *Client) PutPaywall(ctx context.Context, offeringLookupKey string, body map[string]any) error {
	if err := c.requireProject(); err != nil {
		return err
	}
	path := c.projectPath("/offerings/" + url.PathEscape(offeringLookupKey) + "/paywall")
	return c.Do(ctx, "PUT", path, body, nil)
}
