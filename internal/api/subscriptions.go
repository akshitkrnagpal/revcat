package api

import (
	"context"
	"errors"
	"net/url"
)

// SubscriptionFull is the full v2 subscription record (more fields than the
// embedded Subscription type used in customer snapshots).
type SubscriptionFull struct {
	ID                 string   `json:"id"`
	Status             string   `json:"status"`
	ProductID          string   `json:"product_id"`
	Store              string   `json:"store"`
	StartsAt           int64    `json:"starts_at"`
	CurrentEndsAt      int64    `json:"current_period_ends_at,omitempty"`
	WillRenew          bool     `json:"will_renew"`
	IsTrial            bool     `json:"is_in_trial_period"`
	IsSandbox          bool     `json:"is_sandbox"`
	UnsubscribeAt      int64    `json:"unsubscribe_detected_at,omitempty"`
	CustomerID         string   `json:"customer_id,omitempty"`
	EntitlementIDs     []string `json:"entitlement_ids,omitempty"`
}

// Transaction is a single billing event under a subscription.
type Transaction struct {
	ID         string  `json:"id"`
	StoreID    string  `json:"store_transaction_id,omitempty"`
	Status     string  `json:"status"`
	Amount     float64 `json:"amount"`
	Currency   string  `json:"currency"`
	OccurredAt int64   `json:"occurred_at"`
	IsSandbox  bool    `json:"is_sandbox"`
}

// GetSubscription fetches one subscription by id.
func (c *Client) GetSubscription(ctx context.Context, subID string) (*SubscriptionFull, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	var s SubscriptionFull
	if err := c.Do(ctx, "GET", c.projectPath("/subscriptions/"+url.PathEscape(subID)), nil, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// ListSubscriptionTransactions pages billing events for a subscription.
func (c *Client) ListSubscriptionTransactions(ctx context.Context, subID string) ([]Transaction, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	return paginate[Transaction](ctx, c, c.projectPath("/subscriptions/"+url.PathEscape(subID)+"/transactions"))
}

// ListSubscriptionEntitlements lists entitlements granted by a subscription.
func (c *Client) ListSubscriptionEntitlements(ctx context.Context, subID string) ([]Entitlement, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	return paginate[Entitlement](ctx, c, c.projectPath("/subscriptions/"+url.PathEscape(subID)+"/entitlements"))
}

// CancelSubscription cancels a Web Billing subscription.
func (c *Client) CancelSubscription(ctx context.Context, subID string) (map[string]any, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	var out map[string]any
	path := c.projectPath("/subscriptions/" + url.PathEscape(subID) + "/actions/cancel")
	if err := c.Do(ctx, "POST", path, struct{}{}, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// RefundSubscription refunds a subscription (Web Billing).
func (c *Client) RefundSubscription(ctx context.Context, subID string) (map[string]any, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	var out map[string]any
	path := c.projectPath("/subscriptions/" + url.PathEscape(subID) + "/actions/refund")
	if err := c.Do(ctx, "POST", path, struct{}{}, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// SubscriptionManagementURL returns the store-specific cancel URL.
func (c *Client) SubscriptionManagementURL(ctx context.Context, subID string) (string, error) {
	if err := c.requireProject(); err != nil {
		return "", err
	}
	var out struct {
		URL string `json:"management_url"`
	}
	path := c.projectPath("/subscriptions/" + url.PathEscape(subID) + "/management_url")
	if err := c.Do(ctx, "GET", path, nil, &out); err != nil {
		return "", err
	}
	return out.URL, nil
}

// SearchSubscriptions looks up a subscription by store id (App Store
// transaction id, Play original_purchase_token, Stripe sub_..., etc.).
// v2 returns 404 instead of an empty list when there are no matches;
// we translate that to the empty list so callers get a uniform shape.
func (c *Client) SearchSubscriptions(ctx context.Context, query string) ([]SubscriptionFull, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	q := url.Values{}
	q.Set("query", query)
	q.Set("limit", "100")
	path := c.projectPath("/subscriptions/search?" + q.Encode())
	var page listResp[SubscriptionFull]
	if err := c.Do(ctx, "GET", path, nil, &page); err != nil {
		var apiErr *APIError
		if errors.As(err, &apiErr) && apiErr.IsNotFound() {
			return []SubscriptionFull{}, nil
		}
		return nil, err
	}
	if page.Items == nil {
		return []SubscriptionFull{}, nil
	}
	return page.Items, nil
}

// PurchaseFull is the full v2 non-renewing purchase record.
type PurchaseFull struct {
	ID         string  `json:"id"`
	ProductID  string  `json:"product_id"`
	Store      string  `json:"store"`
	StoreID    string  `json:"store_transaction_id,omitempty"`
	PurchaseAt int64   `json:"purchased_at"`
	Amount     float64 `json:"amount,omitempty"`
	Currency   string  `json:"currency,omitempty"`
	IsSandbox  bool    `json:"is_sandbox"`
	CustomerID string  `json:"customer_id,omitempty"`
}

// GetPurchase fetches one non-renewing purchase by id.
func (c *Client) GetPurchase(ctx context.Context, purchaseID string) (*PurchaseFull, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	var p PurchaseFull
	if err := c.Do(ctx, "GET", c.projectPath("/purchases/"+url.PathEscape(purchaseID)), nil, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// ListPurchaseEntitlements lists entitlements granted by a purchase.
func (c *Client) ListPurchaseEntitlements(ctx context.Context, purchaseID string) ([]Entitlement, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	return paginate[Entitlement](ctx, c, c.projectPath("/purchases/"+url.PathEscape(purchaseID)+"/entitlements"))
}

// RefundPurchase refunds a non-renewing purchase (Web Billing).
func (c *Client) RefundPurchase(ctx context.Context, purchaseID string) (map[string]any, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	var out map[string]any
	path := c.projectPath("/purchases/" + url.PathEscape(purchaseID) + "/actions/refund")
	if err := c.Do(ctx, "POST", path, struct{}{}, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// SearchPurchases looks up a purchase by store id. v2 returns 404
// instead of an empty list when there are no matches; we normalize.
func (c *Client) SearchPurchases(ctx context.Context, query string) ([]PurchaseFull, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	q := url.Values{}
	q.Set("query", query)
	q.Set("limit", "100")
	path := c.projectPath("/purchases/search?" + q.Encode())
	var page listResp[PurchaseFull]
	if err := c.Do(ctx, "GET", path, nil, &page); err != nil {
		var apiErr *APIError
		if errors.As(err, &apiErr) && apiErr.IsNotFound() {
			return []PurchaseFull{}, nil
		}
		return nil, err
	}
	if page.Items == nil {
		return []PurchaseFull{}, nil
	}
	return page.Items, nil
}

// GetInvoice fetches one invoice by id.
func (c *Client) GetInvoice(ctx context.Context, invoiceID string) (map[string]any, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	var out map[string]any
	if err := c.Do(ctx, "GET", c.projectPath("/invoices/"+url.PathEscape(invoiceID)), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}
