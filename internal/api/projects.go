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
