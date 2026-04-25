package api

import (
	"context"
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

// Webhook is a project integration that receives event POSTs.
type Webhook struct {
	ID          string   `json:"id"`
	URL         string   `json:"url"`
	Description string   `json:"description,omitempty"`
	Events      []string `json:"events,omitempty"`
	Disabled    bool     `json:"disabled,omitempty"`
	CreatedAt   int64    `json:"created_at"`
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
type VirtualCurrency struct {
	ID          string `json:"id"`
	LookupKey   string `json:"lookup_key"`
	DisplayName string `json:"display_name"`
	Code        string `json:"code,omitempty"`
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

// ListCustomerVCBalances lists a customer's VC balances.
func (c *Client) ListCustomerVCBalances(ctx context.Context, customerID string) ([]map[string]any, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	return paginate[map[string]any](ctx, c, c.projectPath("/customers/"+url.PathEscape(customerID)+"/virtual_currencies_balances"))
}

// CreateCustomerVCTransaction posts a VC transaction (credit or debit).
func (c *Client) CreateCustomerVCTransaction(ctx context.Context, customerID string, body map[string]any) (map[string]any, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	var out map[string]any
	path := c.projectPath("/customers/" + url.PathEscape(customerID) + "/virtual_currencies/transactions")
	if err := c.Do(ctx, "POST", path, body, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateCustomerVCBalance directly sets a balance.
func (c *Client) UpdateCustomerVCBalance(ctx context.Context, customerID string, body map[string]any) (map[string]any, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	var out map[string]any
	path := c.projectPath("/customers/" + url.PathEscape(customerID) + "/virtual_currencies_balances")
	if err := c.Do(ctx, "POST", path, body, &out); err != nil {
		return nil, err
	}
	return out, nil
}
