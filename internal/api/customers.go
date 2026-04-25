package api

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
)

// Customer is the v2 representation of a single end-user (subscriber).
// Trimmed to the fields revcat surfaces; extend as commands need more.
type Customer struct {
	ID             string            `json:"id"`
	ProjectID      string            `json:"project_id"`
	FirstSeen      int64             `json:"first_seen_at"`
	LastSeen       int64             `json:"last_seen_at"`
	Country        string            `json:"country"`
	ActiveEntCount int               `json:"active_entitlements_count,omitempty"`
	Attributes     map[string]string `json:"attributes,omitempty"`
}

// ActiveEntitlement is the per-customer view of a granted entitlement.
type ActiveEntitlement struct {
	ID              string `json:"id"`
	LookupKey       string `json:"lookup_key"`
	DisplayName     string `json:"display_name"`
	GrantedAt       int64  `json:"granted_at"`
	ExpiresAt       int64  `json:"expires_at,omitempty"`
	WillRenew       bool   `json:"will_renew"`
	UnsubscribeAt   int64  `json:"unsubscribe_detected_at,omitempty"`
	IsSandbox       bool   `json:"is_sandbox"`
	ProductID       string `json:"product_id,omitempty"`
	Store           string `json:"store,omitempty"`
	IsPromotional   bool   `json:"is_promotional,omitempty"`
}

// Subscription is the v2 representation of an active or recent subscription.
type Subscription struct {
	ID             string `json:"id"`
	ProductID      string `json:"product_id"`
	Status         string `json:"status"`
	Store          string `json:"store"`
	StartsAt       int64  `json:"starts_at"`
	CurrentEndsAt  int64  `json:"current_period_ends_at,omitempty"`
	WillRenew      bool   `json:"will_renew"`
	IsTrial        bool   `json:"is_in_trial_period,omitempty"`
	IsSandbox      bool   `json:"is_sandbox"`
	UnsubscribeAt  int64  `json:"unsubscribe_detected_at,omitempty"`
}

// Purchase is a non-renewing one-shot purchase.
type Purchase struct {
	ID         string `json:"id"`
	ProductID  string `json:"product_id"`
	Store      string `json:"store"`
	PurchaseAt int64  `json:"purchased_at"`
	IsSandbox  bool   `json:"is_sandbox"`
}

// Alias is one of the alternate ids merged into a customer.
type Alias struct {
	Alias string `json:"alias"`
	Type  string `json:"type"`
}

type listResp[T any] struct {
	Items []T    `json:"items"`
	Next  string `json:"next_page,omitempty"`
}

// requireProject is used by every project-scoped call.
func (c *Client) requireProject() error {
	if c.projectID == "" {
		return errors.New("no project_id on profile - run `revcat auth login` and pick a project")
	}
	return nil
}

func (c *Client) projectPath(suffix string) string {
	return "/projects/" + c.projectID + suffix
}

// GetCustomer fetches the base customer record.
func (c *Client) GetCustomer(ctx context.Context, customerID string) (*Customer, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	var out Customer
	path := c.projectPath("/customers/" + url.PathEscape(customerID))
	if err := c.Do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListActiveEntitlements pages through a customer's active entitlements.
func (c *Client) ListActiveEntitlements(ctx context.Context, customerID string) ([]ActiveEntitlement, error) {
	return paginate[ActiveEntitlement](ctx, c, c.projectPath("/customers/"+url.PathEscape(customerID)+"/active_entitlements"))
}

// ListSubscriptions pages through a customer's subscriptions (active +
// recently ended).
func (c *Client) ListSubscriptions(ctx context.Context, customerID string) ([]Subscription, error) {
	return paginate[Subscription](ctx, c, c.projectPath("/customers/"+url.PathEscape(customerID)+"/subscriptions"))
}

// ListPurchases pages through a customer's non-renewing purchases.
func (c *Client) ListPurchases(ctx context.Context, customerID string) ([]Purchase, error) {
	return paginate[Purchase](ctx, c, c.projectPath("/customers/"+url.PathEscape(customerID)+"/purchases"))
}

// ListAliases pages through a customer's aliases.
func (c *Client) ListAliases(ctx context.Context, customerID string) ([]Alias, error) {
	return paginate[Alias](ctx, c, c.projectPath("/customers/"+url.PathEscape(customerID)+"/aliases"))
}

// CustomerSnapshot is the assembled view rendered by `revcat subscribers info`.
// Each field is filled by a parallel API call; partial failure is tolerated
// so a missing endpoint doesn't blank the whole report.
type CustomerSnapshot struct {
	Customer      *Customer            `json:"customer,omitempty"`
	Entitlements  []ActiveEntitlement  `json:"active_entitlements,omitempty"`
	Subscriptions []Subscription       `json:"subscriptions,omitempty"`
	Purchases     []Purchase           `json:"purchases,omitempty"`
	Aliases       []Alias              `json:"aliases,omitempty"`
	Errors        map[string]string    `json:"errors,omitempty"`
}

// SnapshotCustomer fans out the per-customer endpoints in parallel and
// assembles a CustomerSnapshot. Errors per-section are collected so a
// missing endpoint or 404 sub-resource doesn't kill the whole render.
func (c *Client) SnapshotCustomer(ctx context.Context, customerID string) (*CustomerSnapshot, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}

	snap := &CustomerSnapshot{Errors: map[string]string{}}
	var mu sync.Mutex
	var wg sync.WaitGroup

	collect := func(section string, fn func() error) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := fn(); err != nil {
				mu.Lock()
				snap.Errors[section] = err.Error()
				mu.Unlock()
			}
		}()
	}

	collect("customer", func() error {
		v, err := c.GetCustomer(ctx, customerID)
		if err != nil {
			return err
		}
		snap.Customer = v
		return nil
	})
	collect("entitlements", func() error {
		v, err := c.ListActiveEntitlements(ctx, customerID)
		if err != nil {
			return err
		}
		snap.Entitlements = v
		return nil
	})
	collect("subscriptions", func() error {
		v, err := c.ListSubscriptions(ctx, customerID)
		if err != nil {
			return err
		}
		snap.Subscriptions = v
		return nil
	})
	collect("purchases", func() error {
		v, err := c.ListPurchases(ctx, customerID)
		if err != nil {
			return err
		}
		snap.Purchases = v
		return nil
	})
	collect("aliases", func() error {
		v, err := c.ListAliases(ctx, customerID)
		if err != nil {
			return err
		}
		snap.Aliases = v
		return nil
	})

	wg.Wait()

	// If the base customer call 404'd, treat as a hard error so the command
	// doesn't print an empty card. Other partial failures are non-fatal.
	if snap.Customer == nil {
		if msg, ok := snap.Errors["customer"]; ok {
			if strings.Contains(msg, "404") {
				return nil, fmt.Errorf("no customer with id %q in this project", customerID)
			}
			return nil, fmt.Errorf("fetch customer: %s", msg)
		}
		return nil, fmt.Errorf("fetch customer: unknown error")
	}
	if len(snap.Errors) == 0 {
		snap.Errors = nil
	}
	return snap, nil
}

// paginate is a generic helper for v2's cursor-paginated list endpoints.
func paginate[T any](ctx context.Context, c *Client, basePath string) ([]T, error) {
	var all []T
	path := basePath + "?limit=100"
	for {
		var page listResp[T]
		if err := c.Do(ctx, "GET", path, nil, &page); err != nil {
			return nil, err
		}
		all = append(all, page.Items...)
		if page.Next == "" {
			return all, nil
		}
		path = basePath + "?limit=100&starting_after=" + page.Next
	}
}
