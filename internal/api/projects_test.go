package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestGetApp_EscapesAppID is the regression test for issue #16. The v2
// API path for /apps/{id} must percent-encode the app id so that ids
// containing path-significant characters (slashes, question marks) do
// not change which endpoint is hit.
func TestGetApp_EscapesAppID(t *testing.T) {
	const trickyID = "app/with?weird"
	// url.PathEscape leaves "/" alone but escapes "?". Confirm we hit a
	// path that no longer contains the raw "?" delimiter.
	var seenPath, seenRawPath, seenRawQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		seenRawPath = r.URL.EscapedPath()
		seenRawQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"app_abc","name":"x","type":"ios","created_at":0}`))
	}))
	defer srv.Close()

	c := New(Options{TokenSource: stubToken("sk"), ProjectID: "proj_xyz", BaseURL: srv.URL})

	if _, err := c.GetApp(context.Background(), trickyID); err != nil {
		t.Fatalf("GetApp: %v", err)
	}

	// "?" in the id must not have leaked into the query string.
	if seenRawQuery != "" {
		t.Errorf("query string should be empty, got %q (path=%q)", seenRawQuery, seenPath)
	}

	// Escaped path must contain the encoded "?" (%3F) inside the id.
	if !strings.Contains(seenRawPath, "%3F") {
		t.Errorf("expected %%3F in escaped path, got %q", seenRawPath)
	}
}
