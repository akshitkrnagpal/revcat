package packages

import (
	"context"
	"errors"
	"testing"

	"github.com/akshitkrnagpal/revcat/internal/api"
)

// TestEnrichProductsFillsDisplayName covers the case A scenario: the v2
// package-products endpoint returns binding metadata with an empty
// display_name. Enrichment must populate it from GetProduct.
func TestEnrichProductsFillsDisplayName(t *testing.T) {
	in := []api.Product{
		{ID: "prod_a", StoreIdentifier: "monthly.sub"},
	}
	get := func(_ context.Context, id string) (*api.Product, error) {
		if id != "prod_a" {
			t.Fatalf("unexpected product fetch: %s", id)
		}
		return &api.Product{ID: id, DisplayName: "Premium Monthly", AppID: "app_x"}, nil
	}
	out, err := enrichProducts(context.Background(), get, in)
	if err != nil {
		t.Fatalf("enrichProducts: %v", err)
	}
	if got := out[0].DisplayName; got != "Premium Monthly" {
		t.Fatalf("display_name: got %q want %q", got, "Premium Monthly")
	}
	if got := out[0].StoreIdentifier; got != "monthly.sub" {
		t.Fatalf("store_identifier: got %q want %q", got, "monthly.sub")
	}
	if got := out[0].AppID; got != "app_x" {
		t.Fatalf("app_id: got %q want %q", got, "app_x")
	}
}

// TestEnrichProductsCachesByID asserts that a product appearing twice
// (e.g. attached with different eligibility criteria) is only fetched
// once. Without caching, the package-products endpoint fan-out is N+1
// against the catalog size.
func TestEnrichProductsCachesByID(t *testing.T) {
	in := []api.Product{
		{ID: "prod_a"},
		{ID: "prod_a"},
		{ID: "prod_b"},
	}
	calls := map[string]int{}
	get := func(_ context.Context, id string) (*api.Product, error) {
		calls[id]++
		return &api.Product{ID: id, DisplayName: id + "-name"}, nil
	}
	out, err := enrichProducts(context.Background(), get, in)
	if err != nil {
		t.Fatalf("enrichProducts: %v", err)
	}
	if calls["prod_a"] != 1 {
		t.Fatalf("prod_a fetched %d times, want 1", calls["prod_a"])
	}
	if calls["prod_b"] != 1 {
		t.Fatalf("prod_b fetched %d times, want 1", calls["prod_b"])
	}
	for i, p := range out {
		if p.DisplayName == "" {
			t.Fatalf("row %d: display_name empty", i)
		}
	}
}

// TestEnrichProductsSkipsWhenAlreadySet keeps the helper a no-op when
// the v2 endpoint does happen to return display_name (case B). No
// network calls should fire in that case.
func TestEnrichProductsSkipsWhenAlreadySet(t *testing.T) {
	in := []api.Product{
		{ID: "prod_a", DisplayName: "Already Set"},
	}
	get := func(_ context.Context, id string) (*api.Product, error) {
		t.Fatalf("unexpected fetch for %s when display_name was already set", id)
		return nil, nil
	}
	out, err := enrichProducts(context.Background(), get, in)
	if err != nil {
		t.Fatalf("enrichProducts: %v", err)
	}
	if out[0].DisplayName != "Already Set" {
		t.Fatalf("display_name overwritten: got %q", out[0].DisplayName)
	}
}

// TestEnrichProductsPropagatesError ensures fetch failures bubble up so
// the command can decide whether to surface them or fall back.
func TestEnrichProductsPropagatesError(t *testing.T) {
	in := []api.Product{{ID: "prod_a"}}
	want := errors.New("boom")
	get := func(_ context.Context, _ string) (*api.Product, error) {
		return nil, want
	}
	_, err := enrichProducts(context.Background(), get, in)
	if !errors.Is(err, want) {
		t.Fatalf("got %v, want %v", err, want)
	}
}
