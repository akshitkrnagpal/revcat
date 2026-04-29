package api

import (
	"context"
	"encoding/json"
	"net/url"
)

// Paywall is a top-level paywall resource. (Distinct from the per-offering
// paywall config that PutPaywall mutates.)
type Paywall struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Template    string `json:"template,omitempty"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at,omitempty"`
}

// ListPaywalls pages through every paywall.
func (c *Client) ListPaywalls(ctx context.Context) ([]Paywall, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	return paginate[Paywall](ctx, c, c.projectPath("/paywalls"))
}

// GetPaywallByID fetches a paywall (top-level resource).
func (c *Client) GetPaywallByID(ctx context.Context, paywallID string) (map[string]any, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	var out map[string]any
	if err := c.Do(ctx, "GET", c.projectPath("/paywalls/"+url.PathEscape(paywallID)), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreatePaywall creates a paywall.
func (c *Client) CreatePaywall(ctx context.Context, body map[string]any) (map[string]any, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	var out map[string]any
	if err := c.Do(ctx, "POST", c.projectPath("/paywalls"), body, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// DeletePaywall removes a paywall.
func (c *Client) DeletePaywall(ctx context.Context, paywallID string) error {
	if err := c.requireProject(); err != nil {
		return err
	}
	return c.Do(ctx, "DELETE", c.projectPath("/paywalls/"+url.PathEscape(paywallID)), nil, nil)
}

// Webhook is a project integration that receives event POSTs. v2 names
// the field event_types and uses lowercase event values like
// "initial_purchase" (not "INITIAL_PURCHASE" - that's the webhook payload
// shape, not the API config).
type Webhook struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	URL        string   `json:"url"`
	EventTypes []string `json:"event_types,omitempty"`
	AppID      string   `json:"app_id,omitempty"`
	Environment string  `json:"environment,omitempty"`
	CreatedAt  int64    `json:"created_at"`
}

// ListWebhooks pages through every webhook integration.
func (c *Client) ListWebhooks(ctx context.Context) ([]Webhook, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	return paginate[Webhook](ctx, c, c.projectPath("/integrations/webhooks"))
}

// GetWebhook fetches one webhook by id.
func (c *Client) GetWebhook(ctx context.Context, id string) (*Webhook, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	var w Webhook
	if err := c.Do(ctx, "GET", c.projectPath("/integrations/webhooks/"+url.PathEscape(id)), nil, &w); err != nil {
		return nil, err
	}
	return &w, nil
}

// GetWebhookRaw fetches one webhook and returns the verbatim v2 response.
func (c *Client) GetWebhookRaw(ctx context.Context, id string) (*Webhook, json.RawMessage, error) {
	if err := c.requireProject(); err != nil {
		return nil, nil, err
	}
	var w Webhook
	raw, err := c.DoRaw(ctx, "GET", c.projectPath("/integrations/webhooks/"+url.PathEscape(id)), nil, &w)
	if err != nil {
		return nil, nil, err
	}
	return &w, raw, nil
}

// CreateWebhook creates a webhook integration.
func (c *Client) CreateWebhook(ctx context.Context, body map[string]any) (*Webhook, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	var w Webhook
	if err := c.Do(ctx, "POST", c.projectPath("/integrations/webhooks"), body, &w); err != nil {
		return nil, err
	}
	return &w, nil
}

// UpdateWebhook partial-updates a webhook.
func (c *Client) UpdateWebhook(ctx context.Context, id string, body map[string]any) (*Webhook, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	var w Webhook
	path := c.projectPath("/integrations/webhooks/" + url.PathEscape(id))
	if err := c.Do(ctx, "POST", path, body, &w); err != nil {
		return nil, err
	}
	return &w, nil
}

// DeleteWebhook removes a webhook.
func (c *Client) DeleteWebhook(ctx context.Context, id string) error {
	if err := c.requireProject(); err != nil {
		return err
	}
	return c.Do(ctx, "DELETE", c.projectPath("/integrations/webhooks/"+url.PathEscape(id)), nil, nil)
}

// VirtualCurrency is a project-level credit/coin balance type.
// Note: v2 keys VCs by their uppercase `code` (e.g., "COIN") - there is
// no separate id or lookup_key field.
type VirtualCurrency struct {
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description,omitempty"`
	State       string `json:"state,omitempty"`
	CreatedAt   int64  `json:"created_at"`
}

// ListVirtualCurrencies pages every VC in the project.
func (c *Client) ListVirtualCurrencies(ctx context.Context) ([]VirtualCurrency, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	return paginate[VirtualCurrency](ctx, c, c.projectPath("/virtual_currencies"))
}

// GetVirtualCurrency fetches one VC by id or lookup_key.
func (c *Client) GetVirtualCurrency(ctx context.Context, idOrKey string) (*VirtualCurrency, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	var v VirtualCurrency
	if err := c.Do(ctx, "GET", c.projectPath("/virtual_currencies/"+url.PathEscape(idOrKey)), nil, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// GetVirtualCurrencyRaw fetches a VC and returns the verbatim v2 response.
func (c *Client) GetVirtualCurrencyRaw(ctx context.Context, idOrKey string) (*VirtualCurrency, json.RawMessage, error) {
	if err := c.requireProject(); err != nil {
		return nil, nil, err
	}
	var v VirtualCurrency
	raw, err := c.DoRaw(ctx, "GET", c.projectPath("/virtual_currencies/"+url.PathEscape(idOrKey)), nil, &v)
	if err != nil {
		return nil, nil, err
	}
	return &v, raw, nil
}

// CreateVirtualCurrency creates a VC.
func (c *Client) CreateVirtualCurrency(ctx context.Context, body map[string]any) (*VirtualCurrency, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	var v VirtualCurrency
	if err := c.Do(ctx, "POST", c.projectPath("/virtual_currencies"), body, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// UpdateVirtualCurrency partial-updates a VC.
func (c *Client) UpdateVirtualCurrency(ctx context.Context, idOrKey string, body map[string]any) (*VirtualCurrency, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	var v VirtualCurrency
	path := c.projectPath("/virtual_currencies/" + url.PathEscape(idOrKey))
	if err := c.Do(ctx, "POST", path, body, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// DeleteVirtualCurrency removes a VC.
func (c *Client) DeleteVirtualCurrency(ctx context.Context, idOrKey string) error {
	if err := c.requireProject(); err != nil {
		return err
	}
	return c.Do(ctx, "DELETE", c.projectPath("/virtual_currencies/"+url.PathEscape(idOrKey)), nil, nil)
}

// ArchiveVirtualCurrency toggles archive state.
func (c *Client) ArchiveVirtualCurrency(ctx context.Context, idOrKey string, archive bool) error {
	if err := c.requireProject(); err != nil {
		return err
	}
	action := "archive"
	if !archive {
		action = "unarchive"
	}
	path := c.projectPath("/virtual_currencies/" + url.PathEscape(idOrKey) + "/actions/" + action)
	return c.Do(ctx, "POST", path, struct{}{}, nil)
}

// Note: per-customer VC endpoints (balances, transactions, set-balance)
// don't exist on the v2 customer surface. Smoke-tested 2026-04-25 - all
// paths 404. Removed from the CLI to avoid shipping broken commands.
