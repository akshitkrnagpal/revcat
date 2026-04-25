package api

import (
	"context"
	"net/url"
)

// Entitlement is the project-level definition of a feature flag tied to
// access (e.g., "premium", "pro"). Customers gain entitlements via
// products attached on offerings, or via promotional grants.
type Entitlement struct {
	ID          string `json:"id"`
	LookupKey   string `json:"lookup_key"`
	DisplayName string `json:"display_name"`
	CreatedAt   int64  `json:"created_at"`
	ProjectID   string `json:"project_id,omitempty"`
}

// ListEntitlements pages through every entitlement in the active project.
func (c *Client) ListEntitlements(ctx context.Context) ([]Entitlement, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	return paginate[Entitlement](ctx, c, c.projectPath("/entitlements"))
}

// GetEntitlement fetches a single entitlement by its lookup_key (the
// human-friendly id like "premium", which RC also accepts as the path id).
func (c *Client) GetEntitlement(ctx context.Context, lookupKey string) (*Entitlement, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	var out Entitlement
	path := c.projectPath("/entitlements/" + url.PathEscape(lookupKey))
	if err := c.Do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
