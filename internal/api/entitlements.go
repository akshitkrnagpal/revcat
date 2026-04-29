package api

import (
	"context"
	"encoding/json"
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

// ResolveEntitlementID accepts either a system id (`entl...`) or a
// human-friendly lookup_key (`premium`) and returns the system id.
func (c *Client) ResolveEntitlementID(ctx context.Context, idOrKey string) (string, error) {
	if err := c.requireProject(); err != nil {
		return "", err
	}
	var probe Entitlement
	if err := c.Do(ctx, "GET", c.projectPath("/entitlements/"+url.PathEscape(idOrKey)), nil, &probe); err == nil {
		return probe.ID, nil
	}
	all, err := c.ListEntitlements(ctx)
	if err != nil {
		return "", err
	}
	for _, e := range all {
		if e.LookupKey == idOrKey || e.ID == idOrKey {
			return e.ID, nil
		}
	}
	return "", &APIError{Status: 404, StatusText: "Not Found", Message: "no entitlement with id or lookup_key " + idOrKey}
}

// GetEntitlement fetches a single entitlement. Accepts id or lookup_key.
func (c *Client) GetEntitlement(ctx context.Context, idOrKey string) (*Entitlement, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	id, err := c.ResolveEntitlementID(ctx, idOrKey)
	if err != nil {
		return nil, err
	}
	var out Entitlement
	path := c.projectPath("/entitlements/" + url.PathEscape(id))
	if err := c.Do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetEntitlementRaw fetches an entitlement and returns the verbatim v2
// response alongside the typed projection.
func (c *Client) GetEntitlementRaw(ctx context.Context, idOrKey string) (*Entitlement, json.RawMessage, error) {
	if err := c.requireProject(); err != nil {
		return nil, nil, err
	}
	id, err := c.ResolveEntitlementID(ctx, idOrKey)
	if err != nil {
		return nil, nil, err
	}
	var out Entitlement
	path := c.projectPath("/entitlements/" + url.PathEscape(id))
	raw, err := c.DoRaw(ctx, "GET", path, nil, &out)
	if err != nil {
		return nil, nil, err
	}
	return &out, raw, nil
}

// CreateEntitlement creates a new entitlement.
func (c *Client) CreateEntitlement(ctx context.Context, body map[string]any) (*Entitlement, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	var out Entitlement
	if err := c.Do(ctx, "POST", c.projectPath("/entitlements"), body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateEntitlement partial-updates an entitlement.
func (c *Client) UpdateEntitlement(ctx context.Context, idOrKey string, body map[string]any) (*Entitlement, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	id, err := c.ResolveEntitlementID(ctx, idOrKey)
	if err != nil {
		return nil, err
	}
	var out Entitlement
	path := c.projectPath("/entitlements/" + url.PathEscape(id))
	if err := c.Do(ctx, "POST", path, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteEntitlement removes an entitlement.
func (c *Client) DeleteEntitlement(ctx context.Context, idOrKey string) error {
	if err := c.requireProject(); err != nil {
		return err
	}
	id, err := c.ResolveEntitlementID(ctx, idOrKey)
	if err != nil {
		return err
	}
	return c.Do(ctx, "DELETE", c.projectPath("/entitlements/"+url.PathEscape(id)), nil, nil)
}

// ArchiveEntitlement toggles archive state.
func (c *Client) ArchiveEntitlement(ctx context.Context, idOrKey string, archive bool) error {
	if err := c.requireProject(); err != nil {
		return err
	}
	id, err := c.ResolveEntitlementID(ctx, idOrKey)
	if err != nil {
		return err
	}
	action := "archive"
	if !archive {
		action = "unarchive"
	}
	path := c.projectPath("/entitlements/" + url.PathEscape(id) + "/actions/" + action)
	return c.Do(ctx, "POST", path, struct{}{}, nil)
}

// ListEntitlementProducts lists products attached to an entitlement.
// v2's `GET /entitlements/{id}/products` returns bare Product items
// (unlike packages, which wrap them in a binding object).
func (c *Client) ListEntitlementProducts(ctx context.Context, idOrKey string) ([]Product, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	id, err := c.ResolveEntitlementID(ctx, idOrKey)
	if err != nil {
		return nil, err
	}
	return paginate[Product](ctx, c, c.projectPath("/entitlements/"+url.PathEscape(id)+"/products"))
}

// ListEntitlementProductsRaw is ListEntitlementProducts plus the
// verbatim per-item bytes for --output json so the full v2 product
// shape (app_id, store_identifier, subscription, etc.) reaches users.
func (c *Client) ListEntitlementProductsRaw(ctx context.Context, idOrKey string) ([]Product, []json.RawMessage, error) {
	if err := c.requireProject(); err != nil {
		return nil, nil, err
	}
	id, err := c.ResolveEntitlementID(ctx, idOrKey)
	if err != nil {
		return nil, nil, err
	}
	return paginateBoth[Product](ctx, c, c.projectPath("/entitlements/"+url.PathEscape(id)+"/products"))
}

// AttachProductsToEntitlement adds products that grant this entitlement.
func (c *Client) AttachProductsToEntitlement(ctx context.Context, idOrKey string, productIDs []string) error {
	if err := c.requireProject(); err != nil {
		return err
	}
	id, err := c.ResolveEntitlementID(ctx, idOrKey)
	if err != nil {
		return err
	}
	body := map[string]any{"product_ids": productIDs}
	path := c.projectPath("/entitlements/" + url.PathEscape(id) + "/actions/attach_products")
	return c.Do(ctx, "POST", path, body, nil)
}

// DetachProductsFromEntitlement removes products from an entitlement.
func (c *Client) DetachProductsFromEntitlement(ctx context.Context, idOrKey string, productIDs []string) error {
	if err := c.requireProject(); err != nil {
		return err
	}
	id, err := c.ResolveEntitlementID(ctx, idOrKey)
	if err != nil {
		return err
	}
	body := map[string]any{"product_ids": productIDs}
	path := c.projectPath("/entitlements/" + url.PathEscape(id) + "/actions/detach_products")
	return c.Do(ctx, "POST", path, body, nil)
}
