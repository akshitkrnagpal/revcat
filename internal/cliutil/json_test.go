package cliutil

import (
	"os"
	"path/filepath"
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

func TestDash(t *testing.T) {
	if Dash("") != "-" {
		t.Fatal("empty should -> -")
	}
	if Dash("hello") != "hello" {
		t.Fatal("non-empty should pass through")
	}
}
