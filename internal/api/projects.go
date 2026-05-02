package api

import (
	"context"
	"encoding/json"
	"net/url"
)

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

// CreateProject creates a new project at the account level. Account-
// scoped, not project-scoped, so this does NOT call requireProject().
//
// v2: POST /v2/projects, scope project_configuration:projects:read_write.
// Body: {"name": "..."}. Returns the created Project.
func (c *Client) CreateProject(ctx context.Context, name string) (*Project, error) {
	var out Project
	body := map[string]any{"name": name}
	if err := c.Do(ctx, "POST", "/projects", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetProject fetches a single project by id. v2 has no GET /projects/{id}
// endpoint, so we list and filter client-side.
func (c *Client) GetProject(ctx context.Context, id string) (*Project, error) {
	all, err := c.ListProjects(ctx)
	if err != nil {
		return nil, err
	}
	for i := range all {
		if all[i].ID == id {
			return &all[i], nil
		}
	}
	return nil, &APIError{Status: 404, StatusText: "Not Found", Message: "no project with id " + id + " accessible to this key"}
}

// GetProjectRaw fetches a project and returns the verbatim list-item bytes
// alongside the typed projection. v2 has no per-id project endpoint, so
// this iterates list pages and matches the wanted id - the raw bytes are
// the matching item exactly as v2 returned it.
func (c *Client) GetProjectRaw(ctx context.Context, id string) (*Project, json.RawMessage, error) {
	path := "/projects?limit=100"
	for {
		var page struct {
			Items []json.RawMessage `json:"items"`
			Next  string            `json:"next_page,omitempty"`
		}
		if err := c.Do(ctx, "GET", path, nil, &page); err != nil {
			return nil, nil, err
		}
		for _, raw := range page.Items {
			var p Project
			if err := json.Unmarshal(raw, &p); err != nil {
				return nil, nil, err
			}
			if p.ID == id {
				return &p, raw, nil
			}
		}
		if page.Next == "" {
			return nil, nil, &APIError{Status: 404, StatusText: "Not Found", Message: "no project with id " + id + " accessible to this key"}
		}
		path = "/projects?limit=100&starting_after=" + page.Next
	}
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
	if err := c.Do(ctx, "GET", c.projectPath("/apps/"+url.PathEscape(appID)), nil, &a); err != nil {
		return nil, err
	}
	return &a, nil
}

// GetAppRaw fetches an app and returns the verbatim v2 response.
func (c *Client) GetAppRaw(ctx context.Context, appID string) (*App, json.RawMessage, error) {
	if err := c.requireProject(); err != nil {
		return nil, nil, err
	}
	var a App
	raw, err := c.DoRaw(ctx, "GET", c.projectPath("/apps/"+url.PathEscape(appID)), nil, &a)
	if err != nil {
		return nil, nil, err
	}
	return &a, raw, nil
}

// ListPublicAPIKeys returns the SDK-side keys for an app.
func (c *Client) ListPublicAPIKeys(ctx context.Context, appID string) ([]PublicAPIKey, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	return paginate[PublicAPIKey](ctx, c, c.projectPath("/apps/"+url.PathEscape(appID)+"/public_api_keys"))
}

// GetStoreKitConfig returns the StoreKit configuration for an app (iOS).
// Returned as a generic map because the schema is broad and changes often.
func (c *Client) GetStoreKitConfig(ctx context.Context, appID string) (map[string]any, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	var out map[string]any
	if err := c.Do(ctx, "GET", c.projectPath("/apps/"+url.PathEscape(appID)+"/store_kit_config"), nil, &out); err != nil {
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

// Collaborator is a member of a RevenueCat project.
//
// `name` is nullable on the v2 API (returned as null for invitees who
// haven't completed signup). `accepted_at` is also nullable - null
// means a pending invite. `role` is free-form per the v2 spec; common
// values include "admin", "developer", "billing", "viewer".
type Collaborator struct {
	ID         string `json:"id"`
	Name       string `json:"name,omitempty"`
	Email      string `json:"email"`
	Role       string `json:"role,omitempty"`
	AcceptedAt int64  `json:"accepted_at,omitempty"`
	HasMFA     bool   `json:"has_mfa"`
}

// ListCollaborators pages through every collaborator on the active
// project. v2: GET /v2/projects/{project_id}/collaborators,
// scope project_configuration:collaborators:read.
func (c *Client) ListCollaborators(ctx context.Context) ([]Collaborator, error) {
	if err := c.requireProject(); err != nil {
		return nil, err
	}
	return paginate[Collaborator](ctx, c, c.projectPath("/collaborators"))
}
