package packages

import (
	"encoding/json"
	"testing"

	"github.com/akshitkrnagpal/revcat/internal/api"
)

// TestPackageProductBindingShape locks in that v2's package-products
// response is parsed as `{eligibility_criteria, product:{...}}` rather
// than a bare Product. This was the bug behind the empty-display_name
// symptom that motivated this PR.
func TestPackageProductBindingShape(t *testing.T) {
	// Verbatim shape we observed from the live API on 2026-04-29:
	//   GET /v2/projects/{p}/packages/{pkg}/products
	// →  {"items":[{"eligibility_criteria":"all","product":{...}}], ...}
	raw := []byte(`{
		"eligibility_criteria": "all",
		"product": {
			"app_id": "appbc40da51e8",
			"created_at": 1777154411736,
			"display_name": "Monthly v2",
			"id": "prodbfa7f763ce",
			"object": "product",
			"state": "active",
			"store_identifier": "com.revcat.test.monthly",
			"type": "subscription"
		}
	}`)

	var got api.PackageProductBinding
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal binding: %v", err)
	}

	if got.EligibilityCriteria != "all" {
		t.Fatalf("eligibility: got %q want all", got.EligibilityCriteria)
	}
	if got.Product.ID != "prodbfa7f763ce" {
		t.Fatalf("product.id: got %q", got.Product.ID)
	}
	if got.Product.DisplayName != "Monthly v2" {
		t.Fatalf("product.display_name: got %q want Monthly v2 (was empty before this fix)", got.Product.DisplayName)
	}
	if got.Product.StoreIdentifier != "com.revcat.test.monthly" {
		t.Fatalf("product.store_identifier: got %q (was empty before this fix)", got.Product.StoreIdentifier)
	}
	if got.Product.AppID != "appbc40da51e8" {
		t.Fatalf("product.app_id: got %q (was empty before this fix)", got.Product.AppID)
	}
}

// TestPackageProductBindingPreservesUnknownFields confirms a future
// field added to the binding object (e.g. a new eligibility category)
// would still survive a JSON round-trip via the Raw passthrough path.
// The typed struct doesn't model it, but `--output json` should.
func TestPackageProductBindingPreservesUnknownFields(t *testing.T) {
	raw := json.RawMessage(`{
		"eligibility_criteria": "all",
		"future_field_revcat_doesnt_model": "still_here",
		"product": {"id": "prod_a", "display_name": "X"}
	}`)

	// Round-trip through json.RawMessage (what ListPackageProductsRaw
	// hands the command layer) keeps the bytes verbatim.
	out, err := json.Marshal(raw)
	if err != nil {
		t.Fatalf("marshal raw: %v", err)
	}
	if !contains(string(out), "future_field_revcat_doesnt_model") {
		t.Fatalf("unknown field dropped during raw round-trip: %s", out)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > 0 && (indexOf(s, substr) >= 0)))
}

func indexOf(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
