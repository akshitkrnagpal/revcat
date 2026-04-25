package api

import (
	"context"
	"net/url"
)

// Product is a project-level catalog entry that mirrors a store SKU
// (App Store / Play / Stripe / Web Billing). Field set kept loose - RC
// adds store-specific fields constantly. CreateProduct accepts a free
// map so callers can pass through fields without revcat re-modeling them.
type Product struct {
	ID            string `json:"id,omitempty"`
	StoreIdentifier string `json:"store_identifier"`
	Type          string `json:"type"`
	DisplayName   string `json:"display_name"`
	AppID         string `json:"app_id"`
	CreatedAt     int64  `json:"created_at,omitempty"`
}

// ListProducts pages every product in the active project.
func (c *Client) ListProducts(ctx context.Context) ([]Product, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	return paginate[Product](ctx, c, c.projectPath("/products"))
}

// GetProduct fetches one product by id.
func (c *Client) GetProduct(ctx context.Context, productID string) (*Product, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	var p Product
	if err := c.Do(ctx, "GET", c.projectPath("/products/"+url.PathEscape(productID)), nil, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// CreateProduct creates a product. Body is a free-form map because the
// v2 product schema differs per store; we don't want revcat to lag the API.
func (c *Client) CreateProduct(ctx context.Context, body map[string]any) (*Product, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	var p Product
	if err := c.Do(ctx, "POST", c.projectPath("/products"), body, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// UpdateProduct partially updates a product (display_name, etc.).
func (c *Client) UpdateProduct(ctx context.Context, productID string, body map[string]any) (*Product, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	var p Product
	if err := c.Do(ctx, "POST", c.projectPath("/products/"+url.PathEscape(productID)), body, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// DeleteProduct removes a product. Most teams archive instead.
func (c *Client) DeleteProduct(ctx context.Context, productID string) error {
	if err := c.requireProject(); err != nil {
		return err
	}
	return c.Do(ctx, "DELETE", c.projectPath("/products/"+url.PathEscape(productID)), nil, nil)
}

// ArchiveProduct hides a product from the catalog without deleting it.
// archive=true archives, archive=false unarchives.
func (c *Client) ArchiveProduct(ctx context.Context, productID string, archive bool) error {
	if err := c.requireProject(); err != nil {
		return err
	}
	action := "archive"
	if !archive {
		action = "unarchive"
	}
	path := c.projectPath("/products/" + url.PathEscape(productID) + "/actions/" + action)
	return c.Do(ctx, "POST", path, struct{}{}, nil)
}

// PushProductToStore propagates the product config to the configured
// store backend (StoreKit, Play, Stripe). Async on the store's side; the
// API returns a job id we surface raw.
func (c *Client) PushProductToStore(ctx context.Context, productID string) (map[string]any, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	var out map[string]any
	path := c.projectPath("/products/" + url.PathEscape(productID) + "/actions/push_to_store")
	if err := c.Do(ctx, "POST", path, struct{}{}, &out); err != nil {
		return nil, err
	}
	return out, nil
}
