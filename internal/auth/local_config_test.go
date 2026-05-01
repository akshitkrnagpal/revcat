package auth

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadLocal_FindsFileInStartDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".revcat", "config.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path,
		[]byte(`{"project_id":"proj_x","profile":{"name":"d","access_token":"atk"}}`),
		0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadLocal(dir)
	if err != nil {
		t.Fatalf("LoadLocal: %v", err)
	}
	if cfg.ProjectID != "proj_x" || cfg.Profile.AccessToken != "atk" {
		t.Fatalf("got %+v", cfg)
	}
	if cfg.Path != path {
		t.Fatalf("Path: got %q want %q", cfg.Path, path)
	}
}

func TestLoadLocal_WalksUp(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	cfgDir := filepath.Join(root, ".revcat")
	if err := os.MkdirAll(cfgDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "config.json"),
		[]byte(`{"project_id":"proj_root","profile":{"name":"d","access_token":"atk"}}`),
		0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadLocal(nested)
	if err != nil {
		t.Fatalf("LoadLocal: %v", err)
	}
	if cfg.ProjectID != "proj_root" {
		t.Fatalf("ProjectID: %q", cfg.ProjectID)
	}
}

func TestLoadLocal_ErrNoLocalConfigWhenAbsent(t *testing.T) {
	dir := t.TempDir()
	_, err := LoadLocal(dir)
	if !errors.Is(err, ErrNoLocalConfig) {
		t.Fatalf("got %v, want ErrNoLocalConfig", err)
	}
}

func TestLoadLocal_RejectsEmptyProfile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".revcat", "config.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path,
		[]byte(`{"project_id":"proj_x","profile":{}}`),
		0o600); err != nil {
		t.Fatal(err)
	}

	_, err := LoadLocal(dir)
	if err == nil || !strings.Contains(err.Error(), "rerun `revcat init`") {
		t.Fatalf("expected helpful error, got %v", err)
	}
}

func TestSaveLocal_Mode0600AndAtomic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".revcat", "config.json")
	in := &LocalConfig{
		ProjectID: "proj_save",
		Profile:   Profile{Name: "d", AccessToken: "atk", RefreshToken: "rtk"},
		Apps:      []LocalApp{{ID: "app_a"}},
	}
	if err := SaveLocal(path, in); err != nil {
		t.Fatalf("SaveLocal: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	// Skip mode check on platforms that don't enforce 0600 strictly.
	if got := info.Mode().Perm(); got != 0o600 {
		t.Errorf("mode: got %o want 0600", got)
	}

	out, err := LoadLocal(dir)
	if err != nil {
		t.Fatalf("LoadLocal: %v", err)
	}
	if out.ProjectID != in.ProjectID || out.Profile.AccessToken != in.Profile.AccessToken {
		t.Fatalf("roundtrip: %+v", out)
	}
}

func TestEnsureGitignored_AppendsWhenAbsent(t *testing.T) {
	dir := t.TempDir()
	added, err := EnsureGitignored(dir)
	if err != nil {
		t.Fatalf("EnsureGitignored: %v", err)
	}
	if !added {
		t.Fatal("expected added=true on fresh dir")
	}
	body, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), ".revcat/") {
		t.Fatalf("gitignore missing entry: %q", body)
	}
}

func TestEnsureGitignored_IdempotentWhenPresent(t *testing.T) {
	dir := t.TempDir()
	gi := filepath.Join(dir, ".gitignore")
	if err := os.WriteFile(gi, []byte("node_modules/\n.revcat/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	added, err := EnsureGitignored(dir)
	if err != nil {
		t.Fatalf("EnsureGitignored: %v", err)
	}
	if added {
		t.Fatal("expected added=false when entry already present")
	}
	body, _ := os.ReadFile(gi)
	if strings.Count(string(body), ".revcat/") != 1 {
		t.Fatalf("entry duplicated: %q", body)
	}
}

func TestEnsureGitignored_AppendsNewlineIfMissing(t *testing.T) {
	dir := t.TempDir()
	gi := filepath.Join(dir, ".gitignore")
	// File without trailing newline.
	if err := os.WriteFile(gi, []byte("node_modules/"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := EnsureGitignored(dir); err != nil {
		t.Fatalf("EnsureGitignored: %v", err)
	}
	body, _ := os.ReadFile(gi)
	if !strings.HasSuffix(string(body), ".revcat/\n") {
		t.Fatalf("missing trailing entry/newline: %q", body)
	}
}
