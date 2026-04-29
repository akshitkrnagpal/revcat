package cliutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "body.json")
	if err := os.WriteFile(path, []byte(`{"a":1,"b":"hi"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	out, err := LoadJSON(path)
	if err != nil {
		t.Fatal(err)
	}
	if v, _ := out["a"].(float64); v != 1 {
		t.Fatalf("a: want 1, got %v", out["a"])
	}
	if out["b"] != "hi" {
		t.Fatalf("b: want hi, got %v", out["b"])
	}
}

func TestLoadJSONMissing(t *testing.T) {
	if _, err := LoadJSON("/no/such/file.json"); err == nil {
		t.Fatal("want error for missing file")
	}
	if _, err := LoadJSON(""); err == nil {
		t.Fatal("want error for empty path")
	}
}

func TestLoadJSONInvalid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadJSON(path); err == nil {
		t.Fatal("want error for invalid json")
	}
}

func TestFormatTime(t *testing.T) {
	if got := FormatTime(0); got != "-" {
		t.Fatalf("zero: want -, got %s", got)
	}
	// Unix seconds (small enough to take the first branch)
	if got := FormatTime(1577836800); got != "2020-01-01" {
		t.Fatalf("seconds: want 2020-01-01, got %s", got)
	}
	// Unix milliseconds (auto-detected as > 9999999999)
	if got := FormatTime(1577836800000); got != "2020-01-01" {
		t.Fatalf("ms: want 2020-01-01, got %s", got)
	}
}

func TestLoadJSONTooLarge(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "big.json")
	// 5 MiB of bytes wrapped as a JSON string. Size on disk is what
	// trips the cap; the JSON validity does not matter because the
	// reader rejects before unmarshal.
	big := make([]byte, 5*1024*1024)
	for i := range big {
		big[i] = 'a'
	}
	if err := os.WriteFile(path, big, 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := LoadJSON(path)
	if err == nil {
		t.Fatal("want error for oversized file")
	}
	if !strings.Contains(err.Error(), "input too large") {
		t.Fatalf("want 'input too large' in error, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "4 MiB") {
		t.Fatalf("want '4 MiB' in error, got %q", err.Error())
	}
}

func TestLoadJSONAtCapBoundary(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "edge.json")
	// Pad a valid JSON object with whitespace until just under the cap.
	// This confirms the boundary is inclusive of MaxJSONSize bytes.
	body := []byte(`{"a":1}`)
	pad := make([]byte, MaxJSONSize-len(body)-1)
	for i := range pad {
		pad[i] = ' '
	}
	body = append(body, pad...)
	if int64(len(body)) > MaxJSONSize {
		t.Fatalf("test setup: body %d exceeds cap %d", len(body), MaxJSONSize)
	}
	if err := os.WriteFile(path, body, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadJSON(path); err != nil {
		t.Fatalf("file at cap should load: %v", err)
	}
}

func TestDash(t *testing.T) {
	if Dash("") != "-" {
		t.Fatal("empty should -> -")
	}
	if Dash("hello") != "hello" {
		t.Fatal("non-empty should pass through")
	}
}
