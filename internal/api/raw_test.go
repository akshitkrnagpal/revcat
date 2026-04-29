package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestGetProductRaw_PreservesAllV2Fields is the regression test for issue
// #4. The typed Product struct only models a handful of fields, but
// `revcat products view --output json` must surface every key the v2 API
// returned. GetProductRaw promises to return the verbatim response.
func TestGetProductRaw_PreservesAllV2Fields(t *testing.T) {
	v2Body := `{
		"id": "prod8af0c2ae8c",
		"store_identifier": "app.monthly",
		"type": "subscription",
		"display_name": "Monthly",
		"app_id": "app_abc",
		"created_at": 1700000000,
		"future_field_revcat_doesnt_model": "still here",
		"nested": {"deep": {"keep": true}}
	}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/products/prod8af0c2ae8c") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(v2Body))
	}))
	defer srv.Close()

	c := New(Options{
		SecretKey: "sk_test",
		ProjectID: "proj_xyz",
		BaseURL:   srv.URL,
	})

	p, raw, err := c.GetProductRaw(context.Background(), "prod8af0c2ae8c")
	if err != nil {
		t.Fatalf("GetProductRaw: %v", err)
	}

	// Typed projection still works for table render.
	if p.ID != "prod8af0c2ae8c" || p.AppID != "app_abc" {
		t.Errorf("typed projection lost fields: %+v", p)
	}

	// Raw bytes must round-trip into a map and contain every v2 key,
	// including ones the typed struct doesn't model.
	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("decode raw: %v", err)
	}
	for _, k := range []string{
		"id", "store_identifier", "type", "display_name",
		"app_id", "created_at",
		"future_field_revcat_doesnt_model", "nested",
	} {
		if _, ok := got[k]; !ok {
			t.Errorf("raw response dropped key %q (issue #4 regression)", k)
		}
	}
}

// TestDoRaw_ReturnsBodyOnSuccess validates the underlying primitive used
// by every Get*Raw helper.
func TestDoRaw_ReturnsBodyOnSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"a":1,"b":"two"}`))
	}))
	defer srv.Close()

	c := New(Options{SecretKey: "sk", BaseURL: srv.URL})

	var dst struct {
		A int `json:"a"`
	}
	raw, err := c.DoRaw(context.Background(), "GET", "/anything", nil, &dst)
	if err != nil {
		t.Fatalf("DoRaw: %v", err)
	}
	if dst.A != 1 {
		t.Errorf("typed decode failed: %+v", dst)
	}
	if string(raw) != `{"a":1,"b":"two"}` {
		t.Errorf("raw body mismatch: %s", raw)
	}
}

// TestCustomerSnapshot_MarshalJSON_PassesRawThrough exercises the
// MarshalJSON override that keeps subscriber snapshots field-complete.
func TestCustomerSnapshot_MarshalJSON_PassesRawThrough(t *testing.T) {
	snap := &CustomerSnapshot{
		RawCustomer:     json.RawMessage(`{"id":"u1","extra":"keep"}`),
		RawEntitlements: []json.RawMessage{json.RawMessage(`{"id":"e1","new_field":42}`)},
	}
	out, err := json.Marshal(snap)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(out), `"extra":"keep"`) {
		t.Errorf("customer raw passthrough lost: %s", out)
	}
	if !strings.Contains(string(out), `"new_field":42`) {
		t.Errorf("entitlements raw passthrough lost: %s", out)
	}
}
