package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"
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

// GetCustomerRaw fetches the customer record and returns the verbatim v2
// response alongside the typed projection.
func (c *Client) GetCustomerRaw(ctx context.Context, customerID string) (*Customer, json.RawMessage, error) {
	if err := c.requireProject(); err != nil {
		return nil, nil, err
	}
	var out Customer
	path := c.projectPath("/customers/" + url.PathEscape(customerID))
	raw, err := c.DoRaw(ctx, "GET", path, nil, &out)
	if err != nil {
		return nil, nil, err
	}
	return &out, raw, nil
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
//
// The Raw* fields hold the verbatim v2 bytes for each section so JSON
// output can pass through every field RC returns rather than the curated
// typed projection. They marshal under the same JSON keys as the typed
// fields when present (via MarshalJSON) so consumers see one shape.
type CustomerSnapshot struct {
	Customer      *Customer            `json:"-"`
	Entitlements  []ActiveEntitlement  `json:"-"`
	Subscriptions []Subscription       `json:"-"`
	Purchases     []Purchase           `json:"-"`
	Aliases       []Alias              `json:"-"`
	Errors        map[string]string    `json:"errors,omitempty"`

	// Raw* hold the verbatim v2 bytes per section; populated on success.
	RawCustomer      json.RawMessage   `json:"-"`
	RawEntitlements  []json.RawMessage `json:"-"`
	RawSubscriptions []json.RawMessage `json:"-"`
	RawPurchases     []json.RawMessage `json:"-"`
	RawAliases       []json.RawMessage `json:"-"`
}

// MarshalJSON renders the snapshot with the raw v2 field set when
// available, falling back to the typed projection otherwise. This keeps
// `revcat subscribers info --output json` field-complete relative to the
// underlying v2 response (issue #4).
func (s *CustomerSnapshot) MarshalJSON() ([]byte, error) {
	type rawArr = []json.RawMessage
	out := struct {
		Customer      json.RawMessage `json:"customer,omitempty"`
		Entitlements  rawArr          `json:"active_entitlements,omitempty"`
		Subscriptions rawArr          `json:"subscriptions,omitempty"`
		Purchases     rawArr          `json:"purchases,omitempty"`
		Aliases       rawArr          `json:"aliases,omitempty"`
		Errors        map[string]string `json:"errors,omitempty"`
	}{Errors: s.Errors}

	if s.RawCustomer != nil {
		out.Customer = s.RawCustomer
	} else if s.Customer != nil {
		b, err := json.Marshal(s.Customer)
		if err != nil {
			return nil, err
		}
		out.Customer = b
	}

	if s.RawEntitlements != nil {
		out.Entitlements = s.RawEntitlements
	} else if s.Entitlements != nil {
		out.Entitlements = make(rawArr, 0, len(s.Entitlements))
		for i := range s.Entitlements {
			b, err := json.Marshal(s.Entitlements[i])
			if err != nil {
				return nil, err
			}
			out.Entitlements = append(out.Entitlements, b)
		}
	}

	if s.RawSubscriptions != nil {
		out.Subscriptions = s.RawSubscriptions
	} else if s.Subscriptions != nil {
		out.Subscriptions = make(rawArr, 0, len(s.Subscriptions))
		for i := range s.Subscriptions {
			b, err := json.Marshal(s.Subscriptions[i])
			if err != nil {
				return nil, err
			}
			out.Subscriptions = append(out.Subscriptions, b)
		}
	}

	if s.RawPurchases != nil {
		out.Purchases = s.RawPurchases
	} else if s.Purchases != nil {
		out.Purchases = make(rawArr, 0, len(s.Purchases))
		for i := range s.Purchases {
			b, err := json.Marshal(s.Purchases[i])
			if err != nil {
				return nil, err
			}
			out.Purchases = append(out.Purchases, b)
		}
	}

	if s.RawAliases != nil {
		out.Aliases = s.RawAliases
	} else if s.Aliases != nil {
		out.Aliases = make(rawArr, 0, len(s.Aliases))
		for i := range s.Aliases {
			b, err := json.Marshal(s.Aliases[i])
			if err != nil {
				return nil, err
			}
			out.Aliases = append(out.Aliases, b)
		}
	}

	return json.Marshal(out)
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
		v, raw, err := c.GetCustomerRaw(ctx, customerID)
		if err != nil {
			return err
		}
		snap.Customer = v
		snap.RawCustomer = raw
		return nil
	})
	collect("entitlements", func() error {
		typed, raw, err := paginateBoth[ActiveEntitlement](ctx, c, c.projectPath("/customers/"+url.PathEscape(customerID)+"/active_entitlements"))
		if err != nil {
			return err
		}
		snap.Entitlements = typed
		snap.RawEntitlements = raw
		return nil
	})
	collect("subscriptions", func() error {
		typed, raw, err := paginateBoth[Subscription](ctx, c, c.projectPath("/customers/"+url.PathEscape(customerID)+"/subscriptions"))
		if err != nil {
			return err
		}
		snap.Subscriptions = typed
		snap.RawSubscriptions = raw
		return nil
	})
	collect("purchases", func() error {
		typed, raw, err := paginateBoth[Purchase](ctx, c, c.projectPath("/customers/"+url.PathEscape(customerID)+"/purchases"))
		if err != nil {
			return err
		}
		snap.Purchases = typed
		snap.RawPurchases = raw
		return nil
	})
	collect("aliases", func() error {
		typed, raw, err := paginateBoth[Alias](ctx, c, c.projectPath("/customers/"+url.PathEscape(customerID)+"/aliases"))
		if err != nil {
			return err
		}
		snap.Aliases = typed
		snap.RawAliases = raw
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

// GrantEntitlementRequest is the body for the v2 grant action. RC v2
// expects an absolute expiry timestamp (milliseconds since Unix epoch),
// not a duration. Use "forever" semantics by passing a far-future value.
type GrantEntitlementRequest struct {
	EntitlementID string `json:"entitlement_id"`
	ExpiresAt     int64  `json:"expires_at"`
}

// GrantEntitlement grants a promotional entitlement to a customer.
// Wraps POST /customers/{id}/actions/grant_entitlement. Returns the
// updated customer record (active_entitlements + ids) which we surface
// raw because the response shape is broad and changes.
func (c *Client) GrantEntitlement(ctx context.Context, customerID string, req GrantEntitlementRequest) (map[string]any, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	var out map[string]any
	path := c.projectPath("/customers/" + url.PathEscape(customerID) + "/actions/grant_entitlement")
	if err := c.Do(ctx, "POST", path, req, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// RevokeEntitlement is implemented as "grant with a near-future expiry"
// because v2 has no first-class revoke endpoint and rejects past
// expires_at values ("must be in the future"). We pick now+1s, which
// expires the grant within a second of the call. RC's customer info
// will reflect the entitlement as expired on the next read.
func (c *Client) RevokeEntitlement(ctx context.Context, customerID, entitlementID string) error {
	if err := c.requireProject(); err != nil {
		return err
	}
	body := GrantEntitlementRequest{
		EntitlementID: entitlementID,
		ExpiresAt:     timeNowMillis() + 1000,
	}
	path := c.projectPath("/customers/" + url.PathEscape(customerID) + "/actions/grant_entitlement")
	return c.Do(ctx, "POST", path, body, nil)
}

// timeNowMillis returns current Unix time in milliseconds. Wrapped so
// tests can stub.
func timeNowMillis() int64 { return timeNowMillisFn() }

var timeNowMillisFn = func() int64 {
	return time.Now().UnixMilli()
}

// RefundTransaction issues a refund through the appropriate store.
// Wraps POST /subscriptions/{sid}/transactions/{tid}/actions/refund.
// The subscription id is required because v2 scopes refunds under the
// subscription, not the customer.
func (c *Client) RefundTransaction(ctx context.Context, subscriptionID, transactionID string) (map[string]any, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	var out map[string]any
	path := c.projectPath("/subscriptions/" + url.PathEscape(subscriptionID) + "/transactions/" + url.PathEscape(transactionID) + "/actions/refund")
	if err := c.Do(ctx, "POST", path, struct{}{}, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ListCustomers pages through customers in the active project. Useful for
// admin/audit; for support workflows search by store id is usually faster.
func (c *Client) ListCustomers(ctx context.Context) ([]Customer, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	return paginate[Customer](ctx, c, c.projectPath("/customers"))
}

// CreateCustomer pre-creates a customer record with optional attributes.
// Most apps let the SDK create customers on first launch; this is for
// migration/imports.
func (c *Client) CreateCustomer(ctx context.Context, body map[string]any) (*Customer, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	var out Customer
	if err := c.Do(ctx, "POST", c.projectPath("/customers"), body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteCustomer permanently removes a customer (GDPR/test cleanup).
func (c *Client) DeleteCustomer(ctx context.Context, customerID string) error {
	if err := c.requireProject(); err != nil {
		return err
	}
	return c.Do(ctx, "DELETE", c.projectPath("/customers/"+url.PathEscape(customerID)), nil, nil)
}

// TransferCustomer moves entitlements/subscriptions from src to dst.
func (c *Client) TransferCustomer(ctx context.Context, srcID, dstID string) error {
	if err := c.requireProject(); err != nil {
		return err
	}
	body := map[string]any{"target_customer_id": dstID}
	path := c.projectPath("/customers/" + url.PathEscape(srcID) + "/actions/transfer")
	return c.Do(ctx, "POST", path, body, nil)
}

// CustomerAttribute is one entry in the v2 attribute array.
type CustomerAttribute struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// GetAttributes returns subscriber attributes (paged list of name/value).
func (c *Client) GetAttributes(ctx context.Context, customerID string) ([]CustomerAttribute, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	return paginate[CustomerAttribute](ctx, c, c.projectPath("/customers/"+url.PathEscape(customerID)+"/attributes"))
}

// SetAttributes upserts subscriber attributes. v2 expects the body to
// be {"attributes": [{"name": "...", "value": "..."}, ...]}.
func (c *Client) SetAttributes(ctx context.Context, customerID string, attrs []CustomerAttribute) error {
	if err := c.requireProject(); err != nil {
		return err
	}
	body := map[string]any{"attributes": attrs}
	path := c.projectPath("/customers/" + url.PathEscape(customerID) + "/attributes")
	return c.Do(ctx, "POST", path, body, nil)
}

// ListInvoices pages a customer's invoices.
func (c *Client) ListInvoices(ctx context.Context, customerID string) ([]map[string]any, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	return paginate[map[string]any](ctx, c, c.projectPath("/customers/"+url.PathEscape(customerID)+"/invoices"))
}

// Note: v2 has no customer-scoped endpoints for the following actions.
// They lived in v1 but the v2 surface either moved them or removed
// them entirely. Smoke-tested 2026-04-25 against blank project; all
// 404. Removed from the CLI rather than ship broken commands:
//   - override_offering / set_offering_override
//   - restore_google_play_purchase
//   - virtual_currencies/balances (per-customer balance read)
//   - virtual_currencies/transactions (per-customer credit/debit)
//   - virtual_currencies_balances (per-customer set balance)

// paginate is a generic helper for v2's cursor-paginated list endpoints.
func paginate[T any](ctx context.Context, c *Client, basePath string) ([]T, error) {
	all := []T{} // never nil; "no items" should JSON-encode as [] not null
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

// paginateBoth is paginate with raw item bytes preserved. The typed slice
// is for table/snapshot rendering, the raw slice keeps every v2 field.
func paginateBoth[T any](ctx context.Context, c *Client, basePath string) ([]T, []json.RawMessage, error) {
	typed := []T{}
	rawAll := []json.RawMessage{}
	path := basePath + "?limit=100"
	for {
		var page struct {
			Items []json.RawMessage `json:"items"`
			Next  string            `json:"next_page,omitempty"`
		}
		if err := c.Do(ctx, "GET", path, nil, &page); err != nil {
			return nil, nil, err
		}
		for _, item := range page.Items {
			var v T
			if err := json.Unmarshal(item, &v); err != nil {
				return nil, nil, fmt.Errorf("decode list item: %w", err)
			}
			typed = append(typed, v)
			rawAll = append(rawAll, item)
		}
		if page.Next == "" {
			return typed, rawAll, nil
		}
		path = basePath + "?limit=100&starting_after=" + page.Next
	}
}
