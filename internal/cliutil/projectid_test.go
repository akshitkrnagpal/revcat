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
	return root
}

// resolvedWith builds a fake Resolved bound to the given project_id, so
// tests can drive ResolveProjectID through the credential-bound branch.
func resolvedWith(projectID string) *authstore.Resolved {
	return &authstore.Resolved{
		Profile:   &authstore.Profile{Name: "test"},
		ProjectID: projectID,
		Source:    authstore.SourceLocal,
	}
}

func TestResolveProjectID_FlagWins(t *testing.T) {
	root := rootWithFlags(t)
	if err := root.PersistentFlags().Set("project-id", "from_flag"); err != nil {
		t.Fatal(err)
	}
	t.Setenv("REVCAT_PROJECT_ID", "from_env")

	got := ResolveProjectID(root, resolvedWith("from_resolved"))
	if got != "from_flag" {
		t.Fatalf("got %q, want from_flag", got)
	}
}

func TestResolveProjectID_EnvWinsOverResolved(t *testing.T) {
	root := rootWithFlags(t)
	t.Setenv("REVCAT_PROJECT_ID", "from_env")

	got := ResolveProjectID(root, resolvedWith("from_resolved"))
	if got != "from_env" {
		t.Fatalf("got %q, want from_env", got)
	}
}

func TestResolveProjectID_ResolvedWinsOverToml(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "revcat.toml"),
		[]byte(`project_id = "from_toml"`), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	t.Setenv("REVCAT_PROJECT_ID", "")

	root := rootWithFlags(t)

	got := ResolveProjectID(root, resolvedWith("from_resolved"))
	if got != "from_resolved" {
		t.Fatalf("got %q, want from_resolved", got)
	}
}

func TestResolveProjectID_FallsBackToToml(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "revcat.toml"),
		[]byte(`project_id = "from_toml"`), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	t.Setenv("REVCAT_PROJECT_ID", "")

	root := rootWithFlags(t)

	got := ResolveProjectID(root, &authstore.Resolved{Profile: &authstore.Profile{}})
	if got != "from_toml" {
		t.Fatalf("got %q, want from_toml", got)
	}
}

func TestResolveProjectID_EmptyWhenNothingConfigured(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	t.Setenv("REVCAT_PROJECT_ID", "")

	root := rootWithFlags(t)
	got := ResolveProjectID(root, &authstore.Resolved{Profile: &authstore.Profile{}})
	if got != "" {
		t.Fatalf("got %q, want empty", got)
	}
}
