package auth

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolve_EnvHatchWinsOverEverything(t *testing.T) {
	t.Setenv("REVCAT_REFRESH_TOKEN", "rtk_from_env")
	t.Setenv("REVCAT_PROJECT_ID", "proj_env")

	got, err := Resolve(ResolveOptions{Cwd: t.TempDir()})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if got.Source != SourceEnv {
		t.Fatalf("Source: got %q want %q", got.Source, SourceEnv)
	}
	if got.Profile.RefreshToken != "rtk_from_env" {
		t.Fatalf("RefreshToken: %q", got.Profile.RefreshToken)
	}
	if got.ProjectID != "proj_env" {
		t.Fatalf("ProjectID: %q", got.ProjectID)
	}
}

func TestResolve_LocalConfigWinsOverGlobal(t *testing.T) {
	t.Setenv("REVCAT_REFRESH_TOKEN", "")
	dir := t.TempDir()
	path := filepath.Join(dir, ".revcat", "config.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path,
		[]byte(`{"project_id":"proj_local","profile":{"name":"d","access_token":"atk_local","refresh_token":"rtk_local"}}`),
		0o600); err != nil {
		t.Fatal(err)
	}

	got, err := Resolve(ResolveOptions{Cwd: dir})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if got.Source != SourceLocal {
		t.Fatalf("Source: got %q want %q", got.Source, SourceLocal)
	}
	if got.ProjectID != "proj_local" {
		t.Fatalf("ProjectID: %q", got.ProjectID)
	}
	if got.Profile.AccessToken != "atk_local" {
		t.Fatalf("AccessToken: %q", got.Profile.AccessToken)
	}
}

func TestDecodeProfile_RejectsLegacySecretKey(t *testing.T) {
	legacy := []byte(`{"name":"default","secret_key":"sk_live_xyz","project_id":"proj_x"}`)
	_, err := decodeProfile(legacy, "default")
	if err != ErrLegacyProfile {
		t.Fatalf("got %v, want ErrLegacyProfile", err)
	}
}

func TestDecodeProfile_AcceptsOAuthShape(t *testing.T) {
	good := []byte(`{"name":"default","access_token":"atk","refresh_token":"rtk"}`)
	p, err := decodeProfile(good, "default")
	if err != nil {
		t.Fatalf("decodeProfile: %v", err)
	}
	if p.AccessToken != "atk" || p.RefreshToken != "rtk" {
		t.Fatalf("got %+v", p)
	}
}

func TestResolve_SelfHealsStaleActiveMarker(t *testing.T) {
	t.Setenv("REVCAT_REFRESH_TOKEN", "")
	t.Setenv("REVCAT_PROFILE", "")
	t.Setenv("HOME", t.TempDir())

	// Set up a global file store with only "default", then point the
	// active marker at "ghost" (a profile that no longer exists).
	store, err := openGlobalFile()
	if err != nil {
		t.Fatalf("openGlobalFile: %v", err)
	}
	if err := store.Set(&Profile{Name: "default", AccessToken: "atk", RefreshToken: "rtk"}); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if err := SetActive("ghost"); err != nil {
		t.Fatalf("SetActive: %v", err)
	}

	got, err := Resolve(ResolveOptions{Cwd: t.TempDir()})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if got.Profile.Name != "default" {
		t.Fatalf("expected fallback to default, got %q", got.Profile.Name)
	}

	// And the stale marker should be cleared so subsequent runs are
	// fast paths.
	if active, _ := GetActive(); active != "" {
		t.Fatalf("expected active marker cleared, got %q", active)
	}
}
