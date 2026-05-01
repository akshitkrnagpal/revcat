package cliutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"

	authstore "github.com/akshitkrnagpal/revcat/internal/auth"
)

// rootWithFlags constructs a cobra root with the global --project-id flag
// wired up, mirroring commands/root.go. Used to drive ProjectIDFlag /
// ResolveProjectID under test without importing the commands package
// (which would create a cycle).
func rootWithFlags(t *testing.T) *cobra.Command {
	t.Helper()
	root := &cobra.Command{Use: "revcat"}
	pf := root.PersistentFlags()
	pf.String("project-id", "", "")
	pf.String("profile", "", "")
	pf.Bool("bypass-keychain", false, "")
	return root
}

func TestResolveProjectID_FlagWins(t *testing.T) {
	root := rootWithFlags(t)
	if err := root.PersistentFlags().Set("project-id", "from_flag"); err != nil {
		t.Fatal(err)
	}
	t.Setenv("REVCAT_PROJECT_ID", "from_env")
	prof := &authstore.Profile{ProjectID: "from_profile"}

	got := ResolveProjectID(root, prof)
	if got != "from_flag" {
		t.Fatalf("got %q, want from_flag", got)
	}
}

func TestResolveProjectID_EnvWinsOverProfile(t *testing.T) {
	root := rootWithFlags(t)
	t.Setenv("REVCAT_PROJECT_ID", "from_env")
	prof := &authstore.Profile{ProjectID: "from_profile"}

	got := ResolveProjectID(root, prof)
	if got != "from_env" {
		t.Fatalf("got %q, want from_env", got)
	}
}

func TestResolveProjectID_TomlWinsOverProfile(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "revcat.toml"),
		[]byte(`project_id = "from_toml"`), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	t.Setenv("REVCAT_PROJECT_ID", "")

	root := rootWithFlags(t)
	prof := &authstore.Profile{ProjectID: "from_profile"}

	got := ResolveProjectID(root, prof)
	if got != "from_toml" {
		t.Fatalf("got %q, want from_toml", got)
	}
}

func TestResolveProjectID_FallsBackToProfile(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	t.Setenv("REVCAT_PROJECT_ID", "")

	root := rootWithFlags(t)
	prof := &authstore.Profile{ProjectID: "from_profile"}

	got := ResolveProjectID(root, prof)
	if got != "from_profile" {
		t.Fatalf("got %q, want from_profile", got)
	}
}

func TestResolveProjectID_EmptyWhenNothingConfigured(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	t.Setenv("REVCAT_PROJECT_ID", "")

	root := rootWithFlags(t)
	got := ResolveProjectID(root, &authstore.Profile{})
	if got != "" {
		t.Fatalf("got %q, want empty", got)
	}
}
