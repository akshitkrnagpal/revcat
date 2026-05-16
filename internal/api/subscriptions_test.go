package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSearchSubscriptionsPaginates(t *testing.T) {
	var starts []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/projects/proj_xyz/subscriptions/search" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("query"); got != "store_123" {
			t.Fatalf("query: want store_123, got %q", got)
		}
		if got := r.URL.Query().Get("limit"); got != "100" {
			t.Fatalf("limit: want 100, got %q", got)
		}

		startingAfter := r.URL.Query().Get("starting_after")
		starts = append(starts, startingAfter)
		w.Header().Set("Content-Type", "application/json")
		if startingAfter == "" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"items":     []map[string]any{{"id": "sub_1"}},
				"next_page": "cursor_1",
			})
			return
		}
		if startingAfter != "cursor_1" {
			t.Fatalf("starting_after: want cursor_1, got %q", startingAfter)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{{"id": "sub_2"}},
		})
	}))
	defer srv.Close()

	c := New(Options{TokenSource: stubToken("sk"), ProjectID: "proj_xyz", BaseURL: srv.URL})
	got, err := c.SearchSubscriptions(context.Background(), "store_123")
	if err != nil {
		t.Fatalf("SearchSubscriptions: %v", err)
	}
	if len(got) != 2 || got[0].ID != "sub_1" || got[1].ID != "sub_2" {
		t.Fatalf("unexpected subscriptions: %+v", got)
	}
	if len(starts) != 2 || starts[0] != "" || starts[1] != "cursor_1" {
		t.Fatalf("unexpected pagination cursors: %#v", starts)
	}
}

func TestSearchPurchasesPaginates(t *testing.T) {
	var starts []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/projects/proj_xyz/purchases/search" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("query"); got != "store_123" {
			t.Fatalf("query: want store_123, got %q", got)
		}

		startingAfter := r.URL.Query().Get("starting_after")
		starts = append(starts, startingAfter)
		w.Header().Set("Content-Type", "application/json")
		if startingAfter == "" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"items":     []map[string]any{{"id": "pur_1"}},
				"next_page": "cursor_1",
			})
			return
		}
		if startingAfter != "cursor_1" {
			t.Fatalf("starting_after: want cursor_1, got %q", startingAfter)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{{"id": "pur_2"}},
		})
	}))
	defer srv.Close()

	c := New(Options{TokenSource: stubToken("sk"), ProjectID: "proj_xyz", BaseURL: srv.URL})
	got, err := c.SearchPurchases(context.Background(), "store_123")
	if err != nil {
		t.Fatalf("SearchPurchases: %v", err)
	}
	if len(got) != 2 || got[0].ID != "pur_1" || got[1].ID != "pur_2" {
		t.Fatalf("unexpected purchases: %+v", got)
	}
	if len(starts) != 2 || starts[0] != "" || starts[1] != "cursor_1" {
		t.Fatalf("unexpected pagination cursors: %#v", starts)
	}
}
