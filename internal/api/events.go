package api

import (
	"context"
	"net/url"
	"strconv"
	"strings"
)

// Event is the v2 events API row. Fields kept loose because RC adds
// type-specific fields constantly; the raw JSON is preserved on Extra
// for renderers that want to surface uncommon fields.
type Event struct {
	ID         string         `json:"id"`
	Type       string         `json:"type"`
	OccurredAt int64          `json:"event_timestamp_ms"`
	AppUserID  string         `json:"app_user_id,omitempty"`
	ProductID  string         `json:"product_id,omitempty"`
	Store      string         `json:"store,omitempty"`
	Country    string         `json:"country_code,omitempty"`
	Currency   string         `json:"currency,omitempty"`
	Price      float64        `json:"price,omitempty"`
	IsSandbox  bool           `json:"is_sandbox,omitempty"`
	Extra      map[string]any `json:"-"`
}

// ListEventsOptions filters the events query. All optional.
type ListEventsOptions struct {
	// Types restricts to one or more event types (e.g., "INITIAL_PURCHASE").
	Types []string
	// SinceMS returns events at or after this Unix-ms timestamp.
	SinceMS int64
	// Cursor pagination cursor returned by a previous call.
	Cursor string
	// Limit caps the page size; RC default is 100.
	Limit int
}

// ListEventsPage is one page of events plus a cursor for the next.
type ListEventsPage struct {
	Items      []Event `json:"items"`
	NextCursor string  `json:"next_page,omitempty"`
}

// ListEvents returns one page of events. Callers polling for new events
// should use this and track NextCursor + the highest OccurredAt seen.
func (c *Client) ListEvents(ctx context.Context, opts ListEventsOptions) (*ListEventsPage, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	q := url.Values{}
	if opts.Limit > 0 {
		q.Set("limit", strconv.Itoa(opts.Limit))
	} else {
		q.Set("limit", "100")
	}
	if opts.SinceMS > 0 {
		q.Set("since", strconv.FormatInt(opts.SinceMS, 10))
	}
	if opts.Cursor != "" {
		q.Set("starting_after", opts.Cursor)
	}
	if len(opts.Types) > 0 {
		q.Set("type", strings.Join(opts.Types, ","))
	}
	path := c.projectPath("/events?" + q.Encode())
	var out ListEventsPage
	if err := c.Do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
