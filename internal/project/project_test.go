package project

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_FindsFileInStartDir(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, FileName), `project_id = "proj_abc"`)

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.ProjectID != "proj_abc" {
		t.Fatalf("ProjectID: %q", cfg.ProjectID)
	}
	if cfg.Path != filepath.Join(dir, FileName) {
		t.Fatalf("Path: %q", cfg.Path)
	}
}

func TestLoad_WalksUpToParent(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "a", "b", "c")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, FileName), `project_id = "proj_root"`)

	cfg, err := Load(nested)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.ProjectID != "proj_root" {
		t.Fatalf("ProjectID: %q", cfg.ProjectID)
	}
}

func TestLoad_NotFoundWhenNoFileAnywhere(t *testing.T) {
	dir := t.TempDir()
	_, err := Load(dir)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("got %v, want ErrNotFound", err)
	}
}

func TestLoad_ParsesAppsArray(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, FileName), `
project_id = "proj_x"

[[apps]]
id   = "app_ios"
name = "iOS"

[[apps]]
id   = "app_android"
name = "Android"
`)

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got := len(cfg.Apps); got != 2 {
		t.Fatalf("apps: %d", got)
	}
	if cfg.Apps[0].ID != "app_ios" || cfg.Apps[1].ID != "app_android" {
		t.Fatalf("apps: %+v", cfg.Apps)
	}
}

func TestLoad_MalformedReturnsError(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, FileName), `project_id =`)

	_, err := Load(dir)
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestSaveAndReload_Roundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, FileName)
	in := &Config{
		ProjectID: "proj_save",
		Apps: []App{
			{ID: "app_a", Name: "A"},
			{ID: "app_b"},
		},
	}
	if err := Save(path, in); err != nil {
		t.Fatalf("Save: %v", err)
	}
	out, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if out.ProjectID != "proj_save" || len(out.Apps) != 2 {
		t.Fatalf("roundtrip: %+v", out)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
