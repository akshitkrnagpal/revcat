package api

import (
	"context"
	"net/url"
)

// GetMetricsOverview returns the project's headline metrics
// (active subs, MRR, lifetime revenue, etc.). Shape varies; surfaced raw.
func (c *Client) GetMetricsOverview(ctx context.Context) (map[string]any, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	var out map[string]any
	if err := c.Do(ctx, "GET", c.projectPath("/metrics/overview"), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetChart returns the data for a named chart
// (revenue, active_users, conversion, etc.).
type ChartOptions struct {
	StartDate string
	EndDate   string
	Period    string
	Filters   map[string]string
}

func (c *Client) GetChart(ctx context.Context, name string, opts ChartOptions) (map[string]any, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	q := url.Values{}
	if opts.StartDate != "" {
		q.Set("start_date", opts.StartDate)
	}
	if opts.EndDate != "" {
		q.Set("end_date", opts.EndDate)
	}
	if opts.Period != "" {
		q.Set("period", opts.Period)
	}
	for k, v := range opts.Filters {
		q.Set(k, v)
	}
	path := c.projectPath("/charts/" + url.PathEscape(name))
	if encoded := q.Encode(); encoded != "" {
		path += "?" + encoded
	}
	var out map[string]any
	if err := c.Do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetChartOptions returns the available filter/dimension options for a
// chart (so callers can validate before requesting data).
func (c *Client) GetChartOptions(ctx context.Context, name string) (map[string]any, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	var out map[string]any
	if err := c.Do(ctx, "GET", c.projectPath("/charts/"+url.PathEscape(name)+"/options"), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}
