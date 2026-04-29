package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestPreviewOfferings_RequestShape(t *testing.T) {
	cases := []struct {
		name         string
		userID       string
		platform     string
		wantPath     string
		wantPlatform string
	}{
		{
			name:         "ios simple user id",
			userID:       "revcat_preview_123",
			platform:     "iOS",
			wantPath:     "/subscribers/revcat_preview_123/offerings",
			wantPlatform: "iOS",
		},
		{
			name:         "android with $ in user id is escaped",
			userID:       "user/with spaces",
			platform:     "android",
			wantPath:     "/subscribers/user%2Fwith%20spaces/offerings",
			wantPlatform: "android",
		},
		{
			name:         "web platform",
			userID:       "abc",
			platform:     "web",
			wantPath:     "/subscribers/abc/offerings",
			wantPlatform: "web",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var gotPath, gotAuth, gotPlatform string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				gotAuth = r.Header.Get("Authorization")
				gotPlatform = r.Header.Get("X-Platform")
				_ = json.NewEncoder(w).Encode(PreviewOfferingsResponse{
					CurrentOfferingID: "default",
					Offerings: []PreviewOffering{
						{Identifier: "default", Packages: []PreviewPackage{
							{Identifier: "$rc_monthly", PlatformProductIdentifier: "rc_pro_monthly"},
						}},
					},
				})
			}))
			defer srv.Close()
			t.Setenv("REVCAT_V1_BASE_URL", srv.URL)

			c := New(Options{SecretKey: "ignored", Version: "test"})
			resp, err := c.PreviewOfferings(context.Background(), tc.userID, "pk_test_abc123", tc.platform)
			if err != nil {
				t.Fatalf("PreviewOfferings: %v", err)
			}
			// Compare decoded path: httptest already decodes r.URL.Path, so compare against decoded form.
			decodedWant, _ := url.PathUnescape(tc.wantPath)
			if gotPath != decodedWant {
				t.Errorf("path: want %q, got %q", decodedWant, gotPath)
			}
			if gotAuth != "Bearer pk_test_abc123" {
				t.Errorf("auth: want Bearer pk_test_abc123, got %q", gotAuth)
			}
			if gotPlatform != tc.wantPlatform {
				t.Errorf("X-Platform: want %q, got %q", tc.wantPlatform, gotPlatform)
			}
			if resp.CurrentOfferingID != "default" {
				t.Errorf("current_offering_id: want default, got %q", resp.CurrentOfferingID)
			}
			if len(resp.Offerings) != 1 || len(resp.Offerings[0].Packages) != 1 {
				t.Fatalf("unexpected response shape: %+v", resp)
			}
			if resp.Offerings[0].Packages[0].Identifier != "$rc_monthly" {
				t.Errorf("package id: want $rc_monthly, got %q", resp.Offerings[0].Packages[0].Identifier)
			}
		})
	}
}

func TestPreviewOfferings_Validation(t *testing.T) {
	c := New(Options{SecretKey: "k", Version: "test"})
	if _, err := c.PreviewOfferings(context.Background(), "", "pk", "ios"); err == nil {
		t.Error("want error for empty user id")
	}
	if _, err := c.PreviewOfferings(context.Background(), "u", "", "ios"); err == nil {
		t.Error("want error for empty key")
	}
	if _, err := c.PreviewOfferings(context.Background(), "u", "pk", ""); err == nil {
		t.Error("want error for empty platform")
	}
}

func TestPreviewOfferings_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"invalid api key"}`))
	}))
	defer srv.Close()
	t.Setenv("REVCAT_V1_BASE_URL", srv.URL)

	c := New(Options{SecretKey: "k", Version: "test"})
	_, err := c.PreviewOfferings(context.Background(), "u", "bad_key", "iOS")
	if err == nil {
		t.Fatal("want error for 401")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("want APIError, got %T: %v", err, err)
	}
	if apiErr.Status != http.StatusUnauthorized {
		t.Errorf("status: want 401, got %d", apiErr.Status)
	}
	if !strings.Contains(apiErr.Message, "invalid api key") {
		t.Errorf("message: want to contain 'invalid api key', got %q", apiErr.Message)
	}
}
