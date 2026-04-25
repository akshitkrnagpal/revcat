package api

import "context"

// Project is the v2 representation of a RevenueCat project. Trimmed to the
// fields revcat surfaces; extend as commands need more.
type Project struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt int64  `json:"created_at"`
}

type projectListResponse struct {
	Items []Project `json:"items"`
	Next  string    `json:"next_page,omitempty"`
}

// ListProjects returns every project the secret key can access. The v2
// API paginates via next_page cursors; we collapse the pages here.
func (c *Client) ListProjects(ctx context.Context) ([]Project, error) {
	var all []Project
	path := "/projects?limit=100"
	for {
		var page projectListResponse
		if err := c.Do(ctx, "GET", path, nil, &page); err != nil {
			return nil, err
		}
		all = append(all, page.Items...)
		if page.Next == "" {
			return all, nil
		}
		path = "/projects?limit=100&starting_after=" + page.Next
	}
}

// GetProject fetches a single project by id.
func (c *Client) GetProject(ctx context.Context, id string) (*Project, error) {
	var p Project
	if err := c.Do(ctx, "GET", "/projects/"+id, nil, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// App is a per-platform app inside a project (one per bundle id / package).
type App struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	BundleID    string `json:"bundle_id,omitempty"`
	PackageName string `json:"package_name,omitempty"`
	CreatedAt   int64  `json:"created_at"`
	ProjectID   string `json:"project_id,omitempty"`
}

// PublicAPIKey is the SDK-side public key for an app.
type PublicAPIKey struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Key   string `json:"key"`
}

// ListApps pages through every app in the active project.
func (c *Client) ListApps(ctx context.Context) ([]App, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	return paginate[App](ctx, c, c.projectPath("/apps"))
}

// GetApp fetches one app by id.
func (c *Client) GetApp(ctx context.Context, appID string) (*App, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	var a App
	if err := c.Do(ctx, "GET", c.projectPath("/apps/"+appID), nil, &a); err != nil {
		return nil, err
	}
	return &a, nil
}

// ListPublicAPIKeys returns the SDK-side keys for an app.
func (c *Client) ListPublicAPIKeys(ctx context.Context, appID string) ([]PublicAPIKey, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	return paginate[PublicAPIKey](ctx, c, c.projectPath("/apps/"+appID+"/public_api_keys"))
}

// GetStoreKitConfig returns the StoreKit configuration for an app (iOS).
// Returned as a generic map because the schema is broad and changes often.
func (c *Client) GetStoreKitConfig(ctx context.Context, appID string) (map[string]any, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	var out map[string]any
	if err := c.Do(ctx, "GET", c.projectPath("/apps/"+appID+"/store_kit_config"), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// AuditLogEntry is a row from the project's audit log.
type AuditLogEntry struct {
	ID               string         `json:"id"`
	OccurredAt       int64          `json:"occurred_at"`
	ActionType       string         `json:"action_type"`
	ActorType        string         `json:"actor_type"`
	ActorIdentifier  string         `json:"actor_identifier"`
	TargetType       string         `json:"target_type"`
	TargetIdentifier string         `json:"target_identifier"`
	AdditionalData   map[string]any `json:"additional_data,omitempty"`
}

// ListAuditLogs pages every audit log entry the active key can see.
// Available with normal project secret keys (despite earlier docs that
// claimed partner-tier only).
func (c *Client) ListAuditLogs(ctx context.Context) ([]AuditLogEntry, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	return paginate[AuditLogEntry](ctx, c, c.projectPath("/audit_logs"))
}
